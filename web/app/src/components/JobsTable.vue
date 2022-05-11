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
import { toTitleCase } from '@/strings';
import { apiClient } from '@/stores/api-query-count';

import JobActionsBar from '@/components/JobActionsBar.vue'

export default {
  name: 'JobsTable',
  props: ["activeJobID"],
  emits: ["tableRowClicked"],
  components: {
    JobActionsBar,
  },
  data: () => {
    const options = {
      // See pkg/api/flamenco-manager.yaml, schemas Job and JobUpdate.
      columns: [
        // { formatter: "rowSelection", titleFormatter: "rowSelection", hozAlign: "center", headerHozAlign: "center", headerSort: false },
        {
          title: "ID", field: "id", headerSort: false,
          formatter: (cell) => cell.getData().id.substr(0, 8),
        },
        {
          title: 'Status', field: 'status', sorter: 'string',
          formatter(cell, formatterParams) { // eslint-disable-line no-unused-vars
            const cellValue = cell.getData();
            const label = toTitleCase(cellValue.status);
            return `<span title="${label}" class="indicator status-${cellValue.status}"></span>`;
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
      selectable: false, // The active job is tracked by click events, not row selection.
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

    // Set the `rowFormatter` here (instead of with the rest of the options
    // above) as it needs to refer to `this`, which isn't available in the
    // `data` function.
    this.options.rowFormatter = (row) => {
      const data = row.getData();
      const isActive = (data.id === this.activeJobID);
      row.getElement().classList.toggle("active-row", isActive);
    };
    this.tabulator = new Tabulator('#flamenco_job_list', this.options);
    this.tabulator.on("rowClick", this.onRowClick);
    this.tabulator.on("tableBuilt", this.fetchAllJobs);
  },
  watch: {
    activeJobID(newJobID, oldJobID) {
      this._reformatRow(oldJobID);
      this._reformatRow(newJobID);
    },
  },
  computed: {
    selectedIDs() {
      return this.tabulator.getSelectedData().map((job) => job.id);
    }
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

    onRowClick(event, row) {
      // Take a copy of the data, so that it's decoupled from the tabulator data
      // store. There were some issues where navigating to another job would
      // overwrite the old job's ID, and this prevents that.
      const rowData = plain(row.getData());
      this.$emit("tableRowClicked", rowData);
    },

    _reformatRow(jobID) {
      // Use tab.rowManager.findRow() instead of `tab.getRow()` as the latter
      // logs a warning when the row cannot be found.
      const row = this.tabulator.rowManager.findRow(jobID);
      if (!row) return
      if (row.reformat) row.reformat();
      else if (row.reinitialize) row.reinitialize(true);
    }
  }
};
</script>
