<script setup>
import NotificationList from './NotificationList.vue'
import ConnectionStatus from '@/components/ConnectionStatus.vue'

const emit = defineEmits(['clickClose'])
</script>

<template>
  <section class="footer-popup">
    <header>
      <h3 class="sub-title">Notifications</h3>
      <connection-status />
      <button class='close' @click="emit('clickClose')">X</button>
    </header>
    <notification-list />
  </section>
</template>
