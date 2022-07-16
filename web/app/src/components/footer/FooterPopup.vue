<script setup>
import { ref, watch } from 'vue'
import NotificationList from './NotificationList.vue'
import TaskLog from './TaskLog.vue'
import ConnectionStatus from '@/components/ConnectionStatus.vue'

const emit = defineEmits(['clickClose'])

const initialTab = localStorage.getItem("footer-popover-active-tab") || 'NotificationList';
const currentTab = ref(initialTab)
const tabs = { NotificationList, TaskLog }

watch(currentTab, async (newTab) => {
  localStorage.setItem("footer-popover-active-tab", newTab);
});

function showTaskLogTail() {
  currentTab.value = 'TaskLog';
}
defineExpose({
  showTaskLogTail,
});
</script>

<template>
  <footer>
    <nav>
      <ul>
        <li
          :class='["footer-tab", {"active": currentTab == "NotificationList"}]'
          @click="currentTab = 'NotificationList'">
            Notifications
        </li>
        <li
          :class='["footer-tab", {"active": currentTab == "TaskLog"}]'
          @click="currentTab = 'TaskLog'">
            Task Log
        </li>
        <connection-status />
        <li
          class="collapse"
          @click="emit('clickClose')"
          title="Collapse">
          &#10005;
        </li>
      </ul>
    </nav>
    <component :is="tabs[currentTab]" class="tab"></component>
  </footer>
</template>

<style scoped>
footer {
  background-color: var(--color-background-column);
  border-radius: var(--border-radius);
  bottom: var(--grid-gap);
  box-shadow: 0 0 5rem rgba(0, 0, 0, .66), 0 0 1.33rem rgba(0, 0, 0, .66);
  left: var(--grid-gap);
  padding: var(--spacer-xs) var(--spacer-sm) var(--spacer-sm);
  position: fixed;
  right: var(--grid-gap);
}

footer nav {
  margin-bottom: var(--spacer-sm);
}

footer nav ul {
  align-items: center;
  display: flex;
  gap: var(--spacer);
  list-style: none;
}

footer nav ul li {
  border-bottom: var(--border-width) solid transparent;
  color: var(--color-text-hint);
  cursor: pointer;
  padding: var(--spacer-sm) 0;
  transition: border-color var(--transition-speed) ease-in-out, color var(--transition-speed) ease-in-out;
  user-select: none;
}

footer nav ul li:hover {
  border-color: var(--color-text);
  color: var(--color-text);
}

footer nav ul li.active {
  border-color: var(--color-accent);
  color: var(--color-accent);
}

.collapse {
  border: none;
  flex: 1;
  text-align: right;
  font-size: 1.5rem;
  padding: 0 var(--spacer-sm) 0;
}


footer button.footer-tab {
  border: none;
  margin-right: 1rem;
}
footer button.footer-tab.active-tab {
  color: var(--color-accent-text);
  background-color: var(--color-accent-background);
}

footer .tabulator-tableholder {
  /* Force a visible scroll bar, so that the notification history table always
   * has the same available width. Without this, Tabulator won't properly act on
   * the reduction in width when the scrollbar appears, and will show a
   * horizontal scrollbar as well. */
  overflow-y: scroll;
}

footer .tabulator .tabulator-tableholder .tabulator-placeholder .tabulator-placeholder-contents {
  color: var(--color-text-hint);
  font-size: var(--font-size);
  font-weight: normal;
  user-select: none;
}
</style>
