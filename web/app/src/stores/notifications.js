import { defineStore } from 'pinia'

/**
 * Store notifications to users of the web frontend.
 */
export const useNotifs = defineStore('notifications', {
  state: () => ({
    /** @type {{ msg: string, time: Date }[]} */
    history: [],
    /** @type { msg: string, time: Date } */
    last: null,
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
    },

    /* Ensure there is only 1000 items in the history. */
    _prune() {
      if (this.history.length <= 1000) return;
      this.history.splice(0, 1000 - this.history.length);
    },
  },
})
