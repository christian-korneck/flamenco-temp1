<template>
  <router-link :to="{ name: 'workers', params: { workerID: worker.id } }">{{ worker.name }} ({{ worker.address }})</router-link>
</template>

<script setup>
// 'worker' should be a Worker or TaskWorker (see schemas defined in `flamenco-openapi.yaml`).
const props = defineProps(['worker']);
</script>
