<template>
  <select v-model="selectedAction">
    <option value="" selected>
      <template v-if="!hasActiveWorker">Select a Worker</template>
      <template v-else>Choose an action...</template>
    </option>
    <option v-for="(action, key) in WORKER_ACTIONS" :value="key">{{ action.label }}</option>
  </select>
  <button
      :disabled="!canPerformAction"
      class="btn"
      @click.prevent="performWorkerAction"
      >Apply</button>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useWorkers } from '@/stores/workers';
import { useNotifs } from '@/stores/notifications';
import { WorkerMgtApi, WorkerStatusChangeRequest } from '@/manager-api';
import { apiClient } from '@/stores/api-query-count';

/* Freeze to prevent Vue.js from creating getters & setters all over this object.
 * We don't need it to be tracked, as it won't be changed anyway. */
const WORKER_ACTIONS = Object.freeze({
    offline_lazy: {
        label: 'Shut Down (after task is finished)',
        icon: '✝',
        title: 'Shut down the worker after the current task finishes. The worker may automatically restart.',
        target_status: 'offline',
        lazy: true,
    },
    offline_immediate: {
        label: 'Shut Down (immediately)',
        icon: '✝!',
        title: 'Immediately shut down the worker. It may automatically restart.',
        target_status: 'offline',
        lazy: false,
    },
    asleep_lazy: {
        label: 'Send to Sleep (after task is finished)',
        icon: '😴',
        title: 'Let the worker sleep after finishing this task.',
        target_status: 'asleep',
        lazy: true,
    },
    asleep_immediate: {
        label: 'Send to Sleep (immediately)',
        icon: '😴!',
        title: 'Let the worker sleep immediately.',
        target_status: 'asleep',
        lazy: false,
    },
    wakeup: {
        label: 'Wake Up',
        icon: '😃',
        title: 'Wake the worker up. A sleeping worker can take a minute to respond.',
        target_status: 'awake',
        lazy: false,
    },
});

const selectedAction = ref('');
const workers = useWorkers();
const hasActiveWorker = computed(() => !!workers.activeWorkerID);
const canPerformAction = computed(() => hasActiveWorker && !!selectedAction.value);
const notifs = useNotifs();

function performWorkerAction() {
  const workerID = workers.activeWorkerID;
  if (!workerID) {
    notifs.add("Select a Worker before applying an action.");
    return;
  }

  const api = new WorkerMgtApi(apiClient);
  const action = WORKER_ACTIONS[selectedAction.value];
  const statuschange = new WorkerStatusChangeRequest(action.target_status, action.lazy);
  console.log("Requesting worker status change", statuschange);
  api.requestWorkerStatusChange(workerID, statuschange)
    .then((result) => notifs.add(`Worker status change to ${action.target_status} confirmed.`))
    .catch((error) => notifs.add(`Error requesting worker status change: ${error}`));
}

</script>
