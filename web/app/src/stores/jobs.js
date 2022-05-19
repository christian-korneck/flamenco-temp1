import { defineStore } from 'pinia'

import * as API from '@/manager-api';
import { apiClient } from '@/stores/api-query-count';


const jobsAPI = new API.JobsApi(apiClient);

// 'use' prefix is idiomatic for Pinia stores.
// See https://pinia.vuejs.org/core-concepts/
export const useJobs = defineStore('jobs', {
  state: () => ({
    /** @type {API.Job} */
    activeJob: null,
    /**
     * ID of the active job. Easier to query than `activeJob ? activeJob.id : ""`.
     * @type {string}
     */
    activeJobID: "",
  }),
  getters: {
    canDelete() {
      return this._anyJobWithStatus(["queued", "paused", "failed", "completed", "canceled"])
    },
    canCancel() {
      return this._anyJobWithStatus(["queued", "active", "failed"])
    },
    canRequeue() {
      return this._anyJobWithStatus(["canceled", "completed", "failed", "paused"])
    },
  },
  actions: {
    setActiveJobID(jobID) {
      this.$patch({
        activeJob: {id: jobID, settings: {}, metadata: {}},
        activeJobID: jobID,
      });
    },
    setActiveJob(job) {
      // The "function" form of $patch is necessary here, as otherwise it'll
      // merge `job` into `state.activeJob`. As a result, it won't touch missing
      // keys, which means that metadata fields that existed on the previous job
      // but not on the new one will still linger around. By passing a function
      // to `$patch` this is resolved.
      this.$patch((state) => {
        state.activeJob = job;
        state.activeJobID = job.id;
        state.hasChanged = true;
      });
    },
    deselectAllJobs() {
      this.$patch({
        activeJob: null,
        activeJobID: "",
      });
    },

    /**
     * Actions on the selected jobs.
     *
     * All the action functions return a promise that resolves when the action has been performed.
     *
     * TODO: actually have these work on all selected jobs. For simplicity, the
     * code now assumes that only the active job needs to be operated on.
     */
    cancelJobs() { return this._setJobStatus("cancel-requested"); },
    requeueJobs() { return this._setJobStatus("requeueing"); },

    // Internal methods.

    /**
     *
     * @param {string[]} statuses
     * @returns bool indicating whether there is a selected job with any of the given statuses.
     */
    _anyJobWithStatus(statuses) {
      return !!this.activeJob && !!this.activeJob.status && statuses.includes(this.activeJob.status);
      // return this.selectedJobs.reduce((foundJob, job) => (foundJob || statuses.includes(job.status)), false);
    },

    /**
     * Transition the selected job(s) to the new status.
     * @param {string} newStatus
     * @returns a Promise for the API request.
     */
    _setJobStatus(newStatus) {
      if (!this.activeJobID) {
        console.warn(`_setJobStatus(${newStatus}) impossible, no active job ID`);
        return;
      }
      const statuschange = new API.JobStatusChange(newStatus, "requested from web interface");
      return jobsAPI.setJobStatus(this.activeJobID, statuschange);
    },
  },
})
