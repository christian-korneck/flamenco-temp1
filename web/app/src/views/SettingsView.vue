<template>
  <div class="col col-1">
    Settings
  </div>
  <div class="col col-2">
    More settings
  </div>
  <div class="col col-3">
    Completely different settings
  </div>
  <footer>
    No footer here.
  </footer>
</template>

<script>
export default {
  name: 'SettingsView',
  components: {},
  data: () => ({
  }),
  mounted() {
    window.settingsView = this;
  },
  methods: {
  },
}
</script>
