<template>
  <div v-if="isFetching" class="dl-no-data">
    <span>Fetching blocklist...</span>
  </div>
  <template v-else>
    <table class="blocklist" v-if="blocklist.length">
      <tr>
        <th>Worker</th>
        <th>Task Type</th>
      </tr>
      <tr v-for="entry in blocklist">
        <td>
          <link-worker :worker="{ id: entry.worker_id, name: entry.worker_name }" />
        </td>
        <td>{{ entry.task_type }}</td>
      </tr>
    </table>
    <div v-else class="dl-no-data">
      <span>This job has no blocked workers.</span>
    </div>
  </template>
  <p v-if="errorMsg" class="error">Error fetching blocklist: {{ errorMsg }}</p>
</template>

<script setup>
import { apiClient } from '@/stores/api-query-count';
import { JobsApi } from '@/manager-api';
import LinkWorker from '@/components/LinkWorker.vue';
import { watch, onMounted, inject, ref } from 'vue'

// jobID should be the job UUID string.
const props = defineProps(['jobID']);
const jobsApi = new JobsApi(apiClient);
const isVisible = inject("isVisible");
const isFetching = ref(false);
const errorMsg = ref("");
const blocklist = ref([]);

function refreshBlocklist() {
  if (!isVisible.value) {
    return;
  }

  isFetching.value = true;
  jobsApi.fetchJobBlocklist(props.jobID)
    .then((newBlocklist) => {
      console.log(`received blocklist for job ${props.jobID}`, newBlocklist);
      blocklist.value = newBlocklist;
      // TODO: remit 'reshuffled' & handle in parent.
    })
    .catch((error) => {
      errorMsg.value = error.message;
    })
    .finally(() => {
      isFetching.value = false;
    })
}

watch(() => props.jobID, refreshBlocklist);
watch(isVisible, refreshBlocklist);
onMounted(refreshBlocklist);
</script>

<style scoped>
table.blocklist {
  width: 100%;
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
  border-collapse: collapse;
}

table.blocklist td,
table.blocklist th {
  text-align: left;
  padding: calc(var(--spacer-sm)/2) var(--spacer-sm);
}

table.blocklist th {
  color: var(--color-text-muted);
  font-weight: normal;
}

table.blocklist tr {
  background-color: var(--table-color-background-row);
}

table.blocklist tr:nth-child(odd) {
  background-color: var(--table-color-background-row-odd);
}
</style>
