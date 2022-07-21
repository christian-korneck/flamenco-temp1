<script setup>
const props = defineProps({
  title: String,
  nextLabel: {
    type: String,
    default: 'Next'
  },
  isBackVisible: {
    type: Boolean,
    default: true
  },
  isNextClickable: {
    type: Boolean,
    default: false
  },
});
</script>

<template>
  <div>
    <h2>{{ title }}</h2>
    <slot></slot>
    <div class="btn-bar btn-bar-wide">
      <button
        v-show="isBackVisible"
        @click="$emit('backClicked')"
        class="btn btn-lg"
      >Back
      </button>
      <button
        @click="$emit('nextClicked')"
        :disabled="!isNextClickable"
        class="btn btn-lg btn-primary"
      >
      {{ nextLabel }}
      </button>
    </div>
  </div>
</template>
