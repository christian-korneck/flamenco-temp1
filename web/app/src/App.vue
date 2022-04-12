<template>
  <header>{{ flamencoName }}
    <api-spinner :numRunningQueries="numRunningQueries" />
    <span class='flamenco-version'>version: {{ flamencoVersion }}</span>
  </header>
  <div class="col-1">
    <jobs-table ref="jobsTable" :apiClient="apiClient" @selectedJobChange="onSelectedJobChanged" />
  </div>
  <div class="col-2">
    <job-details :apiClient="apiClient" :jobData="selectedJob" />
  </div>
  <div class="col-3">
    <task-details :apiClient="apiClient" />
  </div>
  <footer>Footer
    <update-listener ref="updateListener" :websocketURL="websocketURL" @jobUpdate="onSioJobUpdate"
      @message="onChatMessage" @sioReconnected="onSIOReconnected" @sioDisconnected="onSIODisconnected" />
  </footer>
</template>

<script>
import * as urls from './urls'
import * as API from './manager-api';
import ApiSpinner from './components/ApiSpinner.vue'
import JobsTable from './components/JobsTable.vue'
import JobDetails from './components/JobDetails.vue'
import TaskDetails from './components/TaskDetails.vue'
import UpdateListener from './components/UpdateListener.vue'

const DEFAULT_FLAMENCO_NAME = "Flamenco";
const DEFAULT_FLAMENCO_VERSION = "unknown";

export default {
  name: 'App',
  components: {
    ApiSpinner, JobsTable, JobDetails, TaskDetails, UpdateListener,
  },
  data: () => {
    return {
      apiClient: new API.ApiClient(urls.api()),
      websocketURL: urls.ws(),
      messages: [],

      selectedJob: {},
      flamencoName: DEFAULT_FLAMENCO_NAME,
      flamencoVersion: DEFAULT_FLAMENCO_VERSION,

      numRunningQueries: 0,
    };
  },
  mounted() {
    this.fetchManagerInfo();
  },
  methods: {
    // UI component event handlers:
    onSelectedJobChanged(jobSummary) {
      const jobsAPI = new API.JobsApi(this.apiClient);
      this._wrap(jobsAPI.fetchJob(jobSummary.id))
        .then((job) => {
          this.selectedJob = job;
        });
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
      if (this.selectedJob && this.selectedJob.id == jobUpdate.id) {
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
      this.fetchManagerInfo();
    },
    onSIODisconnected(reason) {
      this.flamencoName = DEFAULT_FLAMENCO_NAME;
      this.flamencoVersion = DEFAULT_FLAMENCO_VERSION;
    },
    fetchManagerInfo() {
      const metaAPI = new API.MetaApi(this.apiClient);
      metaAPI.getVersion().then((version) => {
        this.flamencoName = version.name;
        this.flamencoVersion = version.version;
      })
    },

    // Wrap a Flamenco API promise, to keep track of how many queries are running.
    // This is just a test to see how this works, not a final functional design.
    _wrap(promise) {
      this.numRunningQueries++;
      return promise.finally(() => {
        this.numRunningQueries--;
      });
    },
  },
}
</script>

<style>
@import "tabulator-tables/dist/css/tabulator_midnight.min.css";

body {
  margin: 0;
}

#app {
  font-family: 'Noto Sans', sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #EEEEEE;
  background-color: #222222;

  display: grid;
  width: 100%;
  height: 100vh;
  grid-template-columns: 1fr 1fr 1fr;
  grid-template-rows: 25px 1fr 25px;
  grid-template-areas:
    "header header header"
    "col-1 col-2 col-3"
    "footer footer footer";
}

header {
  grid-area: header;
  background-color: #333;
  color: #EEE;
}

header span.flamenco-version {
  float: right;
  font-size: small;
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
}

.col-2 {
  grid-area: col-2;
}

.col-3 {
  grid-area: col-3;
}

footer {
  grid-area: footer;
  background-color: #333333;
  color: #EEE;
}
</style>
