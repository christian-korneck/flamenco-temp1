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
    <dl class="metadata">
    </dl>
  </div>
</template>

<script lang="js">
import * as datetime from "../datetime";

import {
  JobsApi,
} from '../manager-api'

export default {
  props: [
    "apiClient",  // Flamenco Manager API client.

    // Object, subset of job info, should at least contain an 'id' key. This ID
    // determines the job that's shown here. The rest of the fields are used to
    // initialise the details until the full job has been fetched from the API.
    "jobSummary",
  ],
  data: () => {
    return {
      jobData: {},
      datetime: datetime,
    };
  },
  mounted() {
    // Allow testing from the JS console:
    window.jobDetailsVue = this;
  },
  watch: {
    jobSummary(newSummary, oldSummary) {
      console.log("Updating job details:", JSON.parse(JSON.stringify(newSummary)));
      this.jobData = newSummary;
      // TODO: Fetch the rest of the job.
    },
  },
  methods: {
    onReconnected() {
      // If the connection to the backend was lost, we have likely missed some
      // updates. Just fetch the data and start from scratch.
      this.fetchJob();
    },
    fetchJob() {
      if (!this.apiClient) {
        throw "no apiClient set on JobDetails component";
      }
      if (!this.jobSummary || !this.jobSummary.id) {
        // no job selected, which is fine.
        this.clearJobDetails();
        return "";
      }
      const jobsApi = new JobsApi(this.apiClient);
      const jobID = this.jobSummary.id;
      jobsApi.fetchJob(jobID).then(this.onJobFetched, function (error) {
        // TODO: error handling.
        console.error(error);
      });
      return jobID;
    },
    onJobFetched(data) {
      console.log("Job fetched:", data);
    },
    clearJobDetails() {
      this.jobData = {};
    },
  }
};
</script>

<style scoped>
.job-details {
  font-size: smaller;
  font-family: 'Noto Mono', monospace;
}

tr:hover {
  background-color: lightgrey;
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
