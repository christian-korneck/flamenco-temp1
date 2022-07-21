<template>
  <h2 class="column-title">Worker Details</h2>

  <template v-if="hasWorkerData">
    <dl>
      <dt class="field-id">ID</dt>
      <dd>{{ workerData.id }}</dd>

      <dt class="field-name">Name</dt>
      <dd>{{ workerData.name }}</dd>

      <dt class="field-status">Status</dt>
      <dd v-html="workerStatusHTML"></dd>

      <dt class="field-last_seen">Last Seen</dt>
      <dd v-if="workerData.last_seen">{{ datetime.relativeTime(workerData.last_seen) }}</dd>
      <dd v-else>never</dd>

      <dt class="field-version">Version</dt>
      <dd title="Version of Flamenco">{{ workerData.version }}</dd>

      <dt class="field-ip_address">IP Addr</dt>
      <dd>{{ workerData.ip_address }}</dd>

      <dt class="field-platform">Platform</dt>
      <dd>{{ workerData.platform }}</dd>

      <dt class="field-supported_task_types">Task Types</dt>
      <dd>{{ workerData.supported_task_types.join(', ') }}</dd>

      <dt class="field-task">Last Task</dt>
      <dd>
        <link-worker-task :workerTask="workerData.task" />
      </dd>
    </dl>

    <h3 class="sub-title">Sleep Schedule</h3>
    <button v-if="isScheduleEditing" @click="saveWorkerSleepSchedule" class="btn">Save</button>
    <button v-else @click="isScheduleEditing = true" class="btn">Edit</button>
    <dl>
      <dt class="field-is_active">Use Schedule</dt>
      <dd v-if="isScheduleEditing"><input type="checkbox" v-model="workerSleepSchedule.is_active"></dd>
      <dd v-else><input @change="toggleWorkerSleepSchedule" type="checkbox" v-model="workerSleepSchedule.is_active"></dd>

      <dt class="field-days_of_week">Days of Week</dt>
      <dd v-if="isScheduleEditing"><input type="text" placeholder="mo tu we th fr" v-model="workerSleepSchedule.days_of_week"></dd>
      <dd v-else>{{ workerSleepScheduleFormatted.days_of_week }}</dd>

      <dt class="field-start_time">Start Time</dt>
      <dd v-if="isScheduleEditing"><input type="text" placeholder="09:00" v-model="workerSleepSchedule.start_time"></dd>
      <dd v-else>{{ workerSleepScheduleFormatted.start_time }}</dd>

      <dt class="field-end_time">End Time</dt>
      <dd v-if="isScheduleEditing"><input type="text" placeholder="18:00" v-model="workerSleepSchedule.end_time"></dd>
      <dd v-else>{{ workerSleepScheduleFormatted.end_time }}</dd>
    </dl>
  </template>

  <div v-else class="details-no-item-selected">
    <p>Select a worker to see its details.</p>
  </div>
</template>

<script lang="js">
import { useNotifs } from '@/stores/notifications'

import * as datetime from "@/datetime";
import { WorkerMgtApi, WorkerSleepSchedule } from '@/manager-api';
import { apiClient } from '@/stores/api-query-count';
import { workerStatus } from "../../statusindicator";
import LinkWorkerTask from '@/components/LinkWorkerTask.vue';

export default {
  props: [
    "workerData", // Worker data to show.
  ],
  components: {
    LinkWorkerTask,
  },
  data() {
    return {
      datetime: datetime, // So that the template can access it.
      api: new WorkerMgtApi(apiClient),
      workerStatusHTML: "",
      workerSleepSchedule: this.defaultWorkerSleepSchedule(),
      isScheduleEditing: false,
      notifs: useNotifs(),
    };
  },
  mounted() {
    // Allow testing from the JS console:
    window.workerDetailsVue = this;
  },
  watch: {
    workerData(newData, oldData) {
      if (newData) {
        this.workerStatusHTML = workerStatus(newData);
      } else {
        this.workerStatusHTML = "";
      }
      // Update workerSleepSchedule only if oldData and newData have different ids, or if there is no oldData
      // and we provide newData.
      if (((oldData && newData) && (oldData.id != newData.id)) || !oldData && newData) {
        this.fetchWorkerSleepSchedule();
      }
    },
  },
  computed: {
    hasWorkerData() {
      return !!this.workerData && !!this.workerData.id;
    },
    workerSleepScheduleFormatted() {
      // Utility to display workerSleepSchedule, taking into account the case when the default values are used.
      // This way, empty strings are represented more meaningfully.
      return {
        'days_of_week': this.workerSleepSchedule.days_of_week === '' ? 'every day' : this.workerSleepSchedule.days_of_week,
        'start_time': this.workerSleepSchedule.start_time === '' ? '00:00' : this.workerSleepSchedule.start_time,
        'end_time': this.workerSleepSchedule.end_time === '' ? '24:00' : this.workerSleepSchedule.end_time,
      }
    },
  },
  methods: {
    fetchWorkerSleepSchedule() {
      this.api.fetchWorkerSleepSchedule(this.workerData.id)
        .then((schedule) => {
          // Replace the default workerSleepSchedule if the Worker has one

          if (schedule) {
            this.workerSleepSchedule = schedule;
          } else {
            this.workerSleepSchedule = this.defaultWorkerSleepSchedule();
          }
        })
        .catch((error) => {
          const errorMsg = JSON.stringify(error); // TODO: handle API errors better.
          this.notifs.add(`Error: ${errorMsg}`);
        });
    },
    setWorkerSleepSchedule(notifMessage) {
      this.api.setWorkerSleepSchedule(this.workerData.id, this.workerSleepSchedule).then(
        this.notifs.add(notifMessage));
    },
    toggleWorkerSleepSchedule() {
      let verb = this.workerSleepSchedule.is_active ? 'Enabled' : 'Disabled';
      this.setWorkerSleepSchedule(`${verb} schedule for worker ${this.workerData.name}`);
    },
    saveWorkerSleepSchedule() {
      this.setWorkerSleepSchedule(`Updated schedule for worker ${this.workerData.name}`);
      this.isScheduleEditing = false;
    },
    defaultWorkerSleepSchedule() {
      return new WorkerSleepSchedule(false, '', '', '')  // Default values in OpenAPI
    },
  }
};
</script>

<style scoped>
/* Prevent fields with long IDs from overflowing. */
.field-id+dd {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
