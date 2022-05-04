<template>
  <span class="api-spinner" :class="{ running: apiQueryCount.num > 0 }">
    <svg class="spinner" viewBox="0 0 50 50">
      <circle class="path" cx="25" cy="25" r="20" fill="none" stroke-width="5"></circle>
    </svg>
  </span>
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
  transition-delay: 250ms;
  transition-duration: 500ms;
  font-weight: bold;
  margin-right: 2rem;
  position: relative;
}

.spinner {
  -webkit-animation: rotate 2s linear infinite;
          animation: rotate 2s linear infinite;
  z-index: 2;
  position: absolute;
  top: 50%;
  left: 50%;
  margin: -10px 0 0 -10px;
  width: 20px;
  height: 20px;
}
.spinner .path {
  stroke: var(--color-text-hint);
  stroke-linecap: round;
  -webkit-animation: dash 1.5s ease-in-out infinite;
          animation: dash 1.5s ease-in-out infinite;
}

@-webkit-keyframes rotate {
  100% {
    transform: rotate(360deg);
  }
}

@keyframes rotate {
  100% {
    transform: rotate(360deg);
  }
}

@-webkit-keyframes dash {
  0% {
    stroke-dasharray: 1, 150;
    stroke-dashoffset: 0;
  }
  50% {
    stroke-dasharray: 90, 150;
    stroke-dashoffset: -35;
  }
  100% {
    stroke-dasharray: 90, 150;
    stroke-dashoffset: -124;
  }
}

@keyframes dash {
  0% {
    stroke-dasharray: 1, 150;
    stroke-dashoffset: 0;
  }
  50% {
    stroke-dasharray: 90, 150;
    stroke-dashoffset: -35;
  }
  100% {
    stroke-dasharray: 90, 150;
    stroke-dashoffset: -124;
  }
}

span.running {
  visibility: visible;
  opacity: 1;
  transition-delay: 0;
}
</style>
