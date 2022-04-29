<template>
  <header>{{ flamencoName }}</header>
  <header class="right">
    <api-spinner />
    version: {{ flamencoVersion }}
  </header>
  <div class="col-1">
    <jobs-table ref="jobsTable" @selectedJobChange="onSelectedJobChanged" />
  </div>
  <div class="col-2">
    <job-details :jobData="jobs.activeJob" />
    <tasks-table ref="tasksTable" :jobID="jobs.activeJobID" @selectedTaskChange="onSelectedTaskChanged" />
  </div>
  <div class="col-3">
    <task-details />
  </div>
  <footer>
    <span class='notifications' v-if="notifs.last">{{ notifs.last.msg }}</span>
    <update-listener ref="updateListener" :websocketURL="websocketURL" @jobUpdate="onSioJobUpdate"
      @message="onChatMessage" @sioReconnected="onSIOReconnected" @sioDisconnected="onSIODisconnected" />
  </footer>
</template>

<script>
import * as urls from '@/urls'
import * as API from '@/manager-api';
import { useJobs } from '@/stores/jobs';
import { useTasks } from '@/stores/tasks';
import { useNotifs } from '@/stores/notifications';
import { apiClient } from '@/stores/api-query-count';

import ApiSpinner from '@/components/ApiSpinner.vue'
import JobsTable from '@/components/JobsTable.vue'
import JobDetails from '@/components/JobDetails.vue'
import TaskDetails from '@/components/TaskDetails.vue'
import TasksTable from '@/components/TasksTable.vue'
import UpdateListener from '@/components/UpdateListener.vue'

const DEFAULT_FLAMENCO_NAME = "Flamenco";
const DEFAULT_FLAMENCO_VERSION = "unknown";

export default {
  name: 'App',
  components: {
    ApiSpinner, JobsTable, JobDetails, TaskDetails, TasksTable, UpdateListener,
  },
  data: () => ({
    websocketURL: urls.ws(),
    messages: [],

    jobs: useJobs(),
    tasks: useTasks(),
    notifs: useNotifs(),

    flamencoName: DEFAULT_FLAMENCO_NAME,
    flamencoVersion: DEFAULT_FLAMENCO_VERSION,
  }),
  mounted() {
    window.app = this;
    this.fetchManagerInfo();
  },
  methods: {
    // UI component event handlers:
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
      // const jobsAPI = new API.JobsApi(apiClient);
      // jobsAPI.fetchTask(taskSummary.id)
      //   .then((task) => {
      //     this.tasks.setSelectedTask(task);
      //     // Forward the full task to Tabulator, so that that gets updated too.
      //     this.$refs.tasksTable.processTaskUpdate(task);
      //   });
    },

    sendMessage(message) {
      this.$refs.jobsListener.sendBroadcastMessage("typer", message);
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
      const activeJob = this.jobs.activeJob;
      if (activeJob && activeJob.id == jobUpdate.id) {
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
    onChatMessage(message) {
      console.log("chat message received:", message);
      this.messages.push(`${message.text}`);
    },

    // SocketIO connection event handlers:
    onSIOReconnected() {
      this.$refs.jobsTable.onReconnected();
      this.$refs.tasksTable.onReconnected();
      this.fetchManagerInfo();
    },
    onSIODisconnected(reason) {
      this.flamencoName = DEFAULT_FLAMENCO_NAME;
      this.flamencoVersion = DEFAULT_FLAMENCO_VERSION;
    },
    fetchManagerInfo() {
      const metaAPI = new API.MetaApi(apiClient);
      metaAPI.getVersion().then((version) => {
        this.flamencoName = version.name;
        this.flamencoVersion = version.version;
      })
    },
  },
}
</script>

<style>
@import "tabulator-tables/dist/css/tabulator_midnight.min.css";

:root {
  --header-height: 25px;
  --footer-height: 25px;
  --grid-gap: 4px;

  --action-bar-height: 2em;

  --blender-blue: #265787;
}

html,
body {
  height: calc(100vh - 1px);
  margin: 0;
  padding: 0;

  color: #EEEEEE;
  background-color: #222222;
}

#app {
  font-family: 'Noto Sans', sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;

  display: grid;
  width: 100%;
  height: 100%;
  grid-template-columns: 1fr 1fr 1fr;
  grid-template-rows: var(--header-height) 1fr var(--footer-height);
  grid-gap: var(--grid-gap);
  grid-template-areas:
    "header header header-right"
    "col-1 col-2 col-3"
    "footer footer footer";
}

header,
header.right {
  background-color: #333;
  color: #EEE;
}

header {
  grid-area: header;
  padding-left: 0.2rem;
}

header.right {
  grid-area: header-right;
  text-align: right;
  font-size: smaller;
  padding-right: 0.4rem;
}

h2.column-title {
  margin-top: 0;
  font-size: 12pt;
  border-bottom: 1px solid grey
}

h3.sub-title {
  margin-bottom: 0;
  font-size: 10pt;
  border-bottom: 1px solid grey
}

.col-1 {
  grid-area: col-1;
  max-height: calc(100vh - var(--header-height) - var(--footer-height) - 2*var(--grid-gap));
}

.col-2 {
  grid-area: col-2;
}

.col-3 {
  grid-area: col-3;
}

footer {
  grid-area: footer;
  background-color: var(--blender-blue);
  color: #EEE;
  padding-top: 0.2rem;
  padding-left: 0.2rem;
}

section.action-bar {
  height: var(--action-bar-height);
  display: flex;
  flex-direction: row;
  align-items: center;
}

section.action-bar button.action {
  padding: 0.1rem 0.75rem;
  border-radius: 0.3rem;
  border: thin solid white;

  background: var(--blender-blue);
  color: #DFDEDF;
  touch-action: manipulation;
  margin-right: 0.3rem;

  transition-duration: 300ms;
  transition-property: color, background-color, border-color, box-shadow;
}

section.action-bar button.action[disabled] {
  background-color: #4c4b4d;
  color: #858585;
  border: thin solid #858585;
}

section.action-bar button.action:hover:not([disabled]) {
  transition: all 100ms;
  box-shadow: inset 0 0 0.3rem rgba(255, 255, 255, 0.8);
}

section.action-bar button.action:focus {
  /* Make sure the outline is clearly visible inside the button. */
  outline-offset: -0.5em;
}

section.action-bar button.action.dangerous {
  background-color: #7c4d41;
  color: #e4c5c0;
  margin-left: 1rem;
}

section.action-bar button.action.dangerous[disabled] {
  background-color: #53413e;
}
</style>
