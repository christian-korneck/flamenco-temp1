<template>
  <h2 class="column-title">Job Details</h2>

  <template v-if="hasJobData">
    <dl>
      <dt class="field-id">ID</dt>
      <dd :title="jobData.id">{{ jobData.id }}</dd>

      <dt class="field-name">Name</dt>
      <dd>{{ jobData.name }}</dd>

      <dt class="field-status">Status</dt>
      <dd>{{ jobData.status }}</dd>

      <dt class="field-type">Type</dt>
      <dd>{{ jobType ? jobType.label : jobData.type }}</dd>

      <dt class="field-priority">Priority</dt>
      <dd>{{ jobData.priority }}</dd>

      <dt class="field-created">Created</dt>
      <dd>{{ datetime.relativeTime(jobData.created) }}</dd>

      <dt class="field-updated">Updated</dt>
      <dd>{{ datetime.relativeTime(jobData.updated) }}</dd>

      <dt class="field-activity">Activity</dt>
      <dd>{{ jobData.activity }}</dd>
    </dl>

    <h3 class="sub-title" v-if="hasMetadata">Meta-data</h3>
    <dl>
      <template v-for="value, key in jobData.metadata">
        <dt :class="`field-${key}`" :title="key">{{ key }}</dt>
        <dd>{{ value }}</dd>
      </template>
    </dl>

    <h3 class="sub-title" v-if="hasSettings">Settings</h3>
    <dl>
      <template v-for="value, key in settingsToDisplay">
        <dt :class="`field-${key}`" :title="key">{{ key }}</dt>
        <dd>{{ value }}</dd>
      </template>
    </dl>
  </template>

  <div v-else class="details-no-item-selected">
    <p>Select a job to see its details.</p>
  </div>
</template>

<script lang="js">
import * as datetime from "@/datetime";
import * as API from '@/manager-api';
import { apiClient } from '@/stores/api-query-count';

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
/* Prevent fields with long IDs from overflowing. */
.field-id + dd {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
