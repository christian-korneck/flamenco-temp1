<template>
  <h2 class="column-title">Worker Details</h2>

  <template v-if="hasWorkerData">
    <dl>
      <dt class="field-id">ID</dt>
      <dd>{{ workerData.id }}</dd>

      <dt class="field-nickname">Nickname</dt>
      <dd>{{ workerData.nickname }}</dd>

      <dt class="field-status">Status</dt>
      <dd>{{ workerData.status }}</dd>
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

export default {
  props: [
    "workerData", // Worker data to show.
  ],
  data() {
    return {
      datetime: datetime, // So that the template can access it.
      api: new WorkerMgtApi(apiClient),
    };
  },
  mounted() {
    // Allow testing from the JS console:
    window.workerDetailsVue = this;
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
.field-id + dd {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
