<template>
  <header>
    <router-link :to="{ name: 'index' }" class="navbar-brand">{{ flamencoName }}</router-link>
    <nav></nav>
    <api-spinner />
    <span class="app-version">
      <a href="/api/v3/swagger-ui/">API</a>
      | version: {{ flamencoVersion }}
    </span>
  </header>
  <router-view></router-view>
</template>

<script>

const DEFAULT_FLAMENCO_NAME = "Flamenco";
const DEFAULT_FLAMENCO_VERSION = "unknown";
import ApiSpinner from '@/components/ApiSpinner.vue'
import { MetaApi } from "@/manager-api";
import { apiClient } from '@/stores/api-query-count';

export default {
  name: 'FirstTimeWizard',
  components: {
    ApiSpinner,
  },
  data: () => ({
    flamencoName: DEFAULT_FLAMENCO_NAME,
    flamencoVersion: DEFAULT_FLAMENCO_VERSION,
  }),
  mounted() {
    window.app = this;
    this.fetchManagerInfo();
  },
  methods: {
    // TODO: also call this when SocketIO reconnects.
    fetchManagerInfo() {
      const metaAPI = new MetaApi(apiClient);
      metaAPI.getVersion().then((version) => {
        this.flamencoName = version.name;
        this.flamencoVersion = version.version;
      })
    },
  },
}
</script>

<style>
@import "assets/base.css";
@import "assets/tabulator.css";
</style>
