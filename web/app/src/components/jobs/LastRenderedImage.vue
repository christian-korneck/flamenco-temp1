<script setup>
import { reactive, ref, watch } from 'vue'
import { api } from '@/urls';
import { JobsApi, JobLastRenderedImageInfo, SocketIOLastRenderedUpdate } from '@/manager-api';
import { apiClient } from '@/stores/api-query-count';

const props = defineProps(['jobID']);
const imageURL = ref('');
const cssClasses = reactive({
  lastRendered: true,
  nothingRenderedYet: true,
})

const jobsApi = new JobsApi(apiClient);

/**
 * Fetches the last-rendered info for the given job, then updates the <img> tag for it.
 */
function fetchImageURL(jobID) {
  jobsApi.fetchJobLastRenderedInfo(jobID)
    .then(setImageURL)
    .catch((error) => { console.warn("error fetching last-rendered image info:", error) });
}

/**
 * @param {JobLastRenderedImageInfo} thumbnailInfo
 */
function setImageURL(thumbnailInfo) {
  if (thumbnailInfo == null) {
    // This indicates that there is no last-rendered image.
    // Default to a hard-coded 'nothing to be seen here, move along' image.
    imageURL.value = "/nothing-rendered-yet.svg";
    cssClasses.nothingRenderedYet = true;
    return;
  }

  // Set the image URL to something appropriate.
  for (let suffix of thumbnailInfo.suffixes) {
    if (!suffix.includes("-tiny")) continue;

    let url = new URL(api());
    url.pathname = thumbnailInfo.base + "/" + suffix
    url.search = new Date().getTime(); // This forces the image to be reloaded.
    imageURL.value = url.toString();
    break;
  }
  cssClasses.nothingRenderedYet = false;
}

/**
 * @param {SocketIOLastRenderedUpdate} lastRenderedUpdate
 */
function refreshLastRenderedImage(lastRenderedUpdate) {
  if (lastRenderedUpdate.job_id != props.jobID) {
    console.log(
      "LastRenderedImage.vue: refreshLastRenderedImage() received update for job",
      lastRenderedUpdate.job_id,
      "but this component is showing job", props.jobID);
    return;
  }

  setImageURL(lastRenderedUpdate.thumbnail);
}

// Call fetchImageURL(jobID) whenever the job ID prop changes value.
watch(() => props.jobID, (newJobID) => {
  console.log("Last-Rendered Image: new job ID: ", newJobID);
  fetchImageURL(newJobID);
});
fetchImageURL(props.jobID);

// Expose refreshLastRenderedImage() so that it can be called from the parent
// component in response to SocketIO messages.
defineExpose({
  refreshLastRenderedImage,
});
</script>

<template>
  <div v-if="imageURL != ''" :class="cssClasses">
    <img :src="imageURL" alt="Last-rendered image for this job">
  </div>
</template>

<style scoped>
.lastRendered {
  width: 200px;
  height: 112px;
  float: right;
}

.lastRendered.nothingRenderedYet {
  outline: thin dotted var(--color-text-hint);
}

.lastRendered img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

</style>
