<template>
  <div class="col col-1">
    <jobs-table ref="jobsTable" :activeJobID="jobID" @tableRowClicked="onTableJobClicked" />
  </div>
  <div class="col col-2 job-details-column" id="col-job-details">
    <job-details ref="jobDetails" :jobData="jobs.activeJob" @reshuffled="_recalcTasksTableHeight" />
    <tasks-table v-if="hasJobData" ref="tasksTable" :jobID="jobID" :taskID="taskID" @tableRowClicked="onTableTaskClicked" />
  </div>
  <div class="col col-3">
    <task-details :taskData="tasks.activeTask" />
  </div>

  <footer class="app-footer" v-if="!showFooterPopup" @click="showFooterPopup = true">
    <notification-bar />
    <div class="app-footer-expand">
      <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <line x1="12" y1="19" x2="12" y2="5"></line><polyline points="5 12 12 5 19 12"></polyline>
      </svg>
    </div>
  </footer>
  <footer-popup v-if="showFooterPopup" ref="footerPopup" @clickClose="showFooterPopup = false" />

  <update-listener ref="updateListener"  mainSubscription="allJobs"
    :subscribedJobID="jobID" :subscribedTaskID="taskID"
    @jobUpdate="onSioJobUpdate" @taskUpdate="onSioTaskUpdate" @taskLogUpdate="onSioTaskLogUpdate"
    @lastRenderedUpdate="onSioLastRenderedUpdate"
    @message="onChatMessage"
    @sioReconnected="onSIOReconnected" @sioDisconnected="onSIODisconnected" />
</template>

<script>
import * as API from '@/manager-api';
import { useJobs } from '@/stores/jobs';
import { useTasks } from '@/stores/tasks';
import { useNotifs } from '@/stores/notifications'
import { useTaskLog } from '@/stores/tasklog'
import { apiClient } from '@/stores/api-query-count';

import FooterPopup from '@/components/footer/FooterPopup.vue'
import JobDetails from '@/components/jobs/JobDetails.vue'
import JobsTable from '@/components/jobs/JobsTable.vue'
import NotificationBar from '@/components/footer/NotificationBar.vue'
import TaskDetails from '@/components/jobs/TaskDetails.vue'
import TasksTable from '@/components/jobs/TasksTable.vue'
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
    messages: [],

    jobs: useJobs(),
    tasks: useTasks(),
    notifs: useNotifs(),
    taskLog: useTaskLog(),
    showFooterPopup: !!localStorage.getItem("footer-popover-visible"),
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
    showFooterPopup(shown) {
      if (shown) localStorage.setItem("footer-popover-visible", "true");
      else localStorage.removeItem("footer-popover-visible");
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
     * Event handler for SocketIO task log updates.
     * @param {API.SocketIOTaskLogUpdate} taskLogUpdate
     */
    onSioTaskLogUpdate(taskLogUpdate) {
      this.taskLog.addTaskLogUpdate(taskLogUpdate);
    },

    /**
     * Event handler for SocketIO "last-rendered" updates.
     * @param {API.SocketIOLastRenderedUpdate} lastRenderedUpdate
     */
    onSioLastRenderedUpdate(lastRenderedUpdate) {
      this.$refs.jobDetails.refreshLastRenderedImage(lastRenderedUpdate);
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
          // Forward the full task to Tabulator, so that that gets updated too.\
          if (this.$refs.tasksTable)
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

    _recalcTasksTableHeight() {
      if (!this.$refs.tasksTable) return;
      // Any recalculation should be done after the DOM has updated.
      this.$nextTick(this.$refs.tasksTable.recalcTableHeight);
    },
  },
}
</script>

<style scoped>
.isFetching {
  opacity: 50%;
}
</style>
