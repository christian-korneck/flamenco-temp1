<script setup>
import { inject, computed, provide } from "vue";
const props = defineProps({
  title: String,
});
const selectedTitle = inject("selectedTitle");
const isVisible = computed(() => selectedTitle.value === props.title)
provide("isVisible", isVisible);
</script>

<template>
  <div v-show="isVisible">
    <slot />
  </div>
</template>
