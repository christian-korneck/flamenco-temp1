<template>
  <div class="details-no-item-selected">
    <div class="get-the-addon">
      <p>Get the Blender add-on and submit a job.</p>
      <p><button class="btn btn-primary">Get the add-on!</button></p>
      <p>Use the URL below in the add-on preferences. Click on it to copy.</p>
      <p class="click-to-copy" title="Click to copy this URL" @click="copyElementText">{{ api() }}</p>
    </div>
  </div>
</template>

<script setup>
import { api } from '@/urls';
import { copyElementText } from '@/clipboard.js';
</script>

<style scoped>
.get-the-addon {
  color: var(--color-text-muted);
  display: block;
  text-align: center;
}

button {
  margin: var(--spacer);
  font-size: var(--font-size-lg);
}
</style>
