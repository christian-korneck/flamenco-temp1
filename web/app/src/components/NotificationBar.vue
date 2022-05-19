<template>
  <section class="notification-bar">
    <span class='notifications' v-if="notifs.last">{{ notifs.last.msg }}</span>
    <connection-status />
  </section>
</template>

<script>
import { useNotifs } from '@/stores/notifications';
import ConnectionStatus from '@/components/ConnectionStatus.vue'

export default {
  name: 'NotificationBar',
  components: {
    ConnectionStatus,
  },
  data: () => ({
    notifs: useNotifs(),
  }),
}
</script>

<style scoped>
.notification-bar {
  width: 100%;
}

.socket-status {
  float: right;
}
</style>
