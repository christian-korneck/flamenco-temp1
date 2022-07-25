<template>
  <div class="setup-container">
    <h1>Flamenco Setup Assistant</h1>
    <ul class="progress">
      <li
        v-for="step in totalSetupSteps" :key="step"
        @click="jumpToStep(step)"
        :class="{
          current: step == currentSetupStep,
          done: step < overallSetupStep,
          done_previously: (step < overallSetupStep && currentSetupStep > step),
          done_and_current: step == currentSetupStep && (step < overallSetupStep || step == 1),
          disabled: step > overallSetupStep,
        }"
        >
        <span></span>
      </li>
      <div class="progress-bar"></div>
    </ul>
    <div class="setup-step step-welcome">

      <step-item
        v-show="currentSetupStep == 1"
        @next-clicked="nextStep"
        :is-next-clickable="true"
        :is-back-visible="false"
        title="Welcome!"
        next-label="Let's go"
      >
        <p>
          This setup assistant will guide you through the initial configuration of Flamenco. You will be up
          and running in a few minutes!
        </p>
        <p>Before we start, here is a quick overview of the Flamenco architecture.</p>

        <img src="@/assets/architecture.png" alt="Flamenco architecture"/>

        <p>The illustration shows the key components and how they interact together:</p>
        <ul>
          <li>
            <h3>Manager</h3>This application. It coordinates all the activity.
          </li>
          <li>
            <h3>Worker</h3>A workstation or dedicated rendering machine. It executes the tasks assigned by the Manager.
          </li>
          <li>
            <h3>Shared Storage</h3>A location accessible by the Manager and the Workers, where files, logs and internal previews can be saved.
          </li>
          <li>
            <h3>Blender Add-on</h3>This is needed to connect to the Manager and submit a job from Blender.
          </li>
        </ul>
        <p>More information is available on the online documentation at
          <a href="https://flamenco.blender.org">flamenco.blender.org</a>.</p>
      </step-item>

      <step-item
        v-show="currentSetupStep == 2"
        @next-clicked="nextStepAfterCheckSharedStoragePath"
        @back-clicked="prevStep"
        :is-next-clickable="sharedStoragePath.length > 0"
        title="Shared Storage"
      >
        <p>Specify a path (or drive) where you want to store your Flamenco data.</p>
        <p>
          The location of the shared storage should be accessible by Flamenco Manager and by the Workers.
          This could be:
        </p>

        <ul>
          <li>A local drive or folder (e.g. if you are working by yourself)</li>
          <li>A NAS in your network</li>
          <li>Other file sharing server</li>
        </ul>

        <p>
          Using services like Dropbox, Syncthing, or ownCloud for
          this is not recommended, as Flamenco can't coordinate data synchronization.
          <a href="https://flamenco.blender.org/usage/getting-started/">Learn more</a>.
        </p>

        <input
          v-model="sharedStoragePath"
          @keyup.enter="nextStepAfterCheckSharedStoragePath"
          type="text"
          placeholder="Path to shared storage"
          :class="{
            'is-invalid': (sharedStorageCheckResult != null) && !sharedStorageCheckResult.is_usable
          }"
        >
        <p v-if="sharedStorageCheckResult != null"
          :class="{
            'check-ok': sharedStorageCheckResult.is_usable,
            'check-failed': !sharedStorageCheckResult.is_usable
          }">
          {{ sharedStorageCheckResult.cause }}
        </p>
      </step-item>

      <step-item
        v-show="currentSetupStep == 3"
        @next-clicked="nextStep"
        @back-clicked="prevStep"
        :is-next-clickable="selectedBlender != null && selectedBlender.is_usable"
        title="Blender"
      >

        <div v-if="isBlenderExeFinding" class="is-in-progress">Looking for Blender installs...</div>

        <p v-if="autoFoundBlenders.length === 0">
          Provide a path to a Blender executable accessible by all Workers.
          <br/><br/>
          If your rendering setup features operating systems different form the one you are currently using,
          you can manually set up the other paths later.
        </p>

        <p v-else>Choose how a Worker should invoke the Blender command when performing a task:</p>

        <fieldset v-if="autoFoundBlenders.length >= 1">
          <label v-if="autoFoundBlenderPathEnvvar" for="blender-path_envvar">
            <div>
              <input
                v-model="selectedBlender"
                :value="autoFoundBlenderPathEnvvar"
                id="blender-path_envvar"
                name="blender"
                type="radio">
                {{ sourceLabels[autoFoundBlenderPathEnvvar.source] }}
            </div>
            <div class="setup-path-command">
              <span class="path">
                {{autoFoundBlenderPathEnvvar.path}}
              </span>
              <span
                :aria-label="autoFoundBlenderPathEnvvar.cause"
                class="command-preview"
                data-microtip-position="top"
                role="tooltip">
                [preview output]
              </span>
            </div>
          </label>

          <label v-if="autoFoundBlenderFileAssociation" for="blender-file_association">
            <div>
              <input
                v-model="selectedBlender"
                :value="autoFoundBlenderFileAssociation"
                id="blender-file_association"
                name="blender"
                type="radio">
                {{ sourceLabels[autoFoundBlenderFileAssociation.source] }}
            </div>
            <div class="setup-path-command">
              <span class="path">
                {{autoFoundBlenderFileAssociation.path}}
              </span>
              <span
                :aria-label="autoFoundBlenderFileAssociation.cause"
                class="command-preview"
                data-microtip-position="top"
                role="tooltip">
                [preview output]
              </span>
            </div>
          </label>

          <label for="blender-input_path">
            <div>
              <input
                type="radio"
                v-model="selectedBlender"
                name="blender"
                :value="blenderFromInputPath"
                id="blender-input_path"
                >
              {{ sourceLabels['input_path'] }}
            </div>
            <div>
              <input
                @input="checkBlenderExePath"
                v-model="customBlenderExe"
                :class="{'is-invalid': blenderExeCheckResult != null && !blenderExeCheckResult.is_usable}"
                type="text"
                placeholder="Path to Blender executable"
              >
              <p v-if="isBlenderExeChecking" class="is-in-progress">Checking...</p>
              <p v-if="blenderExeCheckResult != null && !blenderExeCheckResult.is_usable" class="check-failed">
                {{ blenderExeCheckResult.cause }}</p>
            </div>
          </label>
        </fieldset>

        <div v-if="autoFoundBlenders.length === 0">
          <input
            @input="checkBlenderExePath"
            v-model="customBlenderExe"
            :class="{'is-invalid': blenderExeCheckResult != null && !blenderExeCheckResult.is_usable}"
            type="text"
            placeholder="Path to Blender executable"
          >

          <p v-if="isBlenderExeChecking" class="is-in-progress">Checking...</p>
          <p v-if="blenderExeCheckResult != null && !blenderExeCheckResult.is_usable" class="check-failed">
            {{ blenderExeCheckResult.cause }}</p>
        </div>
      </step-item>

      <step-item
        v-show="currentSetupStep == 4"
        @next-clicked="confirmSetupAssistant"
        @back-clicked="prevStep"
        next-label="Confirm"
        title="Review"
        :is-next-clickable="setupConfirmIsClickable"
      >
        <div v-if="isConfigComplete">
          <p>This is the configuration that will be used by Flamenco:</p>
          <dl>
            <dt>Storage</dt>
            <dd>{{ sharedStorageCheckResult.path }}</dd>
            <dt>Blender Command</dt>
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
      </step-item>
    </div>
  </div>

  <footer class="app-footer">
    <notification-bar />
  </footer>

  <update-listener ref="updateListener" @sioReconnected="onSIOReconnected" @sioDisconnected="onSIODisconnected" />
