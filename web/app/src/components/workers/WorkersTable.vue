<template>
  <div>
    <h2 class="column-title">Workers</h2>
    <status-filter-bar
      :availableStatuses="availableStatuses"
      :activeStatuses="shownStatuses"
      classPrefix="worker-"
      @click="toggleStatusFilter"
    />
    <div class="workers-list with-clickable-row" id="flamenco_workers_list"></div>
  </div>
</template>

<script lang="js">
import { TabulatorFull as Tabulator } from 'tabulator-tables';
import { WorkerMgtApi } from '@/manager-api'
import { indicator } from '@/statusindicator';
import { apiClient } from '@/stores/api-query-count';

import StatusFilterBar from '@/components/StatusFilterBar.vue'

export default {
  name: 'WorkersTable',
  props: ["activeWorkerID"],
  emits: ["tableRowClicked"],
  components: {
    StatusFilterBar,
  },
  data: () => {
    return {
      shownStatuses: [],
      availableStatuses: [], // Will be filled after data is loaded from the backend.
    };
  },
  mounted() {
    window.workersTableVue = this;

    const vueComponent = this;
    const options = {
      // See pkg/api/flamenco-openapi.yaml, schemas WorkerSummary and SocketIOWorkerUpdate.
      columns: [
        // Useful for debugging when there are many similar workers:
        // { title: "ID", field: "id", headerSort: false, formatter: (cell) => cell.getData().id.substr(0, 8), },
        {
          title: 'Status', field: 'status', sorter: 'string',
          formatter: (cell) => indicator(cell.getData().status, 'worker-'),
        },
        { title: 'Name', field: 'nickname', sorter: 'string' },
      ],
      rowFormatter(row) {
        const data = row.getData();
        const isActive = (data.id === vueComponent.activeWorkerID);
        row.getElement().classList.toggle("active-row", isActive);
      },
      initialSort: [
        { column: "nickname", dir: "asc" },
      ],
      height: "720px", // Must be set in order for the virtual DOM to function correctly.
      data: [], // Will be filled via a Flamenco API request.
      selectable: false, // The active worker is tracked by click events, not row selection.
    };
    this.tabulator = new Tabulator('#flamenco_workers_list', options);
    this.tabulator.on("rowClick", this.onRowClick);
    this.tabulator.on("tableBuilt", this._onTableBuilt);
  },
  watch: {
    activeWorkerID(newWorkerID, oldWorkerID) {
      this._reformatRow(oldWorkerID);
      this._reformatRow(newWorkerID);
    },
  },
  computed: {
    selectedIDs() {
      return this.tabulator.getSelectedData().map((worker) => worker.id);
    }
  },
  methods: {
    onReconnected() {
      // If the connection to the backend was lost, we have likely missed some
      // updates. Just fetch the data and start from scratch.
      this.fetchAllWorkers();
    },
    sortData() {
      const tab = this.tabulator;
      tab.setSort(tab.getSorters()); // This triggers re-sorting.
    },
    _onTableBuilt() {
      this.tabulator.setFilter(this._filterByStatus);
      this.fetchAllWorkers();
    },
    fetchAllWorkers() {
      const api = new WorkerMgtApi(apiClient);
      api.fetchWorkers().then(this.onWorkersFetched, function (error) {
        // TODO: error handling.
        console.error(error);
      });
    },
    onWorkersFetched(data) {
      this.tabulator.setData(data.workers);
      this._refreshAvailableStatuses();
    },
    // processWorkerUpdate(workerUpdate) {
    //   // updateData() will only overwrite properties that are actually set on
    //   // workerUpdate, and leave the rest as-is.
    //   if (this.tabulator.initialized) {
    //     this.tabulator.updateData([workerUpdate])
    //       .then(this.sortData);
    //   }
    //   this._refreshAvailableStatuses();
    // },
    // processNewWorker(workerUpdate) {
    //   if (this.tabulator.initialized) {
    //     this.tabulator.updateData([workerUpdate])
    //       .then(this.sortData);
    //   }
    //   this.tabulator.addData([workerUpdate])
    //     .then(this.sortData);
    //   this._refreshAvailableStatuses();
    // },

    onRowClick(event, row) {
      // Take a copy of the data, so that it's decoupled from the tabulator data
      // store. There were some issues where navigating to another worker would
      // overwrite the old worker's ID, and this prevents that.
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
    _filterByStatus(worker) {
      if (this.shownStatuses.length == 0) {
        return true;
      }
      return this.shownStatuses.indexOf(worker.status) >= 0;
    },
    _refreshAvailableStatuses() {
      const statuses = new Set();
      for (let row of this.tabulator.getData()) {
        statuses.add(row.status);
      }
      this.availableStatuses = Array.from(statuses).sort();
    },

    _reformatRow(workerID) {
      // Use tab.rowManager.findRow() instead of `tab.getRow()` as the latter
      // logs a warning when the row cannot be found.
      const row = this.tabulator.rowManager.findRow(workerID);
      if (!row) return
      if (row.reformat) row.reformat();
      else if (row.reinitialize) row.reinitialize(true);
    },
  },
};
</script>
