<template>
  <div class="col col-1">
    Workers
  </div>
  <div class="col col-2">
    Worker Details
  </div>
  <div class="col col-3">
    Column 3
  </div>
  <footer>
    <notification-bar />
    <update-listener ref="updateListener" :websocketURL="websocketURL" @sioReconnected="onSIOReconnected"
      @sioDisconnected="onSIODisconnected" />
  </footer>
</template>

<script>
import * as urls from '@/urls'

import NotificationBar from '@/components/footer/NotificationBar.vue'
import UpdateListener from '@/components/UpdateListener.vue'

export default {
  name: 'WorkersView',
  components: {
    NotificationBar, UpdateListener,
  },
  data: () => ({
    websocketURL: urls.ws(),
  }),
  mounted() {
    window.workersView = this;
  },
  methods: {
    // SocketIO connection event handlers:
    onSIOReconnected() {
    },
    onSIODisconnected(reason) {
    },
  },
}
</script>
