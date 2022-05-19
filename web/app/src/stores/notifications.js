import { defineStore } from 'pinia'

/** Time after which a notification is hidden. */
const MESSAGE_HIDE_DELAY_MS = 5000;

/**
 * Store notifications to users of the web frontend.
 */
export const useNotifs = defineStore('notifications', {
  state: () => ({
    /**
     * History of notifications.
     *
     * The 'id' is just for Tabulator to uniquely identify rows, in order to be
     * able to scroll to them.
     *
     * @type {{ id: Number, msg: string, time: Date }[]} */
    history: [],
    /** @type { id: Number, msg: string, time: Date } */
    last: "",

    hideTimerID: 0,
    lastID: 0,
  }),
  getters: {
    empty: (state) => state.history.length == 0,
  },
  actions: {
    /**
     * Add a simple notification.
     * @param {string} message
     */
    add(message) {
      const notif = {id: this._generateID(), msg: message, time: new Date()};
      this.history.push(notif);
      this.last = notif;
      console.log("New notification:", plain(notif));
      this._prune();
      this._restartHideTimer();
    },

    /**
     * @param {API.SioTaskUpdate} taskUpdate Task update received via SocketIO.
     */
    addTaskUpdate(taskUpdate) {
      console.log('Received task update:', taskUpdate);
      let msg = `Task ${taskUpdate.name}`;
      if (taskUpdate.previous_status && taskUpdate.previous_status != taskUpdate.status) {
        msg += ` changed status ${taskUpdate.previous_status} ➜ ${taskUpdate.status}`;
      }
      if (taskUpdate.activity) {
        msg += `: ${taskUpdate.activity}`;
      }
      this.add(msg)
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
    _generateID() {
      return ++this.lastID;
    }
  },
})
