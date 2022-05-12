<template>
  <div class="col col-1">
    <jobs-table ref="jobsTable" :activeJobID="jobID" @tableRowClicked="onTableJobClicked" />
  </div>
  <div class="col col-2">
    <job-details :jobData="jobs.activeJob" />
    <tasks-table v-if="jobID" ref="tasksTable" :jobID="jobID" :taskID="taskID" @tableRowClicked="onTableTaskClicked" />
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
  props: ["jobID", "taskID"], // provided by Vue Router.
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

    this.jobs.$subscribe((mutation, state) => {
      console.log("Pinia mutation:", mutation)
      console.log("Pinia state   :", state)
      // // import { MutationType } from 'pinia'
      // mutation.type // 'direct' | 'patch object' | 'patch function'
      // // same as cartStore.$id
      // mutation.storeId // 'cart'
      // // only available with mutation.type === 'patch object'
      // mutation.payload // patch object passed to cartStore.$patch()
    })

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
      console.log("selected task changed:", taskSummary);
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
      if (this.$refs.jobsTable) {
        if (jobUpdate.previous_status)
          this.$refs.jobsTable.processJobUpdate(jobUpdate);
        else
          this.$refs.jobsTable.processNewJob(jobUpdate);
      }
      if (this.jobID == jobUpdate.id)
        this._fetchJob(this.jobID);
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
