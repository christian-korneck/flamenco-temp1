<template>
  <h2 class="column-title">Task Details</h2>

  <template v-if="hasTaskData">
    <dl>
      <dt class="field-id">ID</dt>
      <dd>{{ taskData.id }}</dd>

      <dt class="field-name">Name</dt>
      <dd>{{ taskData.name }}</dd>

      <dt class="field-status">Status</dt>
      <dd>{{ taskData.status }}</dd>

      <dt class="field-worker">Assigned To</dt>
      <dd><worker-link :worker="taskData.worker" /></dd>

      <dt class="field-activity">Activity</dt>
      <dd>{{ taskData.activity }}</dd>

      <dt class="field-type">Type</dt>
      <dd>{{ taskData.type }}</dd>

      <dt class="field-priority">Priority</dt>
      <dd>{{ taskData.priority }}</dd>

      <dt class="field-created">Created</dt>
      <dd>{{ datetime.relativeTime(taskData.created) }}</dd>

      <dt class="field-updated">Updated</dt>
      <dd>{{ datetime.relativeTime(taskData.updated) }}</dd>

      <dt class="field-last-touched">Last Touched by Worker</dt>
      <dd>{{ datetime.relativeTime(taskData.last_touched) }}</dd>

      <template v-if="taskData.failed_by_workers.length > 0">
      <dt class="field-failed-by-workers">Failed by Workers</dt>
      <dd v-for="worker in taskData.failed_by_workers">
        <worker-link :worker="worker" />
      </dd>
      </template>
    </dl>

    <h3 class="sub-title">Commands</h3>
    <dl>
      <template v-for="cmd in taskData.commands">
        <dt :class="`field-${cmd.name}`" :title="cmd.name">{{ cmd.name }}</dt>
        <dd>{{ cmd.parameters }}</dd>
      </template>
    </dl>
  </template>

  <div v-else class="details-no-item-selected">
    <p>Select a task to see its details.</p>
  </div>
</template>

<script lang="js">
import * as datetime from "@/datetime";
import * as API from '@/manager-api';
import { apiClient } from '@/stores/api-query-count';
import WorkerLink from '@/components/WorkerLink.vue';

export default {
  props: [
    "taskData", // Task data to show.
  ],
  components: {WorkerLink},
  data() {
    return {
      datetime: datetime, // So that the template can access it.
      jobsApi: new API.JobsApi(apiClient),
    };
  },
  mounted() {
    // Allow testing from the JS console:
    window.taskDetailsVue = this;
  },
  computed: {
    hasTaskData() {
      return !!this.taskData && !!this.taskData.id;
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
