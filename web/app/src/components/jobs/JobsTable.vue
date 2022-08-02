<template>
  <h2 class="column-title">Jobs</h2>
  <div class="btn-bar-group">
    <job-actions-bar />
    <div class="align-right">
      <status-filter-bar :availableStatuses="availableStatuses" :activeStatuses="shownStatuses"
        @click="toggleStatusFilter" />
    </div>
  </div>
  <div>
    <div class="job-list with-clickable-row" id="flamenco_job_list"></div>
  </div>
</template>

<script>
import { TabulatorFull as Tabulator } from 'tabulator-tables';
import * as datetime from "@/datetime";
import * as API from '@/manager-api'
import { indicator } from '@/statusindicator';
import { apiClient } from '@/stores/api-query-count';

import JobActionsBar from '@/components/jobs/JobActionsBar.vue'
import StatusFilterBar from '@/components/StatusFilterBar.vue'

export default {
  name: 'JobsTable',
  props: ["activeJobID"],
  emits: ["tableRowClicked"],
  components: {
    JobActionsBar, StatusFilterBar,
  },
  data: () => {
    return {
      shownStatuses: [],
      availableStatuses: [], // Will be filled after data is loaded from the backend.
    };
  },
  mounted() {
    // Allow testing from the JS console:
    // jobsTableVue.processJobUpdate({id: "ad0a5a00-5cb8-4e31-860a-8a405e75910e", status: "heyy", updated: DateTime.local().toISO(), previous_status: "uuuuh", name: "Updated manually"});
    // jobsTableVue.processJobUpdate({id: "ad0a5a00-5cb8-4e31-860a-8a405e75910e", status: "heyy", updated: DateTime.local().toISO()});
    window.jobsTableVue = this;

    const vueComponent = this;
    const options = {
      // See pkg/api/flamenco-openapi.yaml, schemas Job and SocketIOJobUpdate.
      columns: [
        // Useful for debugging when there are many similar jobs:
        // { title: "ID", field: "id", headerSort: false, formatter: (cell) => cell.getData().id.substr(0, 8), },
        {
          title: 'Status', field: 'status', sorter: 'string',
          formatter: (cell) => {
            const status = cell.getData().status;
            const dot = indicator(status);
            return `${dot} ${status}`;
          },
        },
        { title: 'Name', field: 'name', sorter: 'string' },
        { title: 'Type', field: 'type', sorter: 'string' },
        { title: 'Prio', field: 'priority', sorter: 'number' },
        {
          title: 'Updated', field: 'updated',
          sorter: 'alphanum', sorterParams: { alignEmptyValues: "top" },
          formatter(cell) {
            const cellValue = cell.getData().updated;
            // TODO: if any "{amount} {units} ago" shown, the table should be
            // refreshed every few {units}, so that it doesn't show any stale "4
            // seconds ago" for days.
            return datetime.relativeTime(cellValue);
          },
        },
      ],
      rowFormatter(row) {
        const data = row.getData();
        const isActive = (data.id === vueComponent.activeJobID);
        row.getElement().classList.toggle("active-row", isActive);
      },
      initialSort: [
        { column: "updated", dir: "desc" },
      ],
      layout: "fitData",
      layoutColumnsOnNewData: true,
      height: "720px", // Must be set in order for the virtual DOM to function correctly.
      data: [], // Will be filled via a Flamenco API request.
      selectable: false, // The active job is tracked by click events, not row selection.
    };
    this.tabulator = new Tabulator('#flamenco_job_list', options);
    this.tabulator.on("rowClick", this.onRowClick);
    this.tabulator.on("tableBuilt", this._onTableBuilt);

    window.addEventListener('resize', this.recalcTableHeight);
  },
  unmounted() {
    window.removeEventListener('resize', this.recalcTableHeight);
  },
  watch: {
    activeJobID(newJobID, oldJobID) {
      this._reformatRow(oldJobID);
      this._reformatRow(newJobID);
    },
    availableStatuses() {
      // Statuses changed, so the filter bar could have gone from "no statuses"
      // to "any statuses" (or one row of filtering stuff to two, I don't know)
      // and changed height.
      this.$nextTick(this.recalcTableHeight);
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
    _onTableBuilt() {
      this.tabulator.setFilter(this._filterByStatus);
      this.fetchAllJobs();
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
      this._refreshAvailableStatuses();

      this.recalcTableHeight();
    },
    processJobUpdate(jobUpdate) {
      // updateData() will only overwrite properties that are actually set on
      // jobUpdate, and leave the rest as-is.
      if (this.tabulator.initialized) {
        this.tabulator.updateData([jobUpdate])
          .then(this.sortData)
          .then(() => { this.tabulator.redraw(); }) // Resize columns based on new data.
      }
      this._refreshAvailableStatuses();
    },
    processNewJob(jobUpdate) {
      if (this.tabulator.initialized) {
        this.tabulator.addData([jobUpdate])
          .then(this.sortData)
          .then(() => { this.tabulator.redraw(); }) // Resize columns based on new data.
      }

      this._refreshAvailableStatuses();
    },

    onRowClick(event, row) {
      // Take a copy of the data, so that it's decoupled from the tabulator data
      // store. There were some issues where navigating to another job would
      // overwrite the old job's ID, and this prevents that.
      const rowData = plain(row.getData());
      this.$emit("tableRowClicked", rowData);
    },
    toggleStatusFilter(status) {
      const asSet = new Set(this.shownStatuses);
      if (!asSet.delete(status)) {
        asSet.add(status);
      }
      this.shownStatuses = Array.from(asSet).sort();
      this.tabulator.refreshFilter();
    },
    _filterByStatus(job) {
      if (this.shownStatuses.length == 0) {
        return true;
      }
      return this.shownStatuses.indexOf(job.status) >= 0;
    },
    _refreshAvailableStatuses() {
      const statuses = new Set();
      for (let row of this.tabulator.getData()) {
        statuses.add(row.status);
      }
      this.availableStatuses = Array.from(statuses).sort();
    },

    _reformatRow(jobID) {
      // Use tab.rowManager.findRow() instead of `tab.getRow()` as the latter
      // logs a warning when the row cannot be found.
      const row = this.tabulator.rowManager.findRow(jobID);
      if (!row) return
      if (row.reformat) row.reformat();
      else if (row.reinitialize) row.reinitialize(true);
    },

    /**
     * Recalculate the appropriate table height to fit in the column without making that scroll.
     */
    recalcTableHeight() {
      if (!this.tabulator.initialized) {
        // Sometimes this function is called too early, before the table was initialised.
        // After the table is initialised it gets resized anyway, so this call can be ignored.
        return;
      }
      const table = this.tabulator.element;
      const tableContainer = table.parentElement;
      const outerContainer = tableContainer.parentElement;
      if (!outerContainer) {
        // This can happen when the component was removed before the function is
        // called. This is possible due to the use of Vue's `nextTick()`
        // function.
        return;
      }

      const availableHeight = outerContainer.clientHeight - 12; // TODO: figure out where the -12 comes from.

      if (tableContainer.offsetParent != tableContainer.parentElement) {
        // `offsetParent` is assumed to be the actual column in the 3-column
        // view. To ensure this, it's given `position: relative` in the CSS
        // styling.
        console.warn("JobsTable.recalcTableHeight() only works when the offset parent is the real parent of the element.");
        return;
      }

      const tableHeight = availableHeight - tableContainer.offsetTop;
      this.tabulator.setHeight(tableHeight);
    },
  },
};
</script>
