<template>
  <div>
    <h2 class="column-title">Jobs</h2>

    <job-actions-bar />
    <div class="job-list" id="flamenco_job_list"></div>
  </div>
</template>

<script lang="js">
import { TabulatorFull as Tabulator } from 'tabulator-tables';
import * as datetime from "@/datetime";
import * as API from '@/manager-api'
import { apiClient } from '@/stores/api-query-count';

import JobActionsBar from '@/components/JobActionsBar.vue'

export default {
  emits: ["selectedJobChange"],
  components: {
    JobActionsBar,
  },
  data: () => {
    const options = {
      // See pkg/api/flamenco-manager.yaml, schemas Job and JobUpdate.
      columns: [
        { formatter: "rowSelection", titleFormatter: "rowSelection", hozAlign: "center", headerHozAlign: "center", headerSort: false },
        {
          title: 'Status', field: 'status', sorter: 'string',
          formatter(cell, formatterParams) { // eslint-disable-line no-unused-vars
            const cellValue = cell.getData();
            console.log(cellValue);
            return '<span title="' + cellValue.status + '" class="indicator status-' + cellValue.status + '"></span>';
          }
        },
        { title: 'Name', field: 'name', sorter: 'string' },
        { title: 'Type', field: 'type', sorter: 'string' },
        { title: 'Prio', field: 'priority', sorter: 'number' },
        {
          title: 'Updated', field: 'updated',
          sorter: 'alphanum', sorterParams: { alignEmptyValues: "top" },
          formatter(cell, formatterParams) { // eslint-disable-line no-unused-vars
            const cellValue = cell.getData().updated;
            // TODO: if any "{amount} {units} ago" shown, the table should be
            // refreshed every few {units}, so that it doesn't show any stale "4
            // seconds ago" for days.
            return datetime.relativeTime(cellValue);
          }
        },
      ],
      initialSort: [
        { column: "updated", dir: "desc" },
      ],
      data: [], // Will be filled via a Flamenco API request.
      selectable: 1, // Only allow a single row to be selected at a time.
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
    this.tabulator.on("rowSelected", this.onRowSelected);
    this.tabulator.on("rowDeselected", this.onRowDeselected);
    this.tabulator.on("tableBuilt", this.fetchAllJobs);
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
      const jobsApi = new API.JobsApi(apiClient);
      const jobsQuery = {};
      jobsApi.queryJobs(jobsQuery).then(this.onJobsFetched, function (error) {
        // TODO: error handling.
        console.error(error);
      });
    },
    onJobsFetched(data) {
      // "Down-cast" to JobUpdate to only get those fields, just for debugging things:
      // data.jobs = data.jobs.map((j) => API.JobUpdate.constructFromObject(j));
      this.tabulator.setData(data.jobs);
      this._restoreRowSelection();
    },
    processJobUpdate(jobUpdate) {
      // updateData() will only overwrite properties that are actually set on
      // jobUpdate, and leave the rest as-is.
      this.tabulator.updateData([jobUpdate])
        .then(this.sortData);
    },
    processNewJob(jobUpdate) {
      this.tabulator.addData([jobUpdate])
        .then(this.sortData);
    },

    // Selection handling.
    onRowSelected(selectedRow) {
      const selectedData = selectedRow.getData();
      this._storeRowSelection([selectedData]);
      this.$emit("selectedJobChange", selectedData);
    },
    onRowDeselected(deselectedRow) {
      this._storeRowSelection([]);
      this.$emit("selectedJobChange", null);
    },
    _storeRowSelection(selectedData) {
      const selectedJobIDs = selectedData.map((row) => row.id);
      localStorage.setItem("selectedJobIDs", selectedJobIDs);
    },
    _restoreRowSelection() {
      const selectedJobIDs = localStorage.getItem('selectedJobIDs');
      if (!selectedJobIDs) {
        return;
      }
      this.tabulator.selectRow(selectedJobIDs);
    },
  }
};
</script>
