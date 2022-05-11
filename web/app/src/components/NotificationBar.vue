<template>
  <section class="notification-bar">
    <span class='notifications' v-if="notifs.last">{{ notifs.last.msg }}</span>
    <span class='socket-status' v-if="!sockStatus.isConnected" :title="sockStatus.message">Connection Lost</span>
  </section>
</template>

<script>
import { useNotifs } from '@/stores/notifications';
import { useSocketStatus } from '@/stores/socket-status';

export default {
  name: 'NotificationBar',
  data: () => ({
    notifs: useNotifs(),
    sockStatus: useSocketStatus(),
  }),
}
</script>

<style scoped>
.notification-bar {
  width: 100%;
}

.socket-status {
  float: right;

  font-weight: bold;
  font-size: larger;

  min-width: 1.2em;
  padding: 0.2em 1em;
  border-radius: 0.5em;
  text-align: center;

  color: var(--color-connection-lost-text);
  background-color: var(--color-connection-lost-bg);
  box-shadow: 0 0 1em var(--color-connection-lost-bg);
  border: solid thin var(--color-connection-lost-text);

  cursor: help;
}
</style>
