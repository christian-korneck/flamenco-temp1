<template>
  <div class="global-last-rendered">
      <last-rendered-image
        ref="lastRenderedImage"
        :jobID="false"
        thumbnailSuffix="last-rendered.jpg" />
  </div>

  <footer class="window-footer"><notification-bar /></footer>

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
.last-rendered {
  max-height: 100%;
}

.global-last-rendered {
  grid-column-start: col-1;
  grid-column-end: col-3;

  /* FIXME: the positioning of the image & sizing of the container DIV doesn't
  work well yet. */
  height: 100%;
}

</style>
