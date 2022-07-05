<script setup>
import { useSlots, ref, provide } from "vue";
const slots = useSlots();

const tabTitles = ref(slots.default().map((tab) => tab.props.title));
const selectedTitle = ref(tabTitles.value[0]);
provide("selectedTitle", selectedTitle);
</script>

<template>
  <div>
    <ul class="tabs-header">
      <li
        v-for="title in tabTitles"
        :key="title"
        class="tab-item"
        :class="{ selected: selectedTitle === title }"
        @click="selectedTitle = title"
      >
        {{ title }}
      </li>
    </ul>
    <slot />
  </div>
</template>

<style scoped>
.tabs-header {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  justify-content: space-between;
  gap: 5px;
}
.tab-item {
  flex: 1;
  padding: 5px 0;
  cursor: pointer;
  user-select: none;
}
.tab-item.selected {
  border-bottom: 1px solid white;
}
</style>
