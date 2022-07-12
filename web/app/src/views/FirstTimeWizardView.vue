<template>
  <div class="first-time-wizard">
    <h1>Welcome to Flamenco 3!</h1>

    <section>
      <p>Before Flamenco 3 can be used, a few things need to be set up.</p>
      <p>This wizard will guide you through the configuration.</p>
    </section>
    <section>
      <h2>Shared Storage</h2>

      <p>Flamenco needs some shared storage, to have a central place where the
        Manager and Workers exchange files. This could be a NAS in your network,
        or some other file sharing server.</p>

      <p class="hint">Using a service like Syncthing, ownCloud, or Dropbox for
        this is not recommended, as Flamenco does not know when every machine has
        received the files.</p>

      <form @submit.prevent="checkSharedStoragePath">
        <input v-model="sharedStoragePath" type="text">
        <button type="submit">Check</button>
      </form>

      <p v-if="sharedStorageCheckResult != null"
        :class="{ 'check-ok': sharedStorageCheckResult.is_usable, 'check-failed': !sharedStorageCheckResult.is_usable }">
        {{ sharedStorageCheckResult.cause }}
      </p>
    </section>

    <section>
      <h2>Which Blender?</h2>
    </section>
  </div>

  <footer class="app-footer">
    <notification-bar />
  </footer>

  <update-listener ref="updateListener" @sioReconnected="onSIOReconnected" @sioDisconnected="onSIODisconnected" />
</template>

<script>
import NotificationBar from '@/components/footer/NotificationBar.vue'
import UpdateListener from '@/components/UpdateListener.vue'
import { MetaApi, PathCheckInput } from "@/manager-api";
import { apiClient } from '@/stores/api-query-count';

export default {
  name: 'FirstTimeWizardView',
  components: {
    NotificationBar,
    UpdateListener,
  },
  data: () => ({
    sharedStoragePath: "",
    sharedStorageCheckResult: null,
    metaAPI: new MetaApi(apiClient),
  }),
  computed: {
    cleanSharedStoragePath() {
      return this.sharedStoragePath.trim();
    },
  },
  methods: {
    // SocketIO connection event handlers:
    onSIOReconnected() {
    },
    onSIODisconnected(reason) {
    },

    checkSharedStoragePath() {
      const pathCheck = new PathCheckInput(this.cleanSharedStoragePath);
      console.log("requesting path check:", pathCheck);
      this.metaAPI.checkSharedStoragePath({ pathCheckInput: pathCheck })
        .then((result) => {
          console.log("Storage path check result:", result);
          this.sharedStorageCheckResult = result;
        })
        .catch((error) => {
          console.log("Error checking storage path:", error);
        })
    },
  },
}
</script>
