<template>
  <div class="btn-bar jobs">
    <button class="btn cancel" :disabled="!jobs.canCancel" v-on:click="onButtonCancel">Cancel Job</button>
    <button class="btn requeue" :disabled="!jobs.canRequeue" v-on:click="onButtonRequeue">Requeue</button>
    <!-- <button class="action delete dangerous" :disabled="!jobs.canDelete" v-on:click="onButtonDelete">Delete</button> -->
  </div>
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
      return promise
        .then(() => {
          // There used to be a call to `this.notifs.add(message)` here, but now
          // that job status changes are logged in the notifications anyway,
          // it's no longer necessary.
          // This function is still kept, in case we want to bring back the
          // notifications when multiple jobs can be selected. Then a summary
          // ("N jobs requeued") could be logged here.
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
