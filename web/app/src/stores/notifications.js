import { defineStore } from 'pinia'

/** Time after which a notification is hidden. */
const MESSAGE_HIDE_DELAY_MS = 5000;

/**
 * Store notifications to users of the web frontend.
 */
export const useNotifs = defineStore('notifications', {
  state: () => ({
    /** @type {{ msg: string, time: Date }[]} */
    history: [],
    /** @type { msg: string, time: Date } */
    last: "",

    hideTimerID: 0,
  }),
  actions: {
    /**
     * Add a simple notification.
     * @param {string} message
     */
    add(message) {
      const notif = {msg: message, time: new Date()};
      this.history.push(notif);
      this.last = notif;
      this._prune();
      this._restartHideTimer();
    },

    /* Ensure there is only 1000 items in the history. */
    _prune() {
      if (this.history.length <= 1000) return;
      this.history.splice(0, 1000 - this.history.length);
    },

    _restartHideTimer() {
      if (this.hideTimerID) window.clearTimeout(this.hideTimerID);
      this.hideTimerID = window.setTimeout(this._hideMessage, MESSAGE_HIDE_DELAY_MS);
    },
    _hideMessage() {
      this.$patch({
        hideTimerID: 0,
        last: "",
      });
    },
  },
})
