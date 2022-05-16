<script setup>
import { ref } from 'vue'
import { indicator } from '@/statusindicator';

defineProps(['availableStatuses', 'activeStatuses']);
const emit = defineEmits(['click'])
</script>

<template>
  <ul class="status-filter-bar">
    <li v-for="status in availableStatuses" class="status-filter-indicator"
      :data-status="status"
      :class="{active: activeStatuses.indexOf(status) >= 0}"
      @click="emit('click', status)"
      v-html="indicator(status)"
    ></li>
  </ul>
</template>
