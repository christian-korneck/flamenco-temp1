<template>
  <h2 class="column-title">Job Details</h2>
  <div class="job-details">
    <table class="details">
      <tr class="field-id">
        <th>ID</th>
        <td>{{ jobData.id }}</td>
      </tr>

      <tr class="field-name">
        <th>Name</th>
        <td>{{ jobData.name }}</td>
      </tr>

      <tr class="field-status">
        <th>Status</th>
        <td>{{ jobData.status }}</td>
      </tr>

      <tr class="field-type">
        <th>Type</th>
        <td>{{ jobData.type }}</td>
      </tr>

      <tr class="field-priority">
        <th>Prio</th>
        <td>{{ jobData.priority }}</td>
      </tr>

      <tr class="field-created">
        <th>Created</th>
        <td>{{ datetime.relativeTime(jobData.created) }}</td>
      </tr>

      <tr class="field-updated">
        <th>Updated</th>
        <td>{{ datetime.relativeTime(jobData.updated) }}</td>
      </tr>
    </table>

    <h3 class="sub-title">Meta-data</h3>
    <table class="metadata">
      <tr v-for="value, key in jobData.metadata" :class="`field-${key}`">
        <th>{{ key }}</th>
        <td>{{ value }}</td>
      </tr>
    </table>

    <h3 class="sub-title">Settings</h3>
    <table class="settings">
      <tr v-for="value, key in jobData.settings" :class="`field-${key}`">
        <th>{{ key }}</th>
        <td>{{ value }}</td>
      </tr>
    </table>

  </div>
</template>

<script lang="js">
import * as datetime from "../datetime";

export default {
  props: [
    "apiClient",  // Flamenco Manager API client.
    "jobData", // Job data to show.
  ],
  data: () => {
    return {
      datetime: datetime, // So that the template can access it.
    };
  },
  mounted() {
    // Allow testing from the JS console:
    window.jobDetailsVue = this;
  },
};
</script>

<style scoped>
.job-details {
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
