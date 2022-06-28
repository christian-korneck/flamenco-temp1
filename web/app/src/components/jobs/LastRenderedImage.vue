<script setup>
import { ref, watch } from 'vue'
import { api } from '@/urls';
import { JobsApi } from '@/manager-api';
import { apiClient } from '@/stores/api-query-count';

const props = defineProps(['jobID']);
const imageURL = ref('');

const jobsApi = new JobsApi(apiClient);

function fetchImageURL(jobID) {
  jobsApi.fetchJobLastRenderedInfo(jobID)
    .then((info) => {
      // info is an api.JobLastRenderedImageInfo object.
      for (let suffix of info.suffixes) {
        if (!suffix.includes("-tiny")) continue;

        let url = new URL(api());
        url.pathname = info.base + "/" + suffix
        imageURL.value = url.toString();
        break;
      }
    })
    .catch((error) => { console.warn("error fetching last-rendered image info:", error) });
}

watch(() => props.jobID, (newJobID) => {
  console.log("Last-Rendered Image: new job ID: ", newJobID);
  fetchImageURL(newJobID);
});
fetchImageURL(props.jobID);

</script>

<template>
  <div v-if="imageURL != ''" class="lastRendered">
    <img :src="imageURL" alt="Last-rendered image for this job">
  </div>
</template>

<style scoped>
.lastRendered {
  width: 200px;
  height: 112px;
  float: right;
}
</style>
