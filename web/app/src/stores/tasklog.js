import { defineStore } from 'pinia'

// Maximum number of task log lines that will be stored.
const capacity = 1000;

/**
 * Store logs of the active task.
 */
export const useTaskLog = defineStore('taskLog', {
  state: () => ({
    /**
     * Task log entries.
     *
     * The 'id' is just for Tabulator to uniquely identify rows, in order to be
     * able to scroll to them and keep them in order.
     *
     * @type {{ id: Number, line: string }[]} */
    history: [],
    /** @type { id: Number, line: string } */
    last: "",

    lastID: 0,
  }),
  getters: {
    empty: (state) => state.history.length == 0,
  },
  actions: {
    /**
     * @param {API.SocketIOTaskLogUpdate} taskLogUpdate
     */
    addTaskLogUpdate(taskLogUpdate) {
      console.log('task log update:', taskLogUpdate);
      this.addChunk(taskLogUpdate.log);
    },

    /**
     * Erase the entire task log history. Use this when switching between tasks.
     */
    clear() {
      this.$patch((state) => {
        state.history = [];
        state.last = null;
        state.lastID = 0;
        state.hasChanged = true;
      });
    },

    /**
     * Add a task log chunk.
     * @param {string} logChunk
     */
    addChunk(logChunk) {
      const lines = logChunk.trimEnd().split('\n');
      if (lines.length == 0)
        return;

      if (lines.length > capacity) {
        // Only keep the `capacity` last lines, so that adding them to the
        // history will not overflow the capacity.
        lines.splice(0, lines.length - capacity);
      }

      this.$patch((state) => {
        let entry = null;

        // Make sure there is enough space to actually add the new lines.
        this._pruneState(state, lines.length);

        for (let line of lines) {
          entry = this._createEntry(state, line);
          state.history.push(entry);
        }

        if (entry == null) {
          console.warn("taskLog.addChunk: there were lines to add, but no entry created. Weird.");
          return;
        }

        state.last = entry;
        state.lastID = entry.id;
        state.hasChanged = true;
      });
    },

    _createEntry(state, line) {
      return {id: this._generateID(state), line: line};
    },

    /**
     * Ensure there is enough space in the history to fit `spaceForLineNum` lines.
     */
    _pruneState(state, spaceForLineNum) {
      if (spaceForLineNum > capacity) {
        // No need to calculate anything, just delete everything.
        state.history = [];
        return;
      }

      const pruneTo = capacity - spaceForLineNum;
      if (state.history.length <= pruneTo) return;

      const deleteCount = state.history.length - pruneTo;
      state.history.splice(0, deleteCount);
    },
    _generateID(state) {
      return ++state.lastID;
    }
  },
})
