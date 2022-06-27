<template>
  <header>
    <router-link :to="{ name: 'index' }" class="navbar-brand">{{ flamencoName }}</router-link>
    <nav>
      <ul>
        <li>
          <router-link :to="{ name: 'jobs' }">Jobs</router-link>
        </li>
        <li>
          <router-link :to="{ name: 'workers' }">Workers</router-link>
        </li>
        <li>
          <router-link :to="{ name: 'settings' }">Settings</router-link>
        </li>
      </ul>
    </nav>
    <api-spinner />
    <span class="app-version">
      <a href="/flamenco-addon.zip">addon</a>
      | <a href="/api/swagger-ui/">API</a>
      | version: {{ flamencoVersion }}
    </span>
  </header>
  <router-view></router-view>
</template>

<script>
import * as API from '@/manager-api';
import { apiClient } from '@/stores/api-query-count';

import ApiSpinner from '@/components/ApiSpinner.vue'

const DEFAULT_FLAMENCO_NAME = "Flamenco";
const DEFAULT_FLAMENCO_VERSION = "unknown";

export default {
  name: 'App',
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
      const metaAPI = new API.MetaApi(apiClient);
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
