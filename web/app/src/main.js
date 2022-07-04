import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from '@/App.vue'
import router from '@/router/index'

// Ensure Tabulator can find `luxon`, which it needs for sorting by
// date/time/datetime.
import { DateTime } from 'luxon';
window.DateTime = DateTime;

// plain removes any Vue reactivity.
window.plain = (x) => JSON.parse(JSON.stringify(x));
// objectEmpty returns whether the object is empty or not.
window.objectEmpty = (o) => !o || Object.entries(o).length == 0;

// Do a full refresh once per hour. This is just to make sure that long-lived
// displays (like the TV in the hallway at Blender HQ) pick up on HTML/JS/CSS
// changes eventually.
window.setTimeout(window.location.reload, 3600 * 1000);

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.mount('#app')

// For debugging.
import { useJobs } from '@/stores/jobs';
import { useNotifs } from '@/stores/notifications';
import { useTaskLog } from '@/stores/tasklog';
import * as API from '@/manager-api';
window.jobs = useJobs();
window.notifs = useNotifs();
window.taskLog = useTaskLog();
window.API = API;
