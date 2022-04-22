<template>
  <span class="api-spinner" :class="{ running: apiQueryCount.num > 0 }">&#128008;</span>
</template>

<script>
import { useAPIQueryCount } from '@/stores/api-query-count';

export default {
  data: () => ({
    apiQueryCount: useAPIQueryCount(),
  }),
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
