import { defineStore } from 'pinia'

import * as API from '@/manager-api';
import { apiClient } from '@/stores/api-query-count';


const jobsAPI = new API.JobsApi(apiClient);

// 'use' prefix is idiomatic for Pinia stores.
// See https://pinia.vuejs.org/core-concepts/
export const useTasks = defineStore('tasks', {
  state: () => ({
    /** @type {API.Task[]} */
    selectedTasks: [],
    /** @type {API.Task} */
    activeTask: null,
    /**
     * ID of the active task. Easier to query than `activeTask ? activeTask.id : ""`.
     * @type {string}
     */
     activeTaskID: "",
  }),
  getters: {
    numSelected() {
      return this.selectedTasks.length;
    },
    canCancel() {
      return this._anyTaskWithStatus(["queued", "active", "soft-failed"])
    },
    canRequeue() {
      return this._anyTaskWithStatus(["canceled", "completed", "failed"])
    },
  },
  actions: {
    // Selection of tasks.
    setSelectedTask(task) {
      this.$patch({
        selectedTasks: [task],
        activeTask: task,
        activeTaskID: task.id,
      });
    },
    setSelectedTasks(tasks) {
      const activeTask =tasks[tasks.length-1]; // Last-selected is the active one.
      this.$patch({
        selectedTasks: tasks,
        activeTask: activeTask,
        activeTaskID: activeTask.id,
      });
    },
    deselectAllTasks() {
      this.$patch({
        selectedTasks: [],
        activeTask: null,
        activeTaskID: "",
      });
    },

    /**
     * Actions on the selected tasks.
     *
     * All the action functions return a promise that resolves when the action has been performed.
     *
     * TODO: actually have these work on all selected tasks. For simplicity, the
     * code now assumes that only the active task needs to be operated on.
     */
    cancelTasks() { return this._setTaskStatus("cancel-requested"); },
    requeueTasks() { return this._setTaskStatus("requeued"); },

    // Internal methods.

    /**
     *
     * @param {string[]} statuses
     * @returns bool indicating whether there is a selected task with any of the given statuses.
     */
    _anyTaskWithStatus(statuses) {
      return this.selectedTasks.reduce((foundTask, task) => (foundTask || statuses.includes(task.status)), false);
    },

    /**
     * Transition the selected task(s) to the new status.
     * @param {string} newStatus
     * @returns a Promise for the API request.
     */
    _setTaskStatus(newStatus) {
      const statuschange = new API.TaskStatusChange(newStatus, "requested from web interface");
      return jobsAPI.setTaskStatus(this.activeTask.id, statuschange);
    },
  },
})
