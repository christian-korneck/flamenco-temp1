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

      <dt class="field-type">Type</dt>
      <dd>{{ taskData.type }}</dd>

      <dt class="field-priority">Priority</dt>
      <dd>{{ taskData.priority }}</dd>

      <dt class="field-created">Created</dt>
      <dd>{{ datetime.relativeTime(taskData.created) }}</dd>

      <dt class="field-updated">Updated</dt>
      <dd>{{ datetime.relativeTime(taskData.updated) }}</dd>
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

function objectEmpty(o) {
  if (!o) return true;
  return Object.entries(o).length == 0;
}

export default {
  props: [
    "taskData", // Task data to show.
  ],
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
