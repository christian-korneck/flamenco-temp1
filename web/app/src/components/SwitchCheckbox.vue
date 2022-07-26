<template>
  <label>
    <span class="switch">
      <input type="checkbox" :checked="isChecked" @change="$emit('switchToggle')">
      <span class="slider round"></span>
    </span>
    <span class="switch-label">{{ label }}</span>
  </label>
</template>

<script setup>
const props = defineProps(['isChecked', 'label']);
</script>

<style scoped>

label {
  display: inline-block;
}
.switch {
  position: relative;
  display: inline-block;
  width: 32px;
  height: 14px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--color-text-muted);
  -webkit-transition: .4s;
  transition: .4s;
}

.slider:before {
  position: absolute;
  content: "";
  height: 10px;
  width: 10px;
  left: 4px;
  bottom: 2px;
  background-color: white;
  -webkit-transition: .4s;
  transition: .4s;
}

input:checked + .slider {
  background-color: var(--color-accent);
}

input:focus + .slider {
  box-shadow: 0 0 1px var(--color-accent);
}

input:checked + .slider:before {
  -webkit-transform: translateX(14px);
  -ms-transform: translateX(14px);
  transform: translateX(14px);
}

/* Rounded sliders */
.slider.round {
  border-radius: 10px;
}

.slider.round:before {
  border-radius: 50%;
}
.switch-label {
  margin-left: 0.5rem;
  cursor: pointer;
}

</style>
