<template>
  <span class='notifications' v-if="notifs.last">{{ notifs.last.msg }}</span>
</template>

<script>
import { useNotifs } from '@/stores/notifications';

export default {
  name: 'NotificationBar',
  data: () => ({
    notifs: useNotifs(),
  }),
}
</script>
