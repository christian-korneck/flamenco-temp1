<template>
  <div class="col col-1">
    <jobs-table ref="jobsTable" @selectedJobChange="onSelectedJobChanged" />
  </div>
  <div class="col col-2">
    <job-details :jobData="jobs.activeJob" />
    <tasks-table v-if="jobs.activeJobID" ref="tasksTable" :jobID="jobs.activeJobID"
      @selectedTaskChange="onSelectedTaskChanged" />
  </div>
  <div class="col col-3">
    <task-details :taskData="tasks.activeTask" />
  </div>
  <footer>
    <notification-bar />
    <update-listener ref="updateListener" :websocketURL="websocketURL" :subscribedJob="jobs.activeJobID"
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
  },
  methods: {
    // onSelectedJobChanged is called whenever the selected job changes; this is
    // both when another job is selected and when the selected job itself gets
    // updated.
    onSelectedJobChanged(jobSummary) {
      if (!jobSummary) { // There is no selected job.
        this.jobs.deselectAllJobs();
        return;
      }

      const jobsAPI = new API.JobsApi(apiClient);
      jobsAPI.fetchJob(jobSummary.id)
        .then((job) => {
          this.jobs.setSelectedJob(job);
          // Forward the full job to Tabulator, so that that gets updated too.
          this.$refs.jobsTable.processJobUpdate(job);
        });
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
      if (!jobUpdate.previous_status)
        return this.onJobNew(jobUpdate);
      return this.onJobUpdate(jobUpdate);
    },
    onJobUpdate(jobUpdate) {
      // this.messages.push(`Job update: ${jobUpdate.id} (${jobUpdate.previous_status} → ${jobUpdate.status})`);
      if (this.$refs.jobsTable) {
        this.$refs.jobsTable.processJobUpdate(jobUpdate);
      } else {
        console.warn("App: this.$refs.jobsTable is", this.$refs.jobsTable);
      }

      if (this.jobs.activeJobID == jobUpdate.id) {
        this.onSelectedJobChanged(jobUpdate);
      }
    },
    onJobNew(jobUpdate) {
      if (!this.$refs.jobsTable) {
        console.warn("App: this.$refs.jobsTable is", this.$refs.jobsTable);
        return;
      }

      // this.messages.push(`New job: ${jobUpdate.id} (${jobUpdate.status})`);
      this.$refs.jobsTable.processNewJob(jobUpdate);
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
      this.jobs.deselectAllJobs();
    },
  },
}
</script>
