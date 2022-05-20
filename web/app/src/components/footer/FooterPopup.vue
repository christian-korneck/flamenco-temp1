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
</script>

<template>
  <section class="footer-popup">
    <header>
      <button :class='["footer-tab", {"active-tab": currentTab == "NotificationList"}]' @click="currentTab = 'NotificationList'">Notifications</button>
      <button :class='["footer-tab", {"active-tab": currentTab == "TaskLog"}]' @click="currentTab = 'TaskLog'">Task Log</button>
      <connection-status />
      <button class='close' @click="emit('clickClose')" title="Close Pop-over">X</button>
    </header>
    <component :is="tabs[currentTab]" class="tab"></component>
  </section>
</template>
