import { defineStore } from 'pinia'

import * as API from '@/manager-api';
import { apiClient } from '@/stores/api-query-count';


const jobsAPI = new API.JobsApi(apiClient);

// 'use' prefix is idiomatic for Pinia stores.
// See https://pinia.vuejs.org/core-concepts/
export const useTasks = defineStore('tasks', {
  state: () => ({
    /** @type {API.Task} */
    activeTask: null,
    /**
     * ID of the active task. Easier to query than `activeTask ? activeTask.id : ""`.
     * @type {string}
     */
     activeTaskID: "",
  }),
  getters: {
    canCancel() {
      return this._anyTaskWithStatus(["queued", "active", "soft-failed"])
    },
    canRequeue() {
      return this._anyTaskWithStatus(["canceled", "completed", "failed"])
    },
  },
  actions: {
    setActiveTaskID(taskID) {
      this.$patch({
        activeTask: {id: taskID},
        activeTaskID: taskID,
      });
    },
    setActiveTask(task) {
      this.$patch({
        activeTask: task,
        activeTaskID: task.id,
      });
    },
    deselectAllTasks() {
      this.$patch({
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
    cancelTasks() { return this._setTaskStatus("canceled"); },
    requeueTasks() { return this._setTaskStatus("queued"); },

    // Internal methods.

    /**
     *
     * @param {string[]} statuses
     * @returns bool indicating whether there is a selected task with any of the given statuses.
     */
    _anyTaskWithStatus(statuses) {
      return !!this.activeTask && !!this.activeTask.status && statuses.includes(this.activeTask.status);
      // return this.selectedTasks.reduce((foundTask, task) => (foundTask || statuses.includes(task.status)), false);
    },

    /**
     * Transition the selected task(s) to the new status.
     * @param {string} newStatus
     * @returns a Promise for the API request.
     */
    _setTaskStatus(newStatus) {
      if (!this.activeTaskID) {
        console.warn(`_setTaskStatus(${newStatus}) impossible, no active task ID`);
        return;
      }
      const statuschange = new API.TaskStatusChange(newStatus, "requested from web interface");
      return jobsAPI.setTaskStatus(this.activeTaskID, statuschange);
    },
  },
})
