<template>
  <div class="job-details">
    <h2 class="column-title">Job Details</h2>
  </div>
</template>

<script lang="js">
import { DateTime } from "luxon";

import {
  JobsApi,
} from '../manager-api'

export default {
  props: ["apiClient"],
  data: () => {
    return {
    };
  },
  mounted() {
    // Allow testing from the JS console:
    window.jobDetailsVue = this;
  },
  methods: {
    onReconnected() {
      // If the connection to the backend was lost, we have likely missed some
      // updates. Just fetch the data and start from scratch.
      this.fetchJob();
    },
    fetchAllJob() {
      if (this.apiClient === undefined) {
        throw "no apiClient set on JobsTable component";
      }
      const jobsApi = new JobsApi(this.apiClient);
      const jobID = ""; // TODO: get from outer scope.
      jobsApi.fetchJob(jobID).then(this.onJobFetched, function (error) {
        // TODO: error handling.
        console.error(error);
      });
    },
    onJobFetched(data) {
      console.log("Job fetched:", data);
    },
  }
};
</script>

<style scoped>
</style>
