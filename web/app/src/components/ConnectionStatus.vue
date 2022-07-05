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
  border-radius: 999em;
  box-shadow: 0 0 1.33em var(--color-connection-lost-bg);
  color: var(--color-connection-lost-text);
  cursor: help;
  font-size: var(--font-size-sm);
  font-weight: bold;
  padding: var(--spacer-xs) var(--spacer);
  text-align: center;
}
</style>
