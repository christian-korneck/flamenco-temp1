<template>
  <section class="action-bar jobs">
    <button class="action delete dangerous" :disabled="!jobs.canDelete" v-on:click="onButtonDelete">Delete</button>
    <button class="action cancel" :disabled="!jobs.canCancel" v-on:click="onButtonCancel">Cancel</button>
    <button class="action requeue" :disabled="!jobs.canRequeue" v-on:click="onButtonRequeue">Requeue</button>
  </section>
</template>

<script>
import { useJobs } from '@/stores/jobs';
import { useNotifs } from '@/stores/notifications';

export default {
  name: "JobActionsBar",
  data: () => ({
    jobs: useJobs(),
    notifs: useNotifs(),
  }),
  computed: {
  },
  methods: {
    onButtonDelete() {
      return this._handleJobActionPromise(
        this.jobs.deleteJobs(), "marked for deletion");
    },
    onButtonCancel() {
      return this._handleJobActionPromise(
        this.jobs.cancelJobs(), "marked for cancellation");
    },
    onButtonRequeue() {
      return this._handleJobActionPromise(
        this.jobs.requeueJobs(), "requeued");
    },

    _handleJobActionPromise(promise, description) {
      const numJobs = this.jobs.numSelected;
      return promise
        .then(() => {
          let message;
          if (numJobs == 1) {
            message = `Job ${description}`;
          } else {
            message = `${numJobs} jobs ${description}`;
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

<style scoped>
</style>
