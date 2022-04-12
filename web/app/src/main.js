import { createApp } from 'vue'

import App from './App.vue'
import router from './router'

// Ensure Tabulator can find `luxon`, which it needs for sorting by
// date/time/datetime.
import { DateTime } from 'luxon';
window.DateTime = DateTime;

const app = createApp(App)
app.use(router)
app.mount('#app')
