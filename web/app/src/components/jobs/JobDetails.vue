<template>
  <template v-if="hasJobData">
    <div class="preview-container">
      <last-rendered-image ref="lastRenderedImage" :jobID="jobData.id" thumbnailSuffix="-small" />
    </div>

    <TabsWrapper @clicked-job-details-tab="emit_reshuffled_delayed">
      <TabItem title="Job Settings">
        <dl v-if="hasSettings">
          <template v-for="value, key in settingsToDisplay">
            <dt :class="`field-${key}`" :title="key">{{ key }}</dt>
            <dd>{{ value }}</dd>
          </template>
        </dl>
        <div v-else class="dl-no-data">
          <span>This job has no settings.</span>
        </div>
      </TabItem>
      <TabItem title="Metadata">
        <dl v-if="hasMetadata">
          <template v-for="value, key in jobData.metadata">
            <dt :class="`field-${key}`" :title="key">{{ key }}</dt>
            <dd>{{ value }}</dd>
          </template>
        </dl>
        <div v-else class="dl-no-data">
          <span>This job has no metadata.</span>
        </div>
      </TabItem>
      <TabItem title="Details">
        <dl>
          <dt class="field-name" title="ID">ID</dt>
          <dd>{{ jobData.id }}</dd>

          <dt class="field-name" title="Name">Name</dt>
          <dd>{{ jobData.name }}</dd>

          <dt class="field-status" title="Status">Status</dt>
          <dd class="field-status-label" :class="'status-' + jobData.status">{{ jobData.status }}</dd>

          <dt class="field-type" title="Type">Type</dt>
          <dd>{{ jobType ? jobType.label : jobData.type }}</dd>

          <dt class="field-priority" title="Priority">Priority</dt>
          <dd>{{ jobData.priority }}</dd>

          <dt class="field-created" title="Created">Created</dt>
          <dd>{{ datetime.relativeTime(jobData.created) }}</dd>

          <dt class="field-updated" title="Updated">Updated</dt>
          <dd>{{ datetime.relativeTime(jobData.updated) }}</dd>

          <dt class="field-activity" title="Activity">Activity</dt>
          <dd>{{ jobData.activity }}</dd>
        </dl>
      </TabItem>
      <TabItem title="Blocklist">
        <blocklist :jobID="jobData.id" @reshuffled="emit_reshuffled_delayed" />
      </TabItem>
    </TabsWrapper>
  </template>

  <div v-else class="details-no-item-selected">
    <p>Select a job to see its details.</p>
  </div>
</template>

<script>
import * as datetime from "@/datetime";
import * as API from '@/manager-api';
import { apiClient } from '@/stores/api-query-count';
import LastRenderedImage from '@/components/jobs/LastRenderedImage.vue'
import Blocklist from './Blocklist.vue'
import TabItem from '@/components/TabItem.vue'
import TabsWrapper from '@/components/TabsWrapper.vue'

export default {
  props: [
    "jobData", // Job data to show.
  ],
  emits: [
    "reshuffled", // Emitted when the size of this component may have changed. Used to resize other components in response.
  ],
  components: {
    LastRenderedImage,
    TabItem,
    TabsWrapper,
    Blocklist,
  },
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
    /**
     * @param {API.SocketIOLastRenderedUpdate} lastRenderedUpdate
     */
    refreshLastRenderedImage(lastRenderedUpdate) {
      this.$refs.lastRenderedImage.refreshLastRenderedImage(lastRenderedUpdate);
    },

    _refreshJobSettings(newJobData) {
      if (objectEmpty(newJobData)) {
        this._clearJobSettings();
        return;
      }

      // Only fetch the job type if it's different from what's already loaded.
      if (objectEmpty(this.jobType) || this.jobType.name != newJobData.type) {
        this._clearJobSettings(); // They should only be shown when the type info is known.

        this.jobsApi.getJobType(newJobData.type)
          .then(this.onJobTypeLoaded)
          .catch((error) => { console.warn("error fetching job type:", error) });
      } else {
        this._setJobSettings(newJobData.settings);
      }
    },

    onJobTypeLoaded(jobType) {
      this.jobType = jobType;

      // Construct a lookup table for the settings.
      const jobTypeSettings = {};
      for (let setting of jobType.settings)
        jobTypeSettings[setting.key] = setting;
      this.jobTypeSettings = jobTypeSettings;

      if (this.jobData) {
        this._setJobSettings(this.jobData.settings);
      }
      this.$emit('reshuffled');
    },

    _clearJobSettings() {
      this.simpleSettings = null;
      this.$emit('reshuffled');
    },

    _setJobSettings(newJobSettings) {
      if (objectEmpty(newJobSettings)) {
        this._clearJobSettings();
        return;
      }

      if (objectEmpty(this.jobTypeSettings)) {
        console.warn("empty job type settings");
        this._clearJobSettings();
        return;
      }

      // Construct a set of `setting.visible` values that should make the
      // setting visible here in the web interface.
      const v = new API.AvailableJobSettingVisibility();
      const visible = new Set([undefined, v.visible, v.web]);

      const filtered = {};
      for (let key in newJobSettings) {
        const setting = this.jobTypeSettings[key];
        if (typeof setting == 'undefined') {
          // Jobs can have settings beyond what the job type defines, for
          // example when the job is older than the latest change to a job type,
          // or when the submission system simply added custom settings.
          continue;
        }
        if (visible.has(setting.visible)) {
          filtered[key] = newJobSettings[key];
        }
      }

      this.simpleSettings = filtered;
      this.$emit('reshuffled');
    },
    emit_reshuffled_delayed() {
      const reshuffle = () => { this.$emit('reshuffled'); }

      // Changing tabs requires two sequential "reshuffled" events, at least it
      // does on Firefox. Not sure what the reason is, but it works to get rid
      // of the scrollbar.
      reshuffle();
      this.$nextTick(reshuffle);
    },
  },
};
</script>

<style scoped>
/* Prevent fields with long IDs from overflowing. */
.field-id+dd {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.field-status-label {
  color: var(--indicator-color);
  font-weight: bold;
}
</style>
