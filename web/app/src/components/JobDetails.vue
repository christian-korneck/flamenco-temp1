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

    <h3 class="sub-title" v-if="hasMetadata">Meta-data</h3>
    <table class="metadata">
      <tr v-for="value, key in jobData.metadata" :class="`field-${key}`">
        <th>{{ key }}</th>
        <td>{{ value }}</td>
      </tr>
    </table>

    <h3 class="sub-title" v-if="hasSettings">Settings</h3>
    <table class="settings">
      <tr v-for="value, key in settings" :class="`field-${key}`">
        <th>{{ key }}</th>
        <td>{{ value }}</td>
      </tr>
    </table>

  </div>
</template>

<script lang="js">
import * as datetime from "../datetime";
import * as API from "../manager-api";

function objectEmpty(o) {
  if (!o) return true;
  return Object.entries(o).length == 0;
}

export default {
  props: [
    "apiClient",  // Flamenco Manager API client.
    "jobData", // Job data to show.
  ],
  data() {
    return {
      datetime: datetime, // So that the template can access it.
      settings: null, // Object with filtered job settings, or null if there is no job.
      jobsApi: new API.JobsApi(this.apiClient),
      jobType: null, // API.AvailableJobType object for the current job type.
    };
  },
  mounted() {
    // Allow testing from the JS console:
    window.jobDetailsVue = this;
  },
  computed: {
    hasMetadata() {
      return this.jobData && !objectEmpty(this.jobData.metadata);
    },
    hasSettings() {
      return this.jobData && !objectEmpty(this.settings);
    },
  },
  watch: {
    jobData(newJobData) {
      if (objectEmpty(newJobData)) {
        this.settings = null;
        return;
      }

      this.settings = newJobData.settings;

      // Only fetch the job type if it's different from what's already loaded.
      if (objectEmpty(this.jobType) || this.jobType.name != newJobData.type) {
        this.jobsApi.getJobType(newJobData.type)
          .then(this.onJobTypeLoaded)
          .catch((error) => { console.warn(error) });
      }
    }
  },
  methods: {
    onJobTypeLoaded(jobType) {
      console.log("Job type loaded: ", jobType);
      this.jobType = jobType;
    }
  }
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
