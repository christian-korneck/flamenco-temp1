<template>
  <h2 class="column-title">Job Details</h2>
  <div v-if="hasJobData" class="job-details">
    <table class="details">
      <tr class="field-id">
        <th>ID</th>
        <td>{{ jobData.id }}</td>
      </tr>

      <tr class="field-name">
        <th>Name</th>
        <td>{{ jobData.name }}</td>
      </tr>

      <tr class="field-status">
        <th>Status</th>
        <td>{{ jobData.status }}</td>
      </tr>

      <tr class="field-type">
        <th>Type</th>
        <td>{{ jobType ? jobType.label : jobData.type }}</td>
      </tr>

      <tr class="field-priority">
        <th>Prio</th>
        <td>{{ jobData.priority }}</td>
      </tr>

      <tr class="field-created">
        <th>Created</th>
        <td>{{ datetime.relativeTime(jobData.created) }}</td>
      </tr>

      <tr class="field-updated">
        <th>Updated</th>
        <td>{{ datetime.relativeTime(jobData.updated) }}</td>
      </tr>

      <tr class="field-activity">
        <th>Activity</th>
        <td>{{ jobData.activity }}</td>
      </tr>
    </table>

    <h3 class="sub-title" v-if="hasMetadata">Meta-data</h3>
    <table class="metadata">
      <tr v-for="value, key in jobData.metadata" :class="`field-${key}`">
        <th>{{ key }}</th>
        <td>{{ value }}</td>
      </tr>
    </table>

    <h3 class="sub-title" v-if="hasSettings">Settings</h3>
    <table class="settings">
      <tr v-for="value, key in settingsToDisplay" :class="`field-${key}`">
        <th>{{ key }}</th>
        <td>{{ value }}</td>
      </tr>
    </table>

  </div>
  <div v-else class="no-job-selected">
    <p>No job selected, pick one from the list on the left.</p>
  </div>
</template>

<script lang="js">
import * as datetime from "@/datetime";
import * as API from '@/manager-api';
import { apiClient } from '@/stores/api-query-count';

function objectEmpty(o) {
  if (!o) return true;
  return Object.entries(o).length == 0;
}
window.objectEmpty = objectEmpty;

export default {
  props: [
    "jobData", // Job data to show.
  ],
  data() {
    return {
      datetime: datetime, // So that the template can access it.
      simpleSettings: null, // Object with filtered job settings, or null if there is no job.
      jobsApi: new API.JobsApi(apiClient),
      jobType: null, // API.AvailableJobType object for the current job type.
      jobTypeSettings: null, // Mapping from setting key to its definition in the job type.
      showAllSettings: false,
    };
  },
  mounted() {
    // Allow testing from the JS console:
    window.jobDetailsVue = this;

    if (!objectEmpty(this.jobData)) {
      this._refreshJobSettings(this.jobData);
    }
  },
  computed: {
    hasJobData() {
      return !!this.jobData && !!this.jobData.id;
    },
    hasMetadata() {
      return this.jobData && !objectEmpty(this.jobData.metadata);
    },
    hasSettings() {
      return this.jobData && !objectEmpty(this.settingsToDisplay);
    },
    settingsToDisplay() {
      if (!this.showAllSettings) {
        return this.simpleSettings;
      }
      if (objectEmpty(this.jobData) || objectEmpty(this.jobData.settings)) {
        return {};
      }
      return this.jobData.settings;
    },
  },
  watch: {
    jobData(newJobData) {
      this._refreshJobSettings(newJobData);
    },
  },
  methods: {
    _refreshJobSettings(newJobData) {
      if (objectEmpty(newJobData)) {
        this.simpleSettings = null;
        return;
      }

      // Only fetch the job type if it's different from what's already loaded.
      if (objectEmpty(this.jobType) || this.jobType.name != newJobData.type) {
        this.simpleSettings = null; // They should only be shown when the type info is known.

        this.jobsApi.getJobType(newJobData.type)
          .then(this.onJobTypeLoaded)
          .catch((error) => { console.warn("error fetching job type:", error) });
      } else {
        this._setJobSettings(newJobData.settings);
      }
    },

    onJobTypeLoaded(jobType) {
      console.log("Job type loaded: ", jobType);
      this.jobType = jobType;

      // Construct a lookup table for the settings.
      const jobTypeSettings = {};
      for (let setting of jobType.settings)
        jobTypeSettings[setting.key] = setting;
      this.jobTypeSettings = jobTypeSettings;

      if (this.jobData) {
        this._setJobSettings(this.jobData.settings);
      }
    },

    _setJobSettings(newJobSettings) {
      if (objectEmpty(newJobSettings)) {
        this.simpleSettings = null;
        return;
      }

      if (objectEmpty(this.jobTypeSettings)) {
        console.warn("empty job type settings");
        return;
      }

      const filtered = {};
      for (let key in newJobSettings) {
        const setting = this.jobTypeSettings[key];
        if (typeof setting == 'undefined') {
          // Jobs can have settings beyond what the job type defines, for
          // example when the job is older than the latest change to a job type,
          // or when the submission system simply added custom settings.
          continue;
        }
        if (setting.visible !== false) {
          filtered[key] = newJobSettings[key];
        }
      }

      this.simpleSettings = filtered;
    }
  }
};
</script>

<style scoped>
.job-details {
  font-size: smaller;
  font-family: 'Noto Mono', monospace;
}

tr:hover {
  background-color: #333333;
}

tr.field-id td {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

th {
  font-weight: bold;
  text-align: right;
  vertical-align: top;
}
</style>
