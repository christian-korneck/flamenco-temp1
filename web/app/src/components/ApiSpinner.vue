<template>
  <span class="api-spinner" :class="{ running: numRunningQueries > 0 }">&#128008;</span>
</template>

<script>
export default {
  props: ["numRunningQueries"],
};
</script>

<style scoped>
span {
  visibility: hidden;
  opacity: 0;
  transition-delay: 100ms;
  transition-duration: 300ms;
}

span.running {
  visibility: visible;
  opacity: 1;
  transition-delay: 0;
}
</style>
