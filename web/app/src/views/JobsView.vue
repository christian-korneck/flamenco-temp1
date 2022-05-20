<template>
  <div class="col col-1">
    <jobs-table ref="jobsTable" :activeJobID="jobID" @tableRowClicked="onTableJobClicked" />
  </div>
  <div class="col col-2">
    <job-details :jobData="jobs.activeJob" />
    <tasks-table v-if="hasJobData" ref="tasksTable" :jobID="jobID" :taskID="taskID" @tableRowClicked="onTableTaskClicked" />
  </div>
  <div class="col col-3">
    <task-details :taskData="tasks.activeTask" />
  </div>

  <footer class="window-footer" v-if="!showFooterPopup" @click="showFooterPopup = true"><notification-bar /></footer>
  <footer-popup v-if="showFooterPopup" ref="footerPopup" @clickClose="showFooterPopup = false" />

  <update-listener ref="updateListener" :websocketURL="websocketURL" :subscribedJobID="jobID"
    @jobUpdate="onSioJobUpdate" @taskUpdate="onSioTaskUpdate" @message="onChatMessage"
    @sioReconnected="onSIOReconnected" @sioDisconnected="onSIODisconnected" />
</template>

<script>
import * as urls from '@/urls'
import * as API from '@/manager-api';
import { useJobs } from '@/stores/jobs';
import { useTasks } from '@/stores/tasks';
import { useNotifs } from '@/stores/notifications'
import { apiClient } from '@/stores/api-query-count';

import FooterPopup from '@/components/FooterPopup.vue'
import JobDetails from '@/components/JobDetails.vue'
import JobsTable from '@/components/JobsTable.vue'
import NotificationBar from '@/components/NotificationBar.vue'
import TaskDetails from '@/components/TaskDetails.vue'
import TasksTable from '@/components/TasksTable.vue'
import UpdateListener from '@/components/UpdateListener.vue'

export default {
  name: 'JobsView',
  props: ["jobID", "taskID"], // provided by Vue Router.
  components: {
    FooterPopup,
    JobDetails,
    JobsTable,
    NotificationBar,
    TaskDetails,
    TasksTable,
    UpdateListener,
  },
  data: () => ({
    websocketURL: urls.ws(),
    messages: [],

    jobs: useJobs(),
    tasks: useTasks(),
    notifs: useNotifs(),
    showFooterPopup: false,
  }),
  computed: {
    hasJobData() {
      return !objectEmpty(this.jobs.activeJob);
    },
  },
  mounted() {
    window.jobsView = this;
    window.footerPopup = this.$refs.footerPopup;

    // Useful for debugging:
    // this.jobs.$subscribe((mutation, state) => {
    //   console.log("Pinia mutation:", mutation)
    //   console.log("Pinia state   :", state)
    // })

    this._fetchJob(this.jobID);
    this._fetchTask(this.taskID);
  },
  watch: {
    jobID(newJobID, oldJobID) {
      this._fetchJob(newJobID);
    },
    taskID(newTaskID, oldTaskID) {
      this._fetchTask(newTaskID);
    },
  },
  methods: {
    onTableJobClicked(rowData) {
      // Don't route to the current job, as that'll deactivate the current task.
      if (rowData.id == this.jobID) return;
      this._routeToJob(rowData.id);
    },
    onTableTaskClicked(rowData) {
      this._routeToTask(rowData.id);
    },

    onSelectedTaskChanged(taskSummary) {
      if (!taskSummary) { // There is no active task.
        this.tasks.deselectAllTasks();
        return;
      }

      const jobsAPI = new API.JobsApi(apiClient);
      jobsAPI.fetchTask(taskSummary.id)
        .then((task) => {
          this.tasks.setActiveTask(task);
          // Forward the full task to Tabulator, so that that gets updated too.
          if (this.$refs.tasksTable)
            this.$refs.tasksTable.processTaskUpdate(task);
        });
    },

    // SocketIO data event handlers:
    onSioJobUpdate(jobUpdate) {
      this.notifs.addJobUpdate(jobUpdate);

      if (this.$refs.jobsTable) {
        if (jobUpdate.previous_status)
          this.$refs.jobsTable.processJobUpdate(jobUpdate);
        else
          this.$refs.jobsTable.processNewJob(jobUpdate);
      }
      if (this.jobID != jobUpdate.id)
        return;

      this._fetchJob(this.jobID);
      if (jobUpdate.refresh_tasks) {
        if (this.$refs.tasksTable)
          this.$refs.tasksTable.fetchTasks();
        this._fetchTask(this.taskID);
      }
    },

    /**
     * Event handler for SocketIO task updates.
     * @param {API.SocketIOTaskUpdate} taskUpdate
     */
    onSioTaskUpdate(taskUpdate) {
      if (this.$refs.tasksTable)
        this.$refs.tasksTable.processTaskUpdate(taskUpdate);
      if (this.taskID == taskUpdate.id)
        this._fetchTask(this.taskID);
      this.notifs.addTaskUpdate(taskUpdate);
    },

    /**
     * @param {string} jobID job ID to navigate to, can be empty string for "no active job".
     */
    _routeToJob(jobID) {
      const route = { name: 'jobs', params: { jobID: jobID } };
      console.log("routing to job", route.params);
      this.$router.push(route);
    },
    /**
     * @param {string} taskID task ID to navigate to within this job, can be
     * empty string for "no active task".
     */
    _routeToTask(taskID) {
      const route = { name: 'jobs', params: { jobID: this.jobID, taskID: taskID } };
      console.log("routing to task", route.params);
      this.$router.push(route);
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
     * Fetch task info and set the active task once it's received.
     * @param {string} taskID task ID, can be empty string for "no task".
     */
    _fetchTask(taskID) {
      if (!taskID) {
        this.tasks.deselectAllTasks();
        return;
      }

      const jobsAPI = new API.JobsApi(apiClient);
      return jobsAPI.fetchTask(taskID)
        .then((task) => {
          this.tasks.setActiveTask(task);
          // Forward the full task to Tabulator, so that that gets updated too.
          this.$refs.tasksTable.processTaskUpdate(task);
        });
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
