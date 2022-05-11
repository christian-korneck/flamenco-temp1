import { stringifyStyle } from '@vue/shared'
import { defineStore } from 'pinia'

/**
 * Status of the SocketIO/Websocket connection to Flamenco Manager.
 */
export const useSocketStatus = defineStore('socket-status', {
  state: () => ({
    /** @type { bool } */
    isConnected: false,

    /** @type {string} */
    message: "",
  }),
  actions: {
    /**
     * Indicate the connection was lost.
     * @param {string} reason
     */
     disconnected(reason) {
      this.$patch({
        isConnected: false,
        message: `${reason}`,
      });
    },
    /**
     * Indicate the connection is good.
     */
     connected() {
      this.$patch({
        isConnected: true,
        message: "",
      });
    },
  }
})
