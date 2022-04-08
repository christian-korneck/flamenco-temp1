import Vue from 'vue'
import App from './App.vue'
import {
  FormInputPlugin,
  NavbarPlugin,
  LayoutPlugin,
  IconsPlugin,
  BCard,
  BInputGroup,
  BButton,
} from "bootstrap-vue";

import URLs from './urls'

import "bootstrap/dist/css/bootstrap.css";
import "bootstrap-vue/dist/bootstrap-vue.css";

let flamencoManager = require('flamenco-manager');
let apiClient = new flamencoManager.ApiClient(URLs.api);

// Ensure Tabulator can find `moment`, which it needs for sorting by
// date/time/datetime.
import moment from 'moment';
window.moment = moment;

// let query = new flamencoManager.JobsQuery();
// // query.status_in = ["active"];
// query.metadata = {project: "Heist"};

// let JobsApi = new flamencoManager.JobsApi(apiClient);
// JobsApi.queryJobs(query).then(function(data) {
//   console.log('API called successfully.');
//   console.log(data);
// }, function(error) {
//   console.error(error);
// });

Vue.config.productionTip = false

Vue.use(FormInputPlugin);
Vue.use(NavbarPlugin);
Vue.use(LayoutPlugin);
Vue.component("b-card", BCard);
Vue.component("b-input-group", BInputGroup);
Vue.component("b-button", BButton);
Vue.use(IconsPlugin);

var vueApp = new Vue(App);
vueApp.websocketURL = URLs.ws;
vueApp.apiClient = apiClient;

vueApp.$mount("#app");
