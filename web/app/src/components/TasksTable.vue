<template>
  <task-actions-bar />
  <div class="task-list-container">
    <div class="task-list" id="flamenco_task_list"></div>
  </div>
</template>

<script lang="js">
import { TabulatorFull as Tabulator } from 'tabulator-tables';
import * as datetime from "@/datetime";
import * as API from '@/manager-api'
import { apiClient } from '@/stores/api-query-count';
import { useTasks } from '@/stores/tasks';

import TaskActionsBar from '@/components/TaskActionsBar.vue'

export default {
  emits: ["selectedTaskChange"],
  props: [
    "jobID", // ID of the job of which the tasks are shown here.
  ],
  components: {
    TaskActionsBar,
  },
  data: () => {
    const options = {
      // See pkg/api/flamenco-manager.yaml, schemas Task and TaskUpdate.
      columns: [
        { formatter: "rowSelection", titleFormatter: "rowSelection", hozAlign: "center", headerHozAlign: "center", headerSort: false },
        { title: 'ID', field: 'id', sorter: 'string', width: "12%" },
        { title: 'Name', field: 'name', sorter: 'string' },
        { title: 'Status', field: 'status', sorter: 'string' },
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
      height: "300px",
      data: [], // Will be filled via a Flamenco API request.
      selectable: 1, // Only allow a single row to be selected at a time.
    };
    return {
      options: options,
      tasks: useTasks(),
    };
  },
  mounted() {
    // Allow testing from the JS console:
    // tasksTableVue.processTaskUpdate({id: "ad0a5a00-5cb8-4e31-860a-8a405e75910e", status: "heyy", updated: DateTime.local().toISO(), previous_status: "uuuuh", name: "Updated manually"});
    // tasksTableVue.processTaskUpdate({id: "ad0a5a00-5cb8-4e31-860a-8a405e75910e", status: "heyy", updated: DateTime.local().toISO()});
    window.tasksTableVue = this;
    this.tabulator = new Tabulator('#flamenco_task_list', this.options);
    this.tabulator.on("rowSelected", this.onRowSelected);
    this.tabulator.on("rowDeselected", this.onRowDeselected);
    this.fetchTasks();
  },
  watch: {
    jobID() {
      this.onRowDeselected([]);
      this.fetchTasks();
    },
  },
  methods: {
    onReconnected() {
      // If the connection to the backend was lost, we have likely missed some
      // updates. Just fetch the data and start from scratch.
      this.fetchTasks();
    },
    sortData() {
      const tab = this.tabulator;
      tab.setSort(tab.getSorters()); // This triggers re-sorting.
    },
    fetchTasks() {
      console.log("Fetching tasks for job", this.jobID);
      if (!this.jobID) {
        // Prevent a warning when fetchTasks() is called before the tabulator is
        // properly initialised. After initialisation the data is empty anyway.
        if (this.tabulator.initialized) {
          this.tabulator.setData([]);
        }
        return;
      }

      const jobsApi = new API.JobsApi(apiClient);
      jobsApi.fetchJobTasks(this.jobID).then(this.onTasksFetched, function (error) {
        // TODO: error handling.
        console.error(error);
      });
    },
    onTasksFetched(data) {
      // "Down-cast" to TaskUpdate to only get those fields, just for debugging things:
      // let tasks = data.tasks.map((j) => API.TaskUpdate.constructFromObject(j));
      this.tabulator.setData(data.tasks);
      this._restoreRowSelection();
    },
    processTaskUpdate(taskUpdate) {
      // updateData() will only overwrite properties that are actually set on
      // taskUpdate, and leave the rest as-is.
      this.tabulator.updateData([taskUpdate])
        .then(this.sortData);
    },

    // Selection handling.
    onRowSelected(selectedRow) {
      const selectedData = selectedRow.getData();
      this._storeRowSelection([selectedData]);
      this.$emit("selectedTaskChange", selectedData);
    },
    onRowDeselected(deselectedRow) {
      this._storeRowSelection([]);
      this.$emit("selectedTaskChange", null);
    },
    _storeRowSelection(selectedData) {
      const selectedTaskIDs = selectedData.map((row) => row.id);
      localStorage.setItem("selectedTaskIDs", selectedTaskIDs);
    },
    _restoreRowSelection() {
      // const selectedTaskIDs = localStorage.getItem('selectedTaskIDs');
      // if (!selectedTaskIDs) {
      //   return;
      // }
      // this.tabulator.selectRow(selectedTaskIDs);
    },
  }
};
</script>

<style>
.task-list-container {
  font-family: 'Noto Mono', monospace;
  font-size: smaller;
  max-height: 300px;
}
</style>
