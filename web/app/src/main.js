import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from '@/App.vue'

// Ensure Tabulator can find `luxon`, which it needs for sorting by
// date/time/datetime.
import { DateTime } from 'luxon';
window.DateTime = DateTime;

// Help with debugging. This removes any Vue reactivity.
window.plain = (x) => { return JSON.parse(JSON.stringify(x)) };

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.mount('#app')
