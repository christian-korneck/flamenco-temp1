<template>
  <div class="job-list" id="flamenco_job_list">
  </div>
</template>

<script lang="js">
import { TabulatorFull as Tabulator } from 'tabulator-tables';
import { DateTime } from "luxon";

import {
  JobsApi,
} from '../manager-api'

export default {
  props: ["apiClient"],
  data: () => {
    const options = {
      // See pkg/api/flamenco-manager.yaml, schemas Job and JobUpdate.
      columns: [
        { title: 'ID', field: 'id', sorter: 'string', width: "15ex" },
        { title: 'Name', field: 'name', sorter: 'string' },
        { title: 'Status', field: 'status', sorter: 'string', width: 100 },
        { title: 'Type', field: 'type', sorter: 'string', width: 150 },
        { title: 'Prio', field: 'priority', sorter: 'number' },
        {
          title: 'Updated', field: 'updated',
          sorter: 'alphanum', sorterParams: { alignEmptyValues: "top" },
          formatter(cell, formatterParams) { // eslint-disable-line no-unused-vars
            const cellValue = cell.getData().updated;
            let updated = null;
            if (cellValue instanceof Date) {
              updated = DateTime.fromJSDate(cellValue);
            } else {
              updated = DateTime.fromISO(cellValue);
            }
            const now = DateTime.local();
            const ageInDays = now.diff(updated).as('days');
            if (ageInDays > 14)
              return updated.toLocaleString(DateTime.DATE_MED_WITH_WEEKDAY);
            return updated.toRelative();
          }
        },
      ],
      initialSort: [
        { column: "updated", dir: "desc" },
      ],
      height: "80%",
      data: [],
    };
    return {
      options: options,
    };
  },
  mounted() {
    // Allow testing from the JS console:
    // jobsTableVue.processJobUpdate({id: "ad0a5a00-5cb8-4e31-860a-8a405e75910e", status: "heyy", updated: DateTime.local().toISO(), previous_status: "uuuuh", name: "Updated manually"});
    // jobsTableVue.processJobUpdate({id: "ad0a5a00-5cb8-4e31-860a-8a405e75910e", status: "heyy", updated: DateTime.local().toISO()});
    window.jobsTableVue = this;
    this.tabulator = new Tabulator('#flamenco_job_list', this.options);
    this.fetchAllJobs();
  },
  methods: {
    onReconnected() {
      // If the connection to the backend was lost, we have likely missed some
      // updates. Just fetch the data and start from scratch.
      this.fetchAllJobs();
    },
    sortData() {
      const tab = this.tabulator;
      tab.setSort(tab.getSorters()); // This triggers re-sorting.
    },
    fetchAllJobs() {
      if (this.apiClient === undefined) {
        throw "no apiClient set on JobsTable component";
      }
      const jobsApi = new JobsApi(this.apiClient);
      const jobsQuery = {};
      jobsApi.queryJobs(jobsQuery).then(this.onJobsFetched, function (error) {
        // TODO: error handling.
        console.error(error);
      });
    },
    onJobsFetched(data) {
      this.tabulator.setData(data.jobs);
    },
    processJobUpdate(jobUpdate) {
      // updateData() will only overwrite properties that are actually set on
      // jobUpdate, and leave the rest as-is.
      this.tabulator.updateData([jobUpdate])
        .then(this.sortData);
    },
    processNewJob(jobUpdate) {
      console.log("processNewJob:", jobUpdate);
      // The update doesn't have all the info we need, so just fetch the job via an API call.
      const jobsApi = new JobsApi(this.apiClient);
      jobsApi.fetchJob(jobUpdate.id).then((job) => {
        console.log('API called successfully. Returned data: ' + job);
        console.log("Fetched job:", JSON.parse(JSON.stringify(job)));
        this.tabulator.addData([job])
          .then(this.sortData);
      }, (error) => {
        console.error(error);
      });
    },
  }
};
</script>

<style scoped>
.job-list {
  border: thick solid fuchsia;
  font-family: 'Noto Mono', monospace;
  font-size: smaller;
}
</style>
