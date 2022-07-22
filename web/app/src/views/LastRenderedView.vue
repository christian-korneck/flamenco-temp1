<template>
  <div class="global-last-rendered">
    <last-rendered-image
      ref="lastRenderedImage"
      :jobID="false"
      thumbnailSuffix="last-rendered.jpg" />
  </div>

  <footer class="app-footer"><notification-bar /></footer>

  <update-listener ref="updateListener" mainSubscription="allLastRendered"
    @lastRenderedUpdate="onSioLastRenderedUpdate"
    @sioReconnected="onSIOReconnected" @sioDisconnected="onSIODisconnected" />
</template>

<script>
import LastRenderedImage from '@/components/jobs/LastRenderedImage.vue'
import NotificationBar from '@/components/footer/NotificationBar.vue'
import UpdateListener from '@/components/UpdateListener.vue'

export default {
  name: 'LastRenderedView',
  components: {
    LastRenderedImage,
    NotificationBar,
    UpdateListener,
  },
  data: () => ({
  }),
  methods: {
    /**
     * Event handler for SocketIO "last-rendered" updates.
     * @param {API.SocketIOLastRenderedUpdate} lastRenderedUpdate
     */
    onSioLastRenderedUpdate(lastRenderedUpdate) {
      this.$refs.lastRenderedImage.refreshLastRenderedImage(lastRenderedUpdate);
    },

    // SocketIO connection event handlers:
    onSIOReconnected() {
    },
    onSIODisconnected(reason) {
    },
  },
}
</script>

<style scoped>
.global-last-rendered {
  align-items: center;
  display: flex;
  grid-column-end: col-3;
  grid-column-start: col-1;
  justify-content: center;
  height: calc(100vh - calc(var(--header-height) - var(--footer-height) - calc(var(--grid-gap) * -12)));
  max-height: 100%;
}

.global-last-rendered .last-rendered {
  max-height: 100%;
  max-width: 100%;
}
</style>
