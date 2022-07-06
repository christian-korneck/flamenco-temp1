<template>
  <span class='socket-status' v-if="!sockStatus.isConnected" :title="sockStatus.message">Connection Lost</span>
</template>

<script setup>
import { useSocketStatus } from '@/stores/socket-status';

const sockStatus = useSocketStatus();
</script>

<style scoped>
.socket-status {
  background-color: var(--color-connection-lost-bg);
  border-radius: var(--border-radius);
  color: var(--color-connection-lost-text);
  cursor: help;
  font-size: var(--font-size-sm);
  padding: 2px var(--spacer-sm);
  text-align: center;
}
</style>
