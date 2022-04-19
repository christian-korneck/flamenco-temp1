import { defineStore } from 'pinia'

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
  },
})
