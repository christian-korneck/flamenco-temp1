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

import "bootstrap/dist/css/bootstrap.css";
import "bootstrap-vue/dist/bootstrap-vue.css";

let flamencoManager = require('flamenco-manager');

let apiClient = new flamencoManager.ApiClient("http://localhost:8080/");
var api = new flamencoManager.JobsApi(apiClient);
var jobId = "07d134bc-0614-4477-9b1f-e238f0f0391a";
api.fetchJob(jobId).then(function(data) {
  console.log('API called successfully. Returned data: ', data);
}, function(error) {
  console.error(error);
});

Vue.config.productionTip = false

Vue.use(FormInputPlugin);
Vue.use(NavbarPlugin);
Vue.use(LayoutPlugin);
Vue.component("b-card", BCard);
Vue.component("b-input-group", BInputGroup);
Vue.component("b-button", BButton);
Vue.use(IconsPlugin);

var app = new Vue({
  render: h => h(App),

});

app.$mount("#app");
