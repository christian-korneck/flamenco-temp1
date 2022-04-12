import { createApp } from 'vue'

import App from './App.vue'
import router from './router'

// Ensure Tabulator can find `luxon`, which it needs for sorting by
// date/time/datetime.
import { DateTime } from 'luxon';
window.DateTime = DateTime;

// // Example of how to use the Flamenco Manager API:
// import * as urls from './urls'
// import {
//   ApiClient,
//   MetaApi,
// } from './manager-api'

// let apiClient = new ApiClient(urls.api());
// console.log('ApiClient:', apiClient);
// const metaAPI = new MetaApi(apiClient);
// console.log("metaAPI:", metaAPI);
// metaAPI.getVersion().then(function (version) {
//   console.log("version:", version);
// })
// metaAPI.getConfiguration().then(function (config) {
//   console.log("config:", config);
// })

const app = createApp(App)
app.use(router)
app.mount('#app')
