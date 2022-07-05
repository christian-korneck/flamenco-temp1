<script setup>
import { useSlots, ref, provide } from "vue";
const emit = defineEmits(['clickedJobDetailsTab',])
const slots = useSlots();

const tabTitles = ref(slots.default().map((tab) => tab.props.title));
const selectedTitle = ref(tabTitles.value[0]);
provide("selectedTitle", selectedTitle);

function updateTabTitle(title) {
  selectedTitle.value = title;
  emit('clickedJobDetailsTab');
}

</script>

<template>
  <nav>
    <ul class="tabs-header">
      <li
        v-for="title in tabTitles"
        :key="title"
        class="tab-item"
        :class="{ active: selectedTitle === title }"
        @click="updateTabTitle(title)"
      >
        {{ title }}
      </li>
    </ul>
    <slot />
  </nav>
</template>

<style scoped>
nav {
  margin: var(--spacer-sm) 0;
}

.tabs-header {
  align-items: center;
  display: flex;
  gap: var(--spacer);
  list-style: none;
}
.tab-item {
  border-bottom: var(--border-width) solid transparent;
  color: var(--color-text-hint);
  cursor: pointer;
  padding: var(--spacer-sm) 0;
  transition: border-color var(--transition-speed) ease-in-out, color var(--transition-speed) ease-in-out;
  user-select: none;
}

.tab-item:hover {
  border-color: var(--color-text);
  color: var(--color-text);
}

.tab-item.active {
  border-color: var(--color-accent);
  color: var(--color-accent);
}
</style>
