<template>
  <div class="task-details">
    <h2 class="column-title">Task Details</h2>
  </div>
</template>

<script lang="js">
export default {
  props: ["apiClient"],
  data: () => {
    return {
    };
  },
  mounted() {
    // Allow testing from the JS console:
    window.taskDetailsVue = this;
  },
  methods: {
  }
};
</script>

<style scoped>
</style>
