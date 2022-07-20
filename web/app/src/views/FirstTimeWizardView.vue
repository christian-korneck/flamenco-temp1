<template>
  <div class="setup-container">
    <steps-wrapper>
      <step-item title="Welcome">
        <p>Let's set you up.</p>
      </step-item>

      <step-item title="Shared Storage">
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
      </step-item>
      <step-item title="Blender">
        <p>Choose which Blender to use below:</p>

        <div v-if="isBlenderExeFinding" class="is-in-progress">Looking for Blender installs...</div>

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

        <div v-if="isBlenderExeChecking" class="is-in-progress">Checking...</div>

        <p v-if="blenderExeCheckResult != null && blenderExeCheckResult.is_usable" class="check-ok">
          Found something, it is selected above.</p>
        <p v-if="blenderExeCheckResult != null && !blenderExeCheckResult.is_usable" class="check-failed">
          {{ blenderExeCheckResult.cause }}</p>
      </step-item>
      <step-item title="Review">
        <div v-if="isConfigComplete">
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
        </div>
        <p v-if="isConfirmed" class="check-ok">Configuration has been saved, Flamenco will restart.</p>
        <button @click="confirmWizard" :disabled="isConfirming">Confirm</button>
      </step-item>
    </steps-wrapper>
  </div>

  <footer class="app-footer">
    <notification-bar />
  </footer>

  <update-listener ref="updateListener" @sioReconnected="onSIOReconnected" @sioDisconnected="onSIODisconnected" />
</template>

<script>
import NotificationBar from '@/components/footer/NotificationBar.vue'
import UpdateListener from '@/components/UpdateListener.vue'
import StepItem from '@/components/steps/StepItem.vue';
import StepsWrapper from '@/components/steps/StepsWrapper.vue';
import { MetaApi, PathCheckInput, WizardConfig } from "@/manager-api";
import { apiClient } from '@/stores/api-query-count';

export default {
  name: 'FirstTimeWizardView',
  components: {
    NotificationBar,
    UpdateListener,
    StepsWrapper,
    StepItem,
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
    isConfirmed: false,
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

    document.body.classList.add('is-first-time-wizard');
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
      this.isConfirmed = false;
      this.metaAPI.saveWizardConfig({ wizardConfig: wizardConfig })
        .then((result) => {
          console.log("Wizard config saved, reload the page");
          this.isConfirmed = true;
          // Give the Manager some time to restart.
          window.setTimeout(() => { window.location.reload() }, 2000);
        })
        .catch((error) => {
          console.log("Error saving wizard config:", error);
          // Only clear this flag on an error.
          this.isConfirming = false;
        })
    },
  },
}
</script>
<style>


body.is-first-time-wizard #app {
  grid-template-areas:
    "header"
    "col-full-width"
    "footer";
  grid-template-columns: 1fr;
}

@media (max-width: 1280px) {
  body.is-first-time-wizard #app {
    grid-template-areas:
      "header"
      "col-full-width"
      "footer";
    grid-template-columns: 1fr;
    grid-template-rows: var(--header-height) 1fr var(--footer-height);
  }
}

.btn-bar-wide .btn:last-child {
  margin-left: auto;
}

.setup-container {
  --color-check-failed: var(--color-status-failed);
  --color-check-ok: var(--color-status-completed);

  max-width: 640px;
  margin: 20vh auto auto;
  width: 100%;
}

.setup-container h1 {
  font-size: xx-large;
}

.setup-container section {
  font-size: larger;
}

.setup-container p.hint {
  color: var(--color-text-hint);
  font-size: smaller;
}

.setup-container .check-ok {
  color: var(--color-check-ok);
}

.setup-container .check-failed {
  color: var(--color-check-failed);
}

.setup-container .check-ok::before {
  content: "✔ ";
}

.setup-container .check-failed::before {
  content: "❌ ";
}

.setup-container .blender-selector {
  padding: 0.5em;
  outline: thin solid var(--color-border);
  border-radius: var(--border-radius);
}

.setup-container .blender-selector.selected-blender {
  color: var(--color-accent-text);
  background-color: var(--color-accent-background);
  outline-width: var(--border-width);
}

.setup-container .blender-selector button {
  display: block;
  margin-left: auto;
}

.is-in-progress {
  animation: is-in-progress 3s infinite linear;
  background-image: linear-gradient(to left, var(--color-text-muted), rgba(255, 255, 255, 0.25), var(--color-text-muted));
  background-size: 200px;
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

@keyframes is-in-progress {
  0% {
    background-position: 0px;
  }

  100% {
    background-position: 200px;
  }
}
</style>
