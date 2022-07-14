import { createApp } from 'vue'
import { createPinia } from 'pinia'

import FirstTimeWizard from '@/FirstTimeWizard.vue'
import router from '@/router/first-time-wizard'

// Ensure Tabulator can find `luxon`, which it needs for sorting by
// date/time/datetime.
import { DateTime } from 'luxon';
window.DateTime = DateTime;

// plain removes any Vue reactivity.
window.plain = (x) => JSON.parse(JSON.stringify(x));
// objectEmpty returns whether the object is empty or not.
window.objectEmpty = (o) => !o || Object.entries(o).length == 0;

const app = createApp(FirstTimeWizard)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.mount('#app')

// Automatically reload the window after a period of inactivity from the user.
import autoreload from '@/autoreloader'
autoreload();
