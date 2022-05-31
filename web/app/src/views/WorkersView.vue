<template>
  <div class="col col-workers-list">
    <workers-table ref="workersTable" :activeWorkerID="workerID" @tableRowClicked="onTableWorkerClicked" />
  </div>
  <div class="col col-workers-details">
    <worker-details :workerData="workers.activeWorker" />
  </div>
  <footer>
    <notification-bar />
    <update-listener ref="updateListener"
      @sioReconnected="onSIOReconnected" @sioDisconnected="onSIODisconnected" />
</footer>
</template>

<style scoped>
.col-workers-list {
  grid-column-start: col-1;
  grid-column-end: col-2;
}
.col-workers-2 {
  grid-area: col-3;
}
</style>

<script>
import { WorkerMgtApi } from '@/manager-api';
import { useWorkers } from '@/stores/workers';
import { apiClient } from '@/stores/api-query-count';

import NotificationBar from '@/components/footer/NotificationBar.vue'
import UpdateListener from '@/components/UpdateListener.vue'
import WorkerDetails from '@/components/workers/WorkerDetails.vue'
import WorkersTable from '@/components/workers/WorkersTable.vue'

export default {
  name: 'WorkersView',
  props: ["workerID"], // provided by Vue Router.
  components: {
    NotificationBar,
    UpdateListener,
    WorkerDetails,
    WorkersTable,
  },
  data: () => ({
    workers: useWorkers(),
    api: new WorkerMgtApi(apiClient),
  }),
  mounted() {
    window.workersView = this;
    this._fetchWorker(this.workerID);
  },
  watch: {
    workerID(newWorkerID, oldWorkerID) {
      this._fetchWorker(newWorkerID);
    },
  },
  methods: {
    // SocketIO connection event handlers:
    onSIOReconnected() {
    },
    onSIODisconnected(reason) {
    },

    onTableWorkerClicked(rowData) {
      if (rowData.id == this.workerID) return;
      this._routeToWorker(rowData.id);
    },

    /**
     * @param {string} workerID worker ID to navigate to, can be empty string for "no active worker".
     */
    _routeToWorker(workerID) {
      const route = { name: 'workers', params: { workerID: workerID } };
      console.log("routing to worker", route.params);
      this.$router.push(route);
    },

    /**
     * Fetch worker info and set the active worker once it's received.
     * @param {string} workerID worker ID, can be empty string for "no worker".
     */
    _fetchWorker(workerID) {
      if (!workerID) {
        this.workers.deselectAllWorkers();
        return;
      }

      return this.api.fetchWorker(workerID)
        .then((worker) => this.workers.setActiveWorker(worker));
    },
  },
}
</script>
