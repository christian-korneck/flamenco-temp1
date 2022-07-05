<script setup>
import { computed } from 'vue'
import { indicator } from '@/statusindicator';

const props = defineProps(['availableStatuses', 'activeStatuses', 'classPrefix']);
const emit = defineEmits(['click'])

/**
 * visibleStatuses is a union between `availableStatuses` and `activeStatuses`,
 * and ensures that when a filter is active, it's also shown (even when it's no
 * longer available) so that it can be deactivated again.
 */
const visibleStatuses = computed(() => {
  const available = props.availableStatuses;
  const unavailable = props.activeStatuses.filter((status) => available.indexOf(status) == -1);
  return available.concat(unavailable);
})
</script>

<template>
  <ul class="status-filter-bar"
    :class="{'is-filtered': activeStatuses.length > 0}">
    <li v-for="status in visibleStatuses" class="status-filter-indicator"
      :data-status="status"
      :class="{active: activeStatuses.indexOf(status) >= 0}"
      @click="emit('click', status)"
      v-html="indicator(status, classPrefix)"
    ></li>
  </ul>
</template>
