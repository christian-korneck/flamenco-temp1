<template>
  <div class="details-no-item-selected">
    <div class="get-the-addon">
      <p>Get the Blender add-on and submit a job.</p>
      <p><a class="btn btn-primary" :href="backendURL('/flamenco3-addon.zip')">Get the add-on!</a></p>
      <p>Use the URL below in the add-on preferences. Click on it to copy.</p>
      <p><span class="click-to-copy" title="Click to copy this URL" @click="copyElementText">{{ api() }}</span></p>
    </div>
  </div>
</template>

<script setup>
import { api, backendURL } from '@/urls';
import { copyElementText } from '@/clipboard.js';
</script>

<style scoped>
.get-the-addon {
  color: var(--color-text-muted);
  display: block;
  text-align: center;
}

.btn {
  margin: var(--spacer);
  font-size: var(--font-size-lg);
}

.click-to-copy {
  padding: var(--spacer-sm);
}
</style>
