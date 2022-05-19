<template>
  <span class='socket-status' v-if="!sockStatus.isConnected" :title="sockStatus.message">Connection Lost</span>
</template>

<script setup>
import { useSocketStatus } from '@/stores/socket-status';

const sockStatus = useSocketStatus();
</script>

<style scoped>
.socket-status {
  font-size: 1rem;

  min-width: 1.2em;
  padding: 0.1em 1em;
  border-radius: 0.5em;
  text-align: center;
  margin: 0 0.5em;

  color: var(--color-connection-lost-text);
  background-color: var(--color-connection-lost-bg);
  box-shadow: 0 0 1em var(--color-connection-lost-bg);
  border: solid thin var(--color-connection-lost-text);

  cursor: help;
}
</style>
