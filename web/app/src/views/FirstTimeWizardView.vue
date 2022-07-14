<template>
  <div class="first-time-wizard">
    <h1>Welcome to Flamenco!</h1>

    <section>
      <p>Before Flamenco can be used, a few things need to be set up. This
        wizard will guide you through the configuration.</p>
    </section>
    <section>
      <h2>Shared Storage</h2>

      <p>Flamenco needs some shared storage, to have a central place where the
        Manager and Workers exchange files. This could be a NAS in your network,
        or some other file sharing server.</p>

      <p>Make sure this path is the same for all machines involved.</p>

      <p class="hint">Using a service like Syncthing, ownCloud, or Dropbox for
        this is not recommended, as Flamenco does not know when every machine has
        received the files.</p>

      <!-- TODO: @submit.prevent makes the button triggerable by pressing ENTER
        in the input field, but also prevents the browser from caching
        previously-used values. Would be great if we could have both. -->
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

      <p>Choose which Blender to use below:</p>

      <p v-if="isBlenderExeFinding">... finding Blenders ...</p>
      <div v-for="blender in allBlenders" class="blender-selector"
        :class="{ 'selected-blender': (blender == selectedBlender) }">
        <dl>
          <dt>Version</dt>
          <dd>{{ blender.cause }}</dd>

          <dt>Path</dt>
          <dd>{{ blender.path }}</dd>

          <dt>Source</dt>
          <dd>{{ sourceLabels[blender.source] }}</dd>
        </dl>
        <button @click="selectedBlender = blender" :disabled="selectedBlender == blender">Use this Blender</button>
      </div>

      <p>Or provide an alternative command to try:</p>

      <form @submit.prevent="checkBlenderExePath">
        <input v-model="customBlenderExe" type="text">
        <button type="submit">Check</button>
      </form>
      <p v-if="isBlenderExeChecking">... checking ...</p>
      <p v-if="blenderExeCheckResult != null && blenderExeCheckResult.is_usable" class="check-ok">
        Found something, it is selected above.</p>
      <p v-if="blenderExeCheckResult != null && !blenderExeCheckResult.is_usable" class="check-failed">
        {{ blenderExeCheckResult.cause }}</p>
    </section>

    <section v-if="isConfigComplete">
      <h2>The Final Step</h2>
      <p>This is the configuration that will be used by Flamenco:</p>
      <dl>
        <dt>Storage</dt>
        <dd>{{ sharedStorageCheckResult.path }}</dd>
        <dt>Blender</dt>
        <dd v-if="selectedBlender.source == 'file_association'">
          Whatever Blender is associated with .blend files
          (currently "<code>{{ selectedBlender.path }}</code>")
        </dd>
        <dd v-if="selectedBlender.source == 'path_envvar'">
          The command "<code>{{ selectedBlender.input }}</code>" as found on <code>$PATH</code>
          (currently "<code>{{ selectedBlender.path }}</code>")
        </dd>
        <dd v-if="selectedBlender.source == 'input_path'">
          The command you provided:
          "<code>{{ selectedBlender.path }}</code>"
        </dd>
      </dl>
      <button @click="confirmWizard" :disabled="isConfirming">Confirm</button>
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
import { MetaApi, PathCheckInput, WizardConfig } from "@/manager-api";
import { apiClient } from '@/stores/api-query-count';

export default {
  name: 'FirstTimeWizardView',
  components: {
    NotificationBar,
    UpdateListener,
  },
  data: () => ({
    sharedStoragePath: "",
    sharedStorageCheckResult: null, // api.PathCheckResult
    metaAPI: new MetaApi(apiClient),

    allBlenders: [], // combination of autoFoundBlenders and blenderExeCheckResult.

    autoFoundBlenders: [], // list of api.BlenderPathCheckResult
    isBlenderExeFinding: false,
    selectedBlender: null, // the chosen api.BlenderPathCheckResult

    customBlenderExe: "",
    isBlenderExeChecking: false,
    blenderExeCheckResult: null, // api.BlenderPathCheckResult
    sourceLabels: {
      file_association: "This Blender runs when you double-click a .blend file.",
      path_envvar: "This Blender was found on the $PATH environment.",
      input_path: "You pointed Flamenco to this executable.",
    },
    isConfirming: false,
  }),
  computed: {
    cleanSharedStoragePath() {
      return this.sharedStoragePath.trim();
    },
    cleanCustomBlenderExe() {
      return this.customBlenderExe.trim();
    },
    isConfigComplete() {
      return (this.sharedStorageCheckResult != null && this.sharedStorageCheckResult.is_usable) &&
        (this.selectedBlender != null && this.selectedBlender.is_usable);
    },
  },
  mounted() {
    this.findBlenderExePath();
  },
  methods: {
    // SocketIO connection event handlers:
    onSIOReconnected() {
    },
    onSIODisconnected(reason) {
    },

    // TODO: add a Refresh button that calls this again.
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

    findBlenderExePath() {
      this.isBlenderExeFinding = true;
      this.autoFoundBlenders = [];

      console.log("Finding Blender");
      this.metaAPI.findBlenderExePath()
        .then((result) => {
          console.log("Result of finding Blender:", result);
          this.autoFoundBlenders = result;
          this._refreshAllBlenders();
        })
        .catch((error) => {
          console.log("Error finding Blender:", error);
        })
        .finally(() => {
          this.isBlenderExeFinding = false;
        })
    },

    checkBlenderExePath() {
      const exeToTry = this.cleanCustomBlenderExe;
      if (exeToTry == "") {
        // Just erase any previously-found custom Blender executable.
        this.isBlenderExeChecking = false;
        this.blenderExeCheckResult = null;
        this._refreshAllBlenders();
        return;
      }

      this.isBlenderExeChecking = true;
      this.blenderExeCheckResult = null;

      const pathCheck = new PathCheckInput(exeToTry);
      console.log("requesting path check:", pathCheck);
      this.metaAPI.checkBlenderExePath({ pathCheckInput: pathCheck })
        .then((result) => {
          console.log("Blender exe path check result:", result);
          this.blenderExeCheckResult = result;
          if (result.is_usable) {
            this.selectedBlender = result;
          }
          this._refreshAllBlenders();
        })
        .catch((error) => {
          console.log("Error checking storage path:", error);
        })
        .finally(() => {
          this.isBlenderExeChecking = false;
        })
    },

    _refreshAllBlenders() {
      if (this.blenderExeCheckResult == null || !this.blenderExeCheckResult.is_usable) {
        this.allBlenders = this.autoFoundBlenders;
      } else {
        this.allBlenders = this.autoFoundBlenders.concat([this.blenderExeCheckResult]);
      }
    },

    confirmWizard() {
      const wizardConfig = new WizardConfig(
        this.sharedStorageCheckResult.path,
        this.selectedBlender,
      );
      console.log("saving configuration:", wizardConfig);
      this.isConfirming = true;
      this.metaAPI.saveWizardConfig({ wizardConfig: wizardConfig })
        .then((result) => {
          console.log("Wizard config saved, reload the page");
        })
        .catch((error) => {
          console.log("Error saving wizard config:", error);
        })
        .finally(() => {
          this.isConfirming = false;
        })
    },
  },
}
</script>
