
<template>
  <div v-if="imageURL != ''" :class="cssClasses">
    <img :src="imageURL" alt="Last-rendered image for this job">
  </div>
</template>

<script setup>
import { reactive, ref, watch } from 'vue'
import { api } from '@/urls';
import { JobsApi, JobLastRenderedImageInfo, SocketIOLastRenderedUpdate } from '@/manager-api';
import { apiClient } from '@/stores/api-query-count';

const props = defineProps([
  /* The job UUID to show renders for, or some false-y value if renders from all
   * jobs should be accepted. */
  'jobID',
  /* Name of the thumbnail, or subset thereof. See `JobLastRenderedImageInfo` in
   * `flamenco-openapi.yaml`, and * `internal/manager/last_rendered/last_rendered.go`.
   * The component picks the 'suffix' that has the given `thumbnailSuffix` as
   * substring. */
  'thumbnailSuffix',
]);
const imageURL = ref('');
const cssClasses = reactive({
  'last-rendered': true,
  'nothing-rendered-yet': true,
})

const jobsApi = new JobsApi(apiClient);

/**
 * Fetches the last-rendered info for the given job, then updates the <img> tag for it.
 */
function fetchImageURL(jobID) {
  let promise;
  if (jobID)
    promise = jobsApi.fetchJobLastRenderedInfo(jobID);
  else
    promise = jobsApi.fetchGlobalLastRenderedInfo();

  promise
    .then(setImageURL)
    .catch((error) => { console.warn("error fetching last-rendered image info:", error) });
}

/**
 * @param {JobLastRenderedImageInfo} thumbnailInfo
 */
function setImageURL(thumbnailInfo) {
  console.log("LastRenderedImage.vue: setImageURL", thumbnailInfo);

  if (thumbnailInfo == null) {
    // This indicates that there is no last-rendered image.
    // Default to a hard-coded 'nothing to be seen here, move along' image.
    imageURL.value = "/app/v3/nothing-rendered-yet.svg";
    cssClasses['nothing-rendered-yet'] = true;
    console.log("LastRenderedImage.vue: setting image URL to:", imageURL.value);
    return;
  }

  // Set the image URL to something appropriate.
  let foundThumbnail = false;
  const suffixToFind = props.thumbnailSuffix;
  for (let suffix of thumbnailInfo.suffixes) {
    if (!suffix.includes(suffixToFind)) continue;

    // This uses the API URL to construct the image URL, as the image comes from
    // Flamenco Manager, and not from any development server that might be
    // serving the webapp.
    let url = new URL(api());
    url.pathname = thumbnailInfo.base + "/" + suffix
    url.search = new Date().getTime(); // This forces the image to be reloaded.
    imageURL.value = url.toString();
    console.log("LastRenderedImage.vue: setting image URL to:", imageURL.value);
    foundThumbnail = true;
    break;
  }
  if (!foundThumbnail) {
    console.warn(`LastRenderedImage.vue: could not find thumbnail with suffix "${suffixToFind}"; available are:`, thumbnailInfo.suffixes);
  }
  cssClasses['nothing-rendered-yet'] = !foundThumbnail;
}

/**
 * @param {SocketIOLastRenderedUpdate} lastRenderedUpdate
 */
function refreshLastRenderedImage(lastRenderedUpdate) {
  // Only filter out other job IDs if this component has actually a non-empty job ID.
  if (props.jobID && lastRenderedUpdate.job_id != props.jobID) {
    console.log(
      "LastRenderedImage.vue: refreshLastRenderedImage() received update for job",
      lastRenderedUpdate.job_id,
      "but this component is showing job", props.jobID);
    return;
  }

  console.log('refreshLastRenderedImage:', lastRenderedUpdate);
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

<style scoped>
.last-rendered.nothing-rendered-yet {
  outline: thin dotted var(--color-text-hint);
}

.last-rendered img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
</style>
