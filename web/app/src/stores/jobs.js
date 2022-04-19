import { defineStore } from 'pinia'

import * as urls from '@/urls'
import * as API from '@/manager-api';

// 'use' prefix is idiomatic for Pinia stores.
// See https://pinia.vuejs.org/core-concepts/
export const useJobs = defineStore('jobs', {
  state: () => ({
    /** @type API.Job[] */
    selectedJobs: [],
    /** @type API.Job */
    activeJob: null,
  }),
  getters: {
    numSelected() {
      return this.selectedJobs.length;
    },
    canDelete() {
      return this._anyJobWithStatus(["queued", "paused", "failed", "completed"])
    },
  },
  actions: {
    // Selection of jobs.
    setSelectedJob(job) {
      this.selectedJobs = [job];
      this.activeJob = job;
    },
    setSelectedJobs(jobs) {
      this.selectedJobs = jobs;
      this.activeJob = jobs[jobs.length-1]; // Last-selected is the active one.
    },
    deselectAllJobs() {
      this.selectedJobs = [];
      this.activeJob = null;
    },

    // Actions on the selected jobs.
    deleteJobs() {
      const deletionPromise = new Promise( (resolutionFunc, rejectionFunc) => {
        rejectionFunc({code: 327, message: "deleting jobs is not implemented in JS yet"});
      });
      return deletionPromise;
    },
    // Internal methods.
    _anyJobWithStatus(statuses) {
      return this.selectedJobs.reduce((foundJob, job) => (foundJob || statuses.includes(job.status)), false);
    }
  },
})
