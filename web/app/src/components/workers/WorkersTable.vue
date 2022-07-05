<template>
  <h2 class="column-title">Workers</h2>
  <worker-actions-bar />
  <status-filter-bar
    :availableStatuses="availableStatuses"
    :activeStatuses="shownStatuses"
    classPrefix="worker-"
    @click="toggleStatusFilter"
  />
  <div>
    <div class="workers-list with-clickable-row" id="flamenco_workers_list"></div>
  </div>
</template>

<script lang="js">
import { TabulatorFull as Tabulator } from 'tabulator-tables';
import { WorkerMgtApi } from '@/manager-api'
import { indicator, workerStatus } from '@/statusindicator';
import { apiClient } from '@/stores/api-query-count';
import { useWorkers } from '@/stores/workers';

import StatusFilterBar from '@/components/StatusFilterBar.vue'
import WorkerActionsBar from '@/components/workers/WorkerActionsBar.vue'

export default {
  name: 'WorkersTable',
  props: ["activeWorkerID"],
  emits: ["tableRowClicked"],
  components: {
    StatusFilterBar,
    WorkerActionsBar,
  },
  data: () => {
    return {
      workers: useWorkers(),

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
          formatter: (cell) => {
            const data = cell.getData();
            const dot = indicator(data.status, 'worker-');
            const asString = workerStatus(data);
            return `${dot} ${asString}`;
          },
        },
        { title: 'Name', field: 'name', sorter: 'string' },
        { title: 'Version', field: 'version', sorter: 'string' },
      ],
      rowFormatter(row) {
        const data = row.getData();
        const isActive = (data.id === vueComponent.activeWorkerID);
        row.getElement().classList.toggle("active-row", isActive);
      },
      initialSort: [
        { column: "name", dir: "asc" },
      ],
      layout: "fitData",
      layoutColumnsOnNewData: true,
      height: "360px", // Must be set in order for the virtual DOM to function correctly.
      data: [], // Will be filled via a Flamenco API request.
      selectable: false, // The active worker is tracked by click events, not row selection.
    };
    this.tabulator = new Tabulator('#flamenco_workers_list', options);
    this.tabulator.on("rowClick", this.onRowClick);
    this.tabulator.on("tableBuilt", this._onTableBuilt);

     window.addEventListener('resize', this.recalcTableHeight);
  },
  unmounted() {
    window.removeEventListener('resize', this.recalcTableHeight);
  },
  watch: {
    activeWorkerID(newWorkerID, oldWorkerID) {
      this._reformatRow(oldWorkerID);
      this._reformatRow(newWorkerID);
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

      this.recalcTableHeight();
    },
    processWorkerUpdate(workerUpdate) {
      if (!this.tabulator.initialized) return;

      // Contrary to tabulator.getRow(), rowManager.findRow() doesn't log a
      // warning when the row cannot be found,
      const existingRow = this.tabulator.rowManager.findRow(workerUpdate.id);

      let promise;
      if (existingRow) {
        // Tabbulator doesn't update ommitted fields, but if `status_change`
        // is ommitted it means "no status change requested"; this should still
        // force an update of the `status_change` field.
        if (!workerUpdate.status_change) {
          workerUpdate.status_change = null;
        }
        promise = this.tabulator.updateData([workerUpdate]);
        // Tabulator doesn't know we're using 'status_change' in the 'status'
        // column, so it also won't know to redraw when that field changes.
        promise.then(() => existingRow.reinitialize(true));
      } else {
        promise = this.tabulator.addData([workerUpdate]);
      }
      promise
        .then(this.sortData)
        .then(() => { this.tabulator.redraw(); }) // Resize columns based on new data.
        .then(this.refreshAvailableStatuses);

      // TODO: this should also resize the columns, as the status column can
      // change sizes considerably.
    },

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
