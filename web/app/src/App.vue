<template>
  <header>
    <a href="/" class="navbar-brand">{{ flamencoName }}</a>
    <nav>
      <ul>
        <li>
          <router-link to="/">Jobs</router-link>
        </li>
        <li>
          <router-link to="/workers">Workers</router-link>
        </li>
        <li>
          <router-link to="/settings">Settings</router-link>
        </li>
      </ul>
    </nav>
    <api-spinner />
    <span class="app-version">
      version: {{ flamencoVersion }}
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
