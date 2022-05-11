<template>
  <div class="col col-1">
    <jobs-table ref="jobsTable" :activeJobID="jobID" @tableRowClicked="onTableJobClicked" />
  </div>
  <div class="col col-2">
    <job-details :jobData="jobs.activeJob" />
    <tasks-table v-if="jobID" ref="tasksTable" :jobID="jobID" @selectedTaskChange="onSelectedTaskChanged" />
  </div>
  <div class="col col-3">
    <task-details :taskData="tasks.activeTask" />
  </div>
  <footer>
    <notification-bar />
    <update-listener ref="updateListener" :websocketURL="websocketURL" :subscribedJob="jobID"
      @jobUpdate="onSioJobUpdate" @taskUpdate="onSioTaskUpdate" @message="onChatMessage"
      @sioReconnected="onSIOReconnected" @sioDisconnected="onSIODisconnected" />
  </footer>
</template>

<script>
import * as urls from '@/urls'
import * as API from '@/manager-api';
import { useJobs } from '@/stores/jobs';
import { useTasks } from '@/stores/tasks';
import { apiClient } from '@/stores/api-query-count';

import JobsTable from '@/components/JobsTable.vue'
import JobDetails from '@/components/JobDetails.vue'
import TaskDetails from '@/components/TaskDetails.vue'
import TasksTable from '@/components/TasksTable.vue'
import NotificationBar from '@/components/NotificationBar.vue'
import UpdateListener from '@/components/UpdateListener.vue'

export default {
  name: 'JobsView',
  props: ["jobID"], // provided by Vue Router.
  components: {
    JobsTable, JobDetails, TaskDetails, TasksTable, NotificationBar, UpdateListener,
  },
  data: () => ({
    websocketURL: urls.ws(),
    messages: [],

    jobs: useJobs(),
    tasks: useTasks(),
  }),
  mounted() {
    window.jobsView = this;
    this._fetchJob(this.jobID);
  },
  watch: {
    jobID(newJobID, oldJobID) {
      this._fetchJob(newJobID);
    },
  },
  methods: {
    onTableJobClicked(rowData) {
      this._routeToJob(rowData.id);
    },

    onSelectedTaskChanged(taskSummary) {
      if (!taskSummary) { // There is no selected task.
        this.tasks.deselectAllTasks();
        return;
      }
      console.log("selected task changed:", taskSummary);
      const jobsAPI = new API.JobsApi(apiClient);
      jobsAPI.fetchTask(taskSummary.id)
        .then((task) => {
          this.tasks.setSelectedTask(task);
          // Forward the full task to Tabulator, so that that gets updated too.
          if (this.$refs.tasksTable)
            this.$refs.tasksTable.processTaskUpdate(task);
        });
    },

    // SocketIO data event handlers:
    onSioJobUpdate(jobUpdate) {
      if (this.$refs.jobsTable) {
        if (jobUpdate.previous_status)
          this.$refs.jobsTable.processJobUpdate(jobUpdate);
        else
          this.$refs.jobsTable.processNewJob(jobUpdate);
      }
      if (this.jobID == jobUpdate.id)
        this._fetchJob(this.jobID);
    },

    },

    /**
     * @param {string} jobID job ID to navigate to, can be empty string for "no active job".
     */
    _routeToJob(jobID) {
      this.$router.push({ name: 'jobs', params: { jobID: jobID } });
    },

    /**
     * Fetch job info and set the active job once it's received.
     * @param {string} jobID job ID, can be empty string for "no job".
     */
    _fetchJob(jobID) {
      if (!jobID) {
        this.jobs.deselectAllJobs();
        return;
      }

      const jobsAPI = new API.JobsApi(apiClient);
      return jobsAPI.fetchJob(jobID)
        .then((job) => {
          this.jobs.setActiveJob(job);
          // Forward the full job to Tabulator, so that that gets updated too.
          this.$refs.jobsTable.processJobUpdate(job);
        });
    },

    /**
     * Event handler for SocketIO task updates.
     * @param {API.SocketIOTaskUpdate} taskUpdate
     */
    onSioTaskUpdate(taskUpdate) {
      if (this.$refs.tasksTable)
        this.$refs.tasksTable.processTaskUpdate(taskUpdate);
      if (this.tasks.activeTaskID == taskUpdate.id)
        this.onSelectedTaskChanged(taskUpdate);
    },

    onChatMessage(message) {
      console.log("chat message received:", message);
      this.messages.push(`${message.text}`);
    },

    // SocketIO connection event handlers:
    onSIOReconnected() {
      this.$refs.jobsTable.onReconnected();
      if (this.$refs.tasksTable)
        this.$refs.tasksTable.onReconnected();
    },
    onSIODisconnected(reason) {
    },
  },
}
</script>

<style scoped>
.isFetching {
  opacity: 50%;
}
</style>
