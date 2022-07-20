<script setup>
import { useSlots, ref, provide } from 'vue';
const slots = useSlots();

const stepTitles = ref(slots.default().map((step) => step.props.title));
const selectedStep = ref(stepTitles.value[0]);
provide('selectedStep', selectedStep);

function updateStepTitle(title) {
  selectedStep.value = title;
}

function selectedStepNumber() {
  return stepTitles.value.indexOf(selectedStep.value);
}

function nextStepTitle() {
  selectedStep.value = stepTitles.value[selectedStepNumber() + 1];
}

function prevStepTitle() {
  selectedStep.value = stepTitles.value[selectedStepNumber() - 1];
}

</script>

<template>
  <div class="setup-step">
    <h2>{{ selectedStep }}</h2>
    <nav>
      <ul class="progress">
        <li
          v-for="(title, index) in stepTitles"
          :class="{current: selectedStep == title, done: index < selectedStepNumber() }"
          @click="updateStepTitle(title)"
          >
          <span></span>
        </li>
      </ul>
    </nav>
    <slot />
    <div
      v-if="selectedStepNumber() != stepTitles.length - 1"
      class="btn-bar btn-bar-wide">
      <button
        v-if="selectedStepNumber() > 0"
        @click="prevStepTitle()"
        class="btn">Back
      </button>
      <button
        @click="nextStepTitle()"
        class="btn btn-primary">Next</button>
    </div>
  </div>
</template>

<style scoped>


.progress {
  --wiz-progress-indicator-size: 8px;
  --wiz-progress-indicator-border-width: 2px;
  --wiz-progress-indicator-color: var(--color-text-hint);
  --wiz-progress-indicator-color-current: var(--color-accent);

  display: flex;
  justify-content: space-between;
  list-style: none;
  margin-bottom: 2rem;
  padding: 0;
  position: relative;

}
.progress li {
  cursor: pointer;
}

/* Progress indicator dot.  */
.progress li span {
  background-color: var(--color-background-column);
  border-radius: 50%;
  border: var(--wiz-progress-indicator-border-width) solid var(--color-background-column);
  box-shadow: 0 0 0 var(--wiz-progress-indicator-border-width) var(--wiz-progress-indicator-color);
  content: '';
  cursor: pointer;
  display: block;
  height: var(--wiz-progress-indicator-size);
  position: relative;
  width: var(--wiz-progress-indicator-size);
}

.progress li.done span {
  background-color: var(--wiz-progress-indicator-color-current);
  box-shadow: 0 0 0 var(--wiz-progress-indicator-border-width) var(--wiz-progress-indicator-color-current);
}

.progress li.current span {
  background-color: var(--color-background-column);
  box-shadow: 0 0 0 var(--wiz-progress-indicator-border-width) var(--wiz-progress-indicator-color-current);
}

.progress li.current span {
  box-shadow: 0 0 0 var(--wiz-progress-indicator-border-width) var(--wiz-progress-indicator-color-current);
}


/* Progress indicator line between dots. */
.progress:before {
  background-color: var(--wiz-progress-indicator-color);
  content: '';
  display: block;
  height: var(--wiz-progress-indicator-border-width);
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 100%;
}

.setup-step {
  background-color: var(--color-background-column);
  border-radius: var(--border-radius);
  padding: var(--spacer) var(--spacer-lg);
}

</style>
