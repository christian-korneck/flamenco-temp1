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
  </template>

  <div v-else class="details-no-item-selected">
    <p>Select a worker to see its details.</p>
  </div>
</template>

<script lang="js">
import * as datetime from "@/datetime";
import { WorkerMgtApi } from '@/manager-api';
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
    };
  },
  mounted() {
    // Allow testing from the JS console:
    window.workerDetailsVue = this;
  },
  watch: {
    workerData(newData) {
      console.log("new data:", plain(newData));
      if (newData)
        this.workerStatusHTML = workerStatus(newData);
      else
        this.workerStatusHTML = "";
    },
  },
  computed: {
    hasWorkerData() {
      return !!this.workerData && !!this.workerData.id;
    },
  },
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
