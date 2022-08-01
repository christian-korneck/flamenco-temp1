<template>
  <router-link :to="{ name: 'workers', params: { workerID: worker.id } }">
    {{ worker.name }}{{ workerAddress }}
  </router-link>
</template>

<script setup>
import { computed } from 'vue'

// 'worker' should be a Worker or TaskWorker (see schemas defined in `flamenco-openapi.yaml`).
const props = defineProps(['worker']);
const workerAddress = computed(() => {
  if (props.worker.address) return `(${props.worker.address})`;
  return "";
});
</script>