</template>

<script>
import 'microtip/microtip.css'
import NotificationBar from '@/components/footer/NotificationBar.vue'
import UpdateListener from '@/components/UpdateListener.vue'
import StepItem from '@/components/steps/StepItem.vue';
import { MetaApi, PathCheckInput, SetupAssistantConfig } from "@/manager-api";
import { apiClient } from '@/stores/api-query-count';

function debounce(func, wait, immediate) {
  var timeout;
  return function() {
    var context = this, args = arguments;
    var later = function() {
      timeout = null;
      if (!immediate) func.apply(context, args);
    };
    var callNow = immediate && !timeout;
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
    if (callNow) func.apply(context, args);
  };
}

export default {
  name: 'SetupAssistantView',
  components: {
    UpdateListener,
    StepItem,
    NotificationBar,
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
      file_association: "Blender that runs when you double-click a .blend file:",
      path_envvar: "Blender found on the $PATH environment:",
      input_path: "Another Blender executable:",
    },
    isConfirming: false,
    isConfirmed: false,
    currentSetupStep: 1,
    overallSetupStep: 1,
    totalSetupSteps: 4,
  }),
  computed: {
    cleanSharedStoragePath() {
      return this.sharedStoragePath.trim();
    },
    cleanCustomBlenderExe() {
      return this.customBlenderExe.trim();
    },
    isSharedStorageValid() {
      return this.sharedStorageCheckResult != null && this.sharedStorageCheckResult.is_usable;
    },
    isSelectedBlenderValid() {
      return this.selectedBlender != null && this.selectedBlender.is_usable;
    },
    isConfigComplete() {
      return this.isSharedStorageValid && this.isSelectedBlenderValid;
    },
    autoFoundBlenderPathEnvvar() {
      return this.autoFoundBlenders.find(b => b.source === 'path_envvar');
    },
    autoFoundBlenderFileAssociation() {
      return this.autoFoundBlenders.find(b => b.source === 'file_association');
    },
    blenderFromInputPath() {
      return this.allBlenders.find(b => b.source === 'input_path');
    },
    setupConfirmIsClickable() {
      if (this.isConfirming || this.isConfirmed) {
        return false;
      } else {
        return true;
      }
    }
  },
  mounted() {
    this.findBlenderExePath();

    document.body.classList.add('is-setup-assistant');
  },
  methods: {
    // SocketIO connection event handlers:
    onSIOReconnected() {
    },
    onSIODisconnected(reason) {
    },

    nextStepAfterCheckSharedStoragePath() {
      const pathCheck = new PathCheckInput(this.cleanSharedStoragePath);
      console.log("requesting path check:", pathCheck);
      return this.metaAPI.checkSharedStoragePath({ pathCheckInput: pathCheck })
        .then((result) => {
          console.log("Storage path check result:", result);
          this.sharedStorageCheckResult = result;
          if (this.isSharedStorageValid) {
            this.nextStep();
          }
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
          } else if (this.selectedBlender.source === 'input_path') {
            this.selectedBlender = null;
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

    nextStep() {
      if (this.overallSetupStep <= this.currentSetupStep) {
        this.overallSetupStep = this.currentSetupStep + 1;
      }
      this.currentSetupStep++;
    },

    prevStep() {
      this.currentSetupStep--;
    },

    jumpToStep(step) {
      if (step <= this.overallSetupStep) {
        this.currentSetupStep = step;
      }
    },

    confirmSetupAssistant() {
      const setupAssistantConfig = new SetupAssistantConfig(
        this.sharedStorageCheckResult.path,
        this.selectedBlender,
      );
      console.log("saving configuration:", setupAssistantConfig);
      this.isConfirming = true;
      this.isConfirmed = false;
      this.metaAPI.saveSetupAssistantConfig({ setupAssistantConfig: setupAssistantConfig })
        .then((result) => {
          console.log("Setup Assistant config saved, reload the page");
          this.isConfirmed = true;
          // Give the Manager some time to restart.
          window.setTimeout(() => { window.location.reload() }, 2000);
        })
        .catch((error) => {
          console.log("Error saving setup assistan config:", error);
          // Only clear this flag on an error.
          this.isConfirming = false;
        })
    },
  },
  created() {
    this.checkBlenderExePath = debounce(this.checkBlenderExePath, 200)
  }
}
</script>
<style>

.step-welcome ul {
  padding-left: var(--spacer-xl);
  margin-bottom: var(--spacer-xl);
}

.setup-step ul li {
  margin-bottom: var(--spacer-xs);
}

.setup-step dl {
  background-color: var(--color-background);
  border-radius: var(--border-radius);
  padding: var(--spacer);
}

.setup-step dl dt {
  margin-bottom: var(--spacer-sm);
  width: initial;
}

.step-welcome ul li h3 {
  color: var(--color-text-highlight);
  font-size: var(--font-size-lg);
  margin-bottom: var(--spacer-xs);
}

.progress {
  --setup-progress-indicator-size: 8px;
  --setup-progress-indicator-border-width: 2px;
  --setup-progress-indicator-color: var(--color-text-hint);
  --setup-progress-indicator-color-done: var(--color-accent);
  --setup-progress-indicator-color-current: var(--color-accent-text);

  display: flex;
  justify-content: space-between;
  list-style: none;
  margin-bottom: 2rem;
  padding: 0;
  position: relative;
}

/* Progress indicator line between dots. */
.progress:before {
  background-color: var(--setup-progress-indicator-color);
  content: '';
  display: block;
  height: var(--setup-progress-indicator-border-width);
  position: absolute;
  top: calc(50% - calc(var(--setup-progress-indicator-border-width) / 2));
  transform: translateY(-50%);
  width: 100%;
}

.progress li {
  cursor: pointer;
}

.progress-bar {
  --width-each-segment: calc(100% / calc(v-bind('totalSetupSteps') - 1)); /* Substract 1 because the first step has no progress. */
  --progress-bar-width-at-current-step: calc(var(--width-each-segment) * calc(v-bind('currentSetupStep') - 1));

  position: absolute;
  top: calc(50% - calc(var(--setup-progress-indicator-border-width) / 2));
  transform: translateY(-50%);
  background-image: linear-gradient(to right, var(--color-accent) 85%, var(--color-accent-text));
  border-radius: 999rem;
  height: var(--setup-progress-indicator-border-width);
  position: absolute;
  transition: width 500ms ease-out;
  width: var(--progress-bar-width-at-current-step);
  z-index: 2;
}

/* Progress indicator dot.  */
.progress li span {
  background-color: var(--color-background);
  border-radius: 50%;
  border: var(--setup-progress-indicator-border-width) solid var(--color-background);
  box-shadow: 0 0 0 var(--setup-progress-indicator-border-width) var(--setup-progress-indicator-color);
  content: '';
  cursor: pointer;
  display: inline-block;
  height: var(--setup-progress-indicator-size);
  position: relative;
  width: var(--setup-progress-indicator-size);
  transition-duration: 500ms;
  transition-property: background-color, border, box-shadow;
  transition-delay: 250ms;
  z-index: 3;
}

.progress li.disabled span {
  cursor: not-allowed;
}

.progress li:not(.disabled):hover span {
  transform: scale(1.1);
}

.progress li.done span {
  background-color: var(--setup-progress-indicator-color);
  box-shadow: 0 0 0 var(--setup-progress-indicator-border-width) var(--setup-progress-indicator-color);
}

.progress li.done_previously span {
  background-color: var(--setup-progress-indicator-color-done);
  box-shadow: 0 0 0 var(--setup-progress-indicator-border-width) var(--setup-progress-indicator-color-done);
}

.progress li.current span {
  background-color: var(--color-background-column);
  box-shadow: 0 0 0 var(--setup-progress-indicator-border-width) var(--setup-progress-indicator-color-current);
}

.progress li.done_and_current span {
  background-color: var(--setup-progress-indicator-color-current);
  box-shadow: 0 0 0 var(--setup-progress-indicator-border-width) var(--setup-progress-indicator-color-current);
}

body.is-setup-assistant #app {
  grid-template-areas:
    "header"
    "col-full-width"
    "footer";
  grid-template-columns: 1fr;
}

@media (max-width: 1280px) {
  body.is-setup-assistant #app {
    grid-template-areas:
      "header"
      "col-full-width"
      "footer";
    grid-template-columns: 1fr;
    grid-template-rows: var(--header-height) 1fr var(--footer-height);
  }
}

