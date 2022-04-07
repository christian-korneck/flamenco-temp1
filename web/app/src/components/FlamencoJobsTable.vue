<template>
  <div class="flamenco-jobs-table">
    <JobsTabulator v-model="jobs" :options="options" />
  </div>
</template>

<script>
import { TabulatorComponent } from "vue-tabulator";

const flamencoManager = require('flamenco-manager');

export default {
  props: ["apiClient"],
  components: {
    JobsTabulator: TabulatorComponent,
  },
  data: () => {
    const options = {
      // See pkg/api/flamenco-manager.yaml, schemas Job and JobUpdate.
      columns: [
        { title: 'Name', field: 'name', sorter: 'string', width: 200 },
        { title: 'Status', field: 'status', sorter: 'string', width: 100 },
        { title: 'Type', field: 'type', sorter: 'string', width: 150 },
        { title: 'Prio', field: 'priority', sorter: 'int', width: 60 },
        { title: 'Updated', field: 'updated', sorter: 'date', width: 300 },
      ]
    };

    return {
      jobs: [
        // Will be loaded via a Flamenco API call.
      ],
      options: options,
    };
  },
  mounted: function() {
    this.fetchAllJobs();
  },
  methods: {
    fetchAllJobs: function() {
      // TODO: maybe move fetching of the data out of this component?
      const JobsApi = new flamencoManager.JobsApi(this.apiClient);
      const jobsQuery = {};

      JobsApi.queryJobs(jobsQuery).then(this.onJobsFetched, function(error) {
        // TODO: error handling.
        console.error(error);
      });
    },
    onJobsFetched: function(data) {
      this.jobs = data.jobs;
    },
    processJobUpdate: function(jobUpdate) {
      console.log("processJobUpdate:", jobUpdate);
      for (let job of this.jobs) {
        if (job.id != jobUpdate.id) {
          console.log("not same ID:", job.id, jobUpdate.id);
          continue;
        }
        console.log("Found job:", JSON.parse(JSON.stringify(job)));
        job.name = jobUpdate.name; // Shouldn't be changing, but you never know.
        job.status = jobUpdate.status;
        job.updated = jobUpdate.updated;
      }
    },
    processNewJob: function(jobUpdate) {
      console.log("processNewJob:", jobUpdate);
      // The update doesn't have all the info we need, so just fetch the job via an API call.
      // Updates should match what's needed here, but the API call needs testing too, so whatevah ;-)
      const JobsApi = new flamencoManager.JobsApi(this.apiClient);
      JobsApi.fetchJob(jobUpdate.id).then((job) => {
        console.log('API called successfully. Returned data: ' + job);
        console.log("Fetched job:", JSON.parse(JSON.stringify(job)));
        this.jobs.push(job);
      }, (error) => {
        console.error(error);
      });
    },
  }
};
</script>

<style lang='scss'>
@import "~vue-tabulator/dist/scss/bootstrap/tabulator_bootstrap4";

.flamenco-jobs-table {
  min-height: 50%;
  border: thick solid fuchsia;
}
</style>
