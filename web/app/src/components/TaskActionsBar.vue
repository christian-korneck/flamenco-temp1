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
        this.tasks.requeueTasks(), "requeued");
    },

    _handleTaskActionPromise(promise, description) {
      // const numTasks = this.tasks.numSelected;
      const numTasks = 1;
      return promise
        .then(() => {
          let message;
          if (numTasks == 1) {
            message = `Task ${description}`;
          } else {
            message = `${numTasks} tasks ${description}`;
          }
          this.notifs.add(message);
        })
        .catch((error) => {
          const errorMsg = JSON.stringify(error); // TODO: handle API errors better.
          this.notifs.add(`Error: ${errorMsg}`);
        })
    },
  }
}
</script>