.btn-bar-wide {
  margin-top: var(--spacer);
}

.btn-bar-wide .btn:last-child {
  margin-left: auto;
}

.setup-container {
  --color-check-failed: var(--color-danger);
  --color-check-ok: var(--color-success);

  max-width: 640px;
  margin: 10vh auto auto;
  width: 100%;
}

.setup-step {
  background-color: var(--color-background-column);
  border-radius: var(--border-radius);
  font-size: var(--font-size-lg);
  padding: var(--spacer) var(--spacer-lg);
}

.setup-step p {
  line-height: 1.4rem;
}

.setup-step a {
  text-decoration: underline;
}

input[type="text"] {
  font-size: var(--font-size-lg);
}

h1 {
  display: block;
  font-size: xx-large;
  margin: 0 auto var(--spacer-lg);
  padding-bottom: var(--spacer);
  text-align: center;
}

h2 {
  color: var(--color-text-highlight);
  margin-top: var(--spacer-sm);
}

.setup-container img {
  max-width: 100%;
  border-radius: var(--border-radius);
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

fieldset input[type="text"],
.setup-path-command {
  margin-left: 1.66rem;
  margin-top: var(--spacer-sm);
  width: -moz-available;
  width: -webkit-fill-available;
}

.setup-path-command {
  align-items: center;
  background-color: var(--color-background-column);
  border-radius: var(--border-radius);
  display: flex;
  font-size: var(--font-size-base);
  padding: var(--spacer-sm);
}

.setup-path-command .path {
  font-family: var(--font-family-mono);
  font-weight: bold;
  word-break: break-word;
}

.setup-path-command .command-preview {
  color: var(--color-text-muted);
  cursor: help;
  margin-left: auto;
}

fieldset label {
  margin: var(--spacer-lg) var(--spacer-sm);
}

fieldset .check-failed {
  margin-left: 1.66rem;
}
</style>
