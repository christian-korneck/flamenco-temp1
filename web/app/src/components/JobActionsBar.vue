<template>
  <section class="action-bar jobs">
    <button class="action delete dangerous" :disabled="!jobs.canDelete" v-on:click="onButtonDelete">Delete</button>
  </section>
</template>

<script>
import { useJobs } from '@/stores/jobs';

export default {
  name: "JobActionsBar",
  events: ["actionDone", "apiError"],
  data: () => ({
    jobs: useJobs(),
  }),
  computed: {
  },
  methods: {
    onButtonDelete() {
      const numJobs = this.jobs.numSelected;
      this.jobs.deleteJobs()
        .then(() => {
          this.$emit("actionDone", `${numJobs} jobs marked for deletion`)
        })
        .catch((error) => {
          this.$emit("apiError", error);
        })
    },
  }
}

</script>

<style scoped>
</style>
