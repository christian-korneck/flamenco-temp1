<template>
  <div class="col col-workers-list">
    <workers-table ref="workersTable" :activeWorkerID="workerID" @tableRowClicked="onTableWorkerClicked" />
  </div>
  <div class="col col-workers-details">
    <worker-details :workerData="workers.activeWorker" />
  </div>
  <footer class="app-footer">
    <notification-bar />
    <update-listener ref="updateListener" mainSubscription="allWorkers"
      @workerUpdate="onSIOWorkerUpdate"
      @sioReconnected="onSIOReconnected" @sioDisconnected="onSIODisconnected" />
  </footer>
</template>

<style scoped>
.col-workers-list {
  grid-area: col-1;
}
.col-workers-2 {
  grid-area: col-2;
}
</style>

<script>
import { WorkerMgtApi } from '@/manager-api';
import { useNotifs } from '@/stores/notifications'
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
    notifs: useNotifs(),
    api: new WorkerMgtApi(apiClient),
  }),
  mounted() {
    window.workersView = this;
    this._fetchWorker(this.workerID);

    document.body.classList.add('is-two-columns');
  },
  unmounted() {
    document.body.classList.remove('is-two-columns');
  },
  watch: {
    workerID(newWorkerID, oldWorkerID) {
      this._fetchWorker(newWorkerID);
    },
  },
  methods: {
    // SocketIO connection event handlers:
    onSIOReconnected() {
      this.$refs.workersTable.onReconnected();
      this._fetchWorker(this.workerID);
    },
    onSIODisconnected(reason) {
    },
    onSIOWorkerUpdate(workerUpdate) {
      this.notifs.addWorkerUpdate(workerUpdate);

      if (this.$refs.workersTable) {
        this.$refs.workersTable.processWorkerUpdate(workerUpdate);
      }
      if (this.workerID != workerUpdate.id)
        return;

      this._fetchWorker(this.workerID);
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
