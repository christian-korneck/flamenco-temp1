<template>
  <div class="flamenco-jobs-table">
    <JobsTabulator
      ref="tabulator"
      :options="options"
    />
  </div>
</template>

<script>
/**
 * NOTE: the <JobsTabulator> component does NOT have a v-model attribute.
 * Instead of using Vue, data updates are handled on the Tabulator instance
 * directly. This gives more control over replacing data & updating existing
 * entries.
 */
import moment from "moment";
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
        { title: 'ID', field: 'id', sorter: 'string'},
        { title: 'Name', field: 'name', sorter: 'string' },
        { title: 'Status', field: 'status', sorter: 'string', width: 100 },
        { title: 'Type', field: 'type', sorter: 'string', width: 150 },
        { title: 'Prio', field: 'priority', sorter: 'number'},
        { title: 'Updated', field: 'updated', sorter: 'datetime',
          sorterParams: {
            format: "YYYY-MM-DD hh:mm:ss", // Format suitable for sorting chronologically.
            alignEmptyValues: "top",
          },
          formatter: function (cell, formatterParams) { // eslint-disable-line no-unused-vars
            const updated = moment(cell.getData().updated);
            const now = moment();
            const ageInDays = now.diff(updated, 'days');
            if (ageInDays > 14)
              return updated.format('YYYY-MM-DD HH:mm:ss')
            return updated.fromNow();
          }
        },
      ],
      initialSort:[
          {column:"updated", dir:"desc"},
      ],
      pagination: "local",
      paginationSize: 10,
    };
    return {
      options: options,
    };
  },
  mounted: function() {
    // Allow testing from the JS console:
    // jobsTableVue.processJobUpdate({id: "1e42bbf8-6a17-49b9-8a2a-2c80c6532465", status: "heyy", updated: moment().format(), previous_status: "uuuuh", name: "Updated manually"});
    // jobsTableVue.processJobUpdate({id: "7c26cac3-1811-42a8-8577-cd20faeb01a4", status: "heyy", updated: moment().format()});
    window.jobsTableVue = this;
    this.fetchAllJobs();
  },
  methods: {
    tabulator: function() {
      const tab = this.$refs.tabulator.getInstance();
      return tab;
    },
    sortData: function() {
      const tab = this.tabulator();
      tab.setSort(tab.getSorters()); // This triggers re-sorting.
    },
    fetchAllJobs: function() {
      const JobsApi = new flamencoManager.JobsApi(this.apiClient);
      const jobsQuery = {};
      JobsApi.queryJobs(jobsQuery).then(this.onJobsFetched, function(error) {
        // TODO: error handling.
        console.error(error);
      });
    },
    onJobsFetched: function(data) {
      this.tabulator().setData(data.jobs);
    },
    processJobUpdate: function(jobUpdate) {
      // updateData() will only overwrite properties that are actually set on
      // jobUpdate, and leave the rest as-is.
      this.tabulator().updateData([jobUpdate])
        .then(this.sortData);
    },
    processNewJob: function(jobUpdate) {
      console.log("processNewJob:", jobUpdate);
      // The update doesn't have all the info we need, so just fetch the job via an API call.
      const JobsApi = new flamencoManager.JobsApi(this.apiClient);
      JobsApi.fetchJob(jobUpdate.id).then((job) => {
        console.log('API called successfully. Returned data: ' + job);
        console.log("Fetched job:", JSON.parse(JSON.stringify(job)));
        this.tabulator().addData([job])
          .then(this.sortData);
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
