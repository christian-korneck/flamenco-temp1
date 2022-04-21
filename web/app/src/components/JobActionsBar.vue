<template>
  <section class="action-bar jobs">
    <button class="action delete dangerous" :disabled="!jobs.canDelete" v-on:click="onButtonDelete">Delete</button>
    <button class="action cancel" :disabled="!jobs.canCancel" v-on:click="onButtonCancel">Cancel</button>
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
      const numJobs = this.jobs.numSelected;
      this.jobs.deleteJobs()
        .then(() => {
          this.notifs.add(`${numJobs} jobs marked for deletion`);
        })
        .catch((error) => {
          const errorMsg = JSON.stringify(error); // TODO: handle API errors better.
          this.notifs.add(`Error: ${errorMsg}`);
        })
    },
    onButtonCancel() {
      const numJobs = this.jobs.numSelected;
      this.jobs.cancelJobs()
        .then(() => {
          let message;
          if (numJobs == 1) {
            message = `Job marked for cancellation`;
          } else {
            message = `${numJobs} jobs marked for cancellation`;
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
