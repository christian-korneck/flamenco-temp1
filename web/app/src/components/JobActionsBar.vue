<template>
  <section class="action-bar jobs">
    <button class="action delete dangerous" :disabled="!jobs.canDelete" v-on:click="onButtonDelete">Delete</button>
  </section>
</template>

<script>
import { useJobs } from '@/stores/jobs';
import { useNotifs } from '@/stores/notifications';

export default {
  name: "JobActionsBar",
  events: ["actionDone", "apiError"],
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
  }
}

</script>

<style scoped>
</style>
