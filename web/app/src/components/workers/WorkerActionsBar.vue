<template>
  <div class="btn-bar workers">
    <button class="btn wakeup" v-on:click="onButtonWakeup">Wake Up</button>
    <button class="btn sleep" v-on:click="onButtonSleep">Sleep</button>
    <button class="btn offline dangerous" v-on:click="onButtonOffline">Offline</button>
  </div>
</template>

<script>
import { useWorkers } from '@/stores/workers';
import { useNotifs } from '@/stores/notifications';

export default {
  name: "WorkerActionsBar",
  data: () => ({
    workers: useWorkers(),
    notifs: useNotifs(),
  }),
  computed: {
  },
  methods: {
    onButtonWakeup() {
      return this._handleWorkerActionPromise(
        this.workers.reqStatusAwake());
    },
    onButtonSleep() {
      return this._handleWorkerActionPromise(
        this.workers.reqStatusAsleep());
    },
    onButtonOffline() {
      return this._handleWorkerActionPromise(
        this.workers.reqStatusOffline());
    },

    _handleWorkerActionPromise(promise) {
      return promise
        .catch((error) => {
          const errorMsg = JSON.stringify(error); // TODO: handle API errors better.
          this.notifs.add(`Error: ${errorMsg}`);
        })
    },
  }
}

</script>

<style scoped>
</style>
