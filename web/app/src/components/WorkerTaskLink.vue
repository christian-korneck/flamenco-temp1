<template>
  <span>
    <router-link v-if="workerTask" :to="{ name: 'jobs', params: { jobID: workerTask.job_id, taskID: workerTask.id } }">
      {{ workerTask.name }}
      (<span class="status-label" :class="'status-' + workerTask.status">{{ workerTask.status }}</span>)
    </router-link>
    <span v-else>-</span>
  </span>
</template>

<script setup>
// 'workerTask' should be a WorkerTask object (see schema defined in `flamenco-openapi.yaml`).
const props = defineProps(['workerTask']);
</script>

<style scoped>
.status-label {
  color: var(--indicator-color);
  font-weight: bold;
}
</style>
