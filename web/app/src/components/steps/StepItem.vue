<script setup>
import { inject } from 'vue';
const props = defineProps({
  title: String,
});
const selectedStep = inject('selectedStep');
</script>

<template>
  <div v-show="selectedStep === title">
    <slot />
  </div>
</template>
