import { defineStore } from 'pinia'

// 'use' prefix is idiomatic for Pinia stores.
// See https://pinia.vuejs.org/core-concepts/
export const useWorkers = defineStore('workers', {
  state: () => ({
    /** @type {API.Worker} */
    activeWorker: null,
    /**
     * ID of the active worker. Easier to query than `activeWorker ? activeWorker.id : ""`.
     * @type {string}
     */
    activeWorkerID: "",
  }),
  actions: {
    setActiveWorkerID(workerID) {
      this.$patch({
        activeWorker: {id: workerID, settings: {}, metadata: {}},
        activeWorkerID: workerID,
      });
    },
    setActiveWorker(worker) {
      // The "function" form of $patch is necessary here, as otherwise it'll
      // merge `worker` into `state.activeWorker`. As a result, it won't touch missing
      // keys, which means that metadata fields that existed on the previous worker
      // but not on the new one will still linger around. By passing a function
      // to `$patch` this is resolved.
      this.$patch((state) => {
        state.activeWorker = worker;
        state.activeWorkerID = worker.id;
        state.hasChanged = true;
      });
    },
    deselectAllWorkers() {
      this.$patch({
        activeWorker: null,
        activeWorkerID: "",
      });
    },
  },
})
