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
import { indicator } from '@/statusindicator';
import { apiClient } from '@/stores/api-query-count';
import { useTasks } from '@/stores/tasks';

import TaskActionsBar from '@/components/TaskActionsBar.vue'

export default {
  emits: ["tableRowClicked"],
  props: [
    "jobID", // ID of the job of which the tasks are shown here.
    "taskID", // The active task.
  ],
  components: {
    TaskActionsBar,
  },
  data: () => {
    const options = {
      // See pkg/api/flamenco-manager.yaml, schemas Task and TaskUpdate.
      columns: [
        // { formatter: "rowSelection", titleFormatter: "rowSelection", hozAlign: "center", headerHozAlign: "center", headerSort: false },
        // { title: "ID", field: "id", headerSort: false, formatter: (cell) => cell.getData().id.substr(0, 8), },
        {
          title: 'Status', field: 'status', sorter: 'string',
          formatter: (cell) => indicator(cell.getData().status),
        },
        { title: 'Name', field: 'name', sorter: 'string' },
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
      height: "300px", // Must be set in order for the virtual DOM to function correctly.
      data: [], // Will be filled via a Flamenco API request.
      selectable: false, // The active task is tracked by click events.
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

    // Set the `rowFormatter` here (instead of with the rest of the options
    // above) as it needs to refer to `this`, which isn't available in the
    // `data` function.
    this.options.rowFormatter = (row) => {
      const data = row.getData();
      const isActive = (data.id === this.taskID);
      row.getElement().classList.toggle("active-row", isActive);
    };

    this.tabulator = new Tabulator('#flamenco_task_list', this.options);
    this.tabulator.on("rowClick", this.onRowClick);
    this.tabulator.on("tableBuilt", this.fetchTasks);
  },
  watch: {
    jobID() {
      this.fetchTasks();
    },
    taskID(oldID, newID) {
      this._reformatRow(oldID);
      this._reformatRow(newID);
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
        this.tabulator.setData([]);
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
    },
    processTaskUpdate(taskUpdate) {
      // updateData() will only overwrite properties that are actually set on
      // taskUpdate, and leave the rest as-is.
      this.tabulator.updateData([taskUpdate])
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
