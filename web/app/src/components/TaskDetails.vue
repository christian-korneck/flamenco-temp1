<template>
  <h2 class="column-title">Task Details</h2>
  <div v-if="hasTaskData" class="task-details">
    <table class="details">
      <tr class="field-id">
        <th>ID</th>
        <td>{{ taskData.id }}</td>
      </tr>

      <tr class="field-name">
        <th>Name</th>
        <td>{{ taskData.name }}</td>
      </tr>

      <tr class="field-status">
        <th>Status</th>
        <td>{{ taskData.status }}</td>
      </tr>

      <tr class="field-type">
        <th>Type</th>
        <td>{{ taskData.type }}</td>
      </tr>

      <tr class="field-priority">
        <th>Prio</th>
        <td>{{ taskData.priority }}</td>
      </tr>

      <tr class="field-created">
        <th>Created</th>
        <td>{{ datetime.relativeTime(taskData.created) }}</td>
      </tr>
      <tr class="field-updated">
        <th>Updated</th>
        <td>{{ datetime.relativeTime(taskData.updated) }}</td>
      </tr>
    </table>

    <h3 class="sub-title">Commands</h3>
    <table class="commands">
      <tr v-for="cmd in taskData.commands" :class="`field-${cmd.name}`">
        <th>{{ cmd.name }}</th>
        <td>{{ cmd.parameters }}</td>
      </tr>
    </table>

  </div>
  <div v-else class="no-task-selected">
    <p>No task selected, pick one from the list on the left.</p>
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
.task-details {
  font-size: smaller;
  font-family: 'Noto Mono', monospace;
}

tr:hover {
  background-color: #333333;
}

tr.field-id td {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

th {
  font-weight: bold;
  text-align: right;
  vertical-align: top;
}
</style>
