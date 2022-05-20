import { defineStore } from 'pinia'
import { useNotifs } from '@/stores/notifications'

// Not sure if this is the best way to deal with those notifications. It feels a
// bit spaghetto to have one Pinia store influence another. Maybe move this to
// the app level once the Workers and Settings views are fleshed out. Maybe
// that'll cause the Notifications popover to be handled at the app-global
// level, instead of per view, creating a better place to put this logic.
const notifs = useNotifs();

/**
 * Status of the SocketIO/Websocket connection to Flamenco Manager.
 */
export const useSocketStatus = defineStore('socket-status', {
  state: () => ({
    /** @type { bool } */
    isConnected: false,

    wasEverDisconnected: false,

    /** @type {string} */
    message: "",
  }),
  actions: {
    /**
     * Indicate the connection was lost.
     * @param {string} reason
     */
    disconnected(reason) {
      // Only patch the state if it actually will change.
      if (!this.isConnected)
        return;
      notifs.add(`Connection to Flamenco Manager lost`);
      this.$patch({
        isConnected: false,
        wasEverDisconnected: true,
        message: `${reason}`,
      });
    },
    /**
     * Indicate the connection is good.
     */
    connected() {
      // Only patch the state if it actually will change.
      if (this.isConnected)
        return;

      if (this.wasEverDisconnected)
        notifs.add("Connection to Flamenco Manager established");
      this.$patch({
        isConnected: true,
        message: "",
      });
    },
  }
})
