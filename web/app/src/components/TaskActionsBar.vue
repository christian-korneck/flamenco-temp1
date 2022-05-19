<template>
  <section class="btn-bar tasks">
    <button class="btn cancel" :disabled="!tasks.canCancel" v-on:click="onButtonCancel">Cancel Task</button>
    <button class="btn requeue" :disabled="!tasks.canRequeue" v-on:click="onButtonRequeue">Requeue</button>
  </section>
</template>

<script>
import { useTasks } from '@/stores/tasks';
import { useNotifs } from '@/stores/notifications';

export default {
  name: "TaskActionsBar",
  data: () => ({
    tasks: useTasks(),
    notifs: useNotifs(),
  }),
  computed: {
  },
  methods: {
    onButtonCancel() {
      return this._handleTaskActionPromise(
        this.tasks.cancelTasks(), "cancelled");
    },
    onButtonRequeue() {
      return this._handleTaskActionPromise(
        this.tasks.requeueTasks(), "requeueing");
    },

    _handleTaskActionPromise(promise, description) {
      // const numTasks = this.tasks.numSelected;
      const numTasks = 1;
      return promise
        .then(() => {
          // There used to be a call to `this.notifs.add(message)` here, but now
          // that task status changes are logged in the notifications anyway,
          // it's no longer necessary.
          // This function is still kept, in case we want to bring back the
          // notifications when multiple tasks can be selected. Then a summary
          // ("N tasks requeued") could be logged here.
        })
        .catch((error) => {
          const errorMsg = JSON.stringify(error); // TODO: handle API errors better.
          this.notifs.add(`Error: ${errorMsg}`);
        })
    },
  }
}
</script>
