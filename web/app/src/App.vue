<template>
  <header>{{ flamencoName }}<span class='flamenco-version'>version: {{ flamencoVersion }}</span></header>
  <div class="col-1">
    <jobs-table ref="jobsTable" :apiClient="apiClient" @activeJobChange="onActiveJobChanged" />
  </div>
  <div class="col-2">
    <job-details :apiClient="apiClient" :jobSummary="activeJobSummary" />
  </div>
  <div class="col-3">
    <task-details :apiClient="apiClient" />
  </div>
  <footer>Footer
    <update-listener ref="updateListener" :websocketURL="websocketURL" @jobUpdate="onJobUpdate" @message="onChatMessage"
      @sioReconnected="onSIOReconnected" @sioDisconnected="onSIODisconnected" />
  </footer>
</template>

<script>
import * as urls from './urls'
import * as API from './manager-api';
import JobsTable from './components/JobsTable.vue'
import JobDetails from './components/JobDetails.vue'
import TaskDetails from './components/TaskDetails.vue'
import UpdateListener from './components/UpdateListener.vue'

const DEFAULT_FLAMENCO_NAME = "Flamenco";
const DEFAULT_FLAMENCO_VERSION = "unknown";

export default {
  name: 'App',
  components: {
    JobsTable, JobDetails, TaskDetails, UpdateListener,
  },
  data: () => {
    return {
      apiClient: new API.ApiClient(urls.api()),
      websocketURL: urls.ws(),
      messages: [],

      activeJobSummary: {},
      flamencoName: DEFAULT_FLAMENCO_NAME,
      flamencoVersion: DEFAULT_FLAMENCO_VERSION,
    };
  },
  mounted() {
    this.fetchManagerInfo();
  },
  methods: {
    // UI component event handlers:
    onActiveJobChanged(jobSummary) {
      this.activeJobSummary = jobSummary;
    },

    // SocketIO data event handlers:
    sendMessage(message) {
      this.$refs.jobsListener.sendBroadcastMessage("typer", message);
    },
    onJobUpdate(jobUpdate) {
      console.log("job update received:", jobUpdate);
      if (jobUpdate.previous_status) {
        this.messages.push(`Job update: ${jobUpdate.id} (${jobUpdate.previous_status} → ${jobUpdate.status})`);
        this.$refs.jobsTable.processJobUpdate(jobUpdate);
      } else {
        this.messages.push(`New job: ${jobUpdate.id} (${jobUpdate.status})`);
        this.$refs.jobsTable.processNewJob(jobUpdate);
      }
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
    }
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
