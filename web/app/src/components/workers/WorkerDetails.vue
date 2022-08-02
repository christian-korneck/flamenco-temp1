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

    <section class="sleep-schedule" :class="{'is-schedule-active': workerSleepSchedule.is_active}">
      <h3 class="sub-title">
        <switch-checkbox :isChecked="workerSleepSchedule.is_active" @switch-toggle="toggleWorkerSleepSchedule">
        </switch-checkbox>
        Sleep Schedule
        <div v-if="!isScheduleEditing" class="sub-title-buttons">
          <button @click="isScheduleEditing = true">Edit</button>
        </div>
      </h3>
      <p>Time of the day (and on which days) this worker should go to sleep. </p>

      <div class="sleep-schedule-edit" v-if="isScheduleEditing">
        <div>
          <label>Days of the week</label>
          <input type="text" placeholder="mo tu we th fr" v-model="workerSleepSchedule.days_of_week">
          <span class="input-help-text">
            Write each day name using their first two letters, separated by spaces.
            (e.g. mo tu we th fr)
          </span>
        </div>
        <div class="sleep-schedule-edit-time">
          <div>
            <label>Start Time</label>
            <input type="text" placeholder="09:00" v-model="workerSleepSchedule.start_time" class="time">
          </div>
          <div>
            <label>End Time</label>
            <input type="text" placeholder="18:00" v-model="workerSleepSchedule.end_time" class="time">
          </div>
        </div>
        <span class="input-help-text">
          Use 24-hour format.
        </span>
        <div class="btn-bar-group">
          <div class="btn-bar">
            <button v-if="isScheduleEditing" @click="cancelEditWorkerSleepSchedule" class="btn">Cancel</button>
            <button v-if="isScheduleEditing" @click="saveWorkerSleepSchedule" class="btn btn-primary">Save
              Schedule</button>
          </div>
        </div>
      </div>

      <dl v-if="!isScheduleEditing">
        <dt>Status</dt>
        <dd>
          {{ workerSleepScheduleStatusLabel }}
        </dd>
        <dt>Days of Week</dt>
        <dd>
          {{ workerSleepScheduleFormatted.days_of_week }}
        </dd>
        <dt>Start Time</dt>
        <dd>
          {{ workerSleepScheduleFormatted.start_time }}
        </dd>
        <dt>End Time</dt>
        <dd>
          {{ workerSleepScheduleFormatted.end_time }}
        </dd>
      </dl>
    </section>
  </template>

  <div v-else class="details-no-item-selected">
    <p>Select a worker to see its details.</p>
  </div>
</template>

<script>
import { useNotifs } from '@/stores/notifications'

import * as datetime from "@/datetime";
import { WorkerMgtApi, WorkerSleepSchedule } from '@/manager-api';
import { apiClient } from '@/stores/api-query-count';
import { workerStatus } from "../../statusindicator";
import LinkWorkerTask from '@/components/LinkWorkerTask.vue';
import SwitchCheckbox from '@/components/SwitchCheckbox.vue';

export default {
  props: [
    "workerData", // Worker data to show.
  ],
  components: {
    LinkWorkerTask,
    SwitchCheckbox,
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
    workerSleepScheduleStatusLabel() {
      return this.workerSleepSchedule.is_active ? 'Enabled' : 'Disabled';
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
      this.workerSleepSchedule.is_active = !this.workerSleepSchedule.is_active;
      let verb = this.workerSleepScheduleStatusLabel;
      this.setWorkerSleepSchedule(`${verb} schedule for worker ${this.workerData.name}`);
    },
    saveWorkerSleepSchedule() {
      this.setWorkerSleepSchedule(`Updated schedule for worker ${this.workerData.name}`);
      this.isScheduleEditing = false;
    },
    cancelEditWorkerSleepSchedule() {
      this.fetchWorkerSleepSchedule();
      this.isScheduleEditing = false;
    },
    defaultWorkerSleepSchedule() {
      return new WorkerSleepSchedule(false, '', '', '')  // Default values in OpenAPI
    },
  }
};
</script>

<style scoped>
.sub-title {
  position: relative;
}

.switch {
  top: 1px;
}

.sub-title-buttons {
  position: absolute;
  right: var(--spacer);
  bottom: var(--spacer-xs);
}

.sleep-schedule .btn-bar label+.btn {
  margin-left: var(--spacer-sm);
}

.sleep-schedule dl {
  color: var(--color-text-muted);
}

.sleep-schedule.is-schedule-active dl {
  color: unset;
}

.sleep-schedule-edit {
  display: flex;
  margin: var(--spacer-lg) 0 0;
  flex-direction: column;
}

.sleep-schedule-edit>div {
  margin: var(--spacer-sm) 0;
}

.sleep-schedule-edit label {
  margin-bottom: var(--spacer-xs);
  font-size: var(--font-size-sm);
  font-weight: bold;
  color: var(--color-text-muted);
}

.sleep-schedule-edit>.sleep-schedule-edit-time {
  display: flex;
  margin-bottom: 0;
}

.toggle-sleep-schedule {
  visibility: hidden;
  display: none;
}

.sleep-schedule-edit input[type="text"] {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
  width: 23ch;
}

.sleep-schedule-edit input[type="text"].time {
  width: 10ch;
  margin-right: var(--spacer);
}

/* Prevent fields with long IDs from overflowing. */
.field-id+dd {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
