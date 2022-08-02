<template>
  <h2 class="column-title">Task Details</h2>

  <template v-if="hasTaskData">
    <dl>
      <dt class="field-id" title="ID">ID</dt>
      <dd>{{ taskData.id }}</dd>

      <dt class="field-name" title="Name">Name</dt>
      <dd>{{ taskData.name }}</dd>

      <dt class="field-status" title="Status">Status</dt>
      <dd class="field-status-label" :class="'status-' + taskData.status">{{ taskData.status }}</dd>

      <dt class="field-worker" title="Assigned To">Assigned To</dt>
      <dd>
        <link-worker v-if="taskData.worker" :worker="taskData.worker" />
        <template v-else>-</template>
      </dd>

      <template v-if="taskData.failed_by_workers.length > 0">
        <dt class="field-failed-by-workers" title="Failed by Workers">Failed by Workers</dt>
        <dd>
          <div v-for="worker in taskData.failed_by_workers">
            <link-worker :worker="worker" />
          </div>
        </dd>
      </template>

      <dt class="field-type" title="Task Type">Task Type</dt>
      <dd>{{ taskData.task_type }}</dd>

      <dt class="field-priority" title="Priority">Priority</dt>
      <dd>{{ taskData.priority }}</dd>

      <dt class="field-created" title="Created">Created</dt>
      <dd>{{ datetime.relativeTime(taskData.created) }}</dd>

      <dt class="field-updated" title="Updated">Updated</dt>
      <dd>{{ datetime.relativeTime(taskData.updated) }}</dd>

      <dt class="field-last-touched" title="Last Touched by Worker">Last Touched by Worker</dt>
      <dd>{{ datetime.relativeTime(taskData.last_touched) }}</dd>

      <dt class="field-activity" title="Activity">Activity</dt>
      <dd>{{ taskData.activity }}</dd>
    </dl>

    <h3 class="sub-title">Commands</h3>
    <dl>
      <template v-for="cmd in taskData.commands">
        <dt :class="`field-${cmd.name}`" :title="cmd.name">{{ cmd.name }}</dt>
        <dd>{{ cmd.parameters }}</dd>
      </template>
    </dl>

    <h3 class="sub-title">Task Log</h3>
    <div class="btn-bar-group">
      <section class="btn-bar tasklog">
        <button class="btn" @click="$emit('showTaskLogTail')" title="Open the task log tail in the footer.">
          Follow Task Log</button>
        <button class="btn" @click="openFullLog" title="Opens the task log in a new window.">Open Full Log</button>
      </section>
    </div>
  </template>

  <div v-else class="details-no-item-selected">
    <p>Select a task to see its details.</p>
  </div>
</template>

<script>
import * as datetime from "@/datetime";
import { JobsApi } from '@/manager-api';
import { backendURL } from '@/urls';
import { apiClient } from '@/stores/api-query-count';
import { useNotifs } from "@/stores/notifications";
import LinkWorker from '@/components/LinkWorker.vue';

export default {
  props: [
    "taskData", // Task data to show.
  ],
  emits: [
    "showTaskLogTail", // Emitted when the user presses the "follow task log" button.
  ],
  components: { LinkWorker },
  data() {
    return {
      datetime: datetime, // So that the template can access it.
      jobsApi: new JobsApi(apiClient),
      notifs: useNotifs(),
    };
  },
  mounted() {
    // Allow testing from the JS console:
    window.taskDetailsVue = this;
  },
  computed: {
    hasTaskData() {
      return !!this.taskData && !!this.taskData.id;
    },
  },
  methods: {
    openFullLog() {
      const taskUUID = this.taskData.id;

      this.jobsApi.fetchTaskLogInfo(taskUUID)
        .then((logInfo) => {
          if (logInfo == null) {
            this.notifs.add(`Task ${taskUUID} has no log yet`)
            return;
          }
          console.log(`task ${taskUUID} log info:`, logInfo);

          const url = backendURL(logInfo.url);
          window.open(url, "_blank");
        })
        .catch((error) => {
          console.log(`Error fetching task ${taskUUID} log info:`, error);
        })
    },
  },
};
</script>

<style scoped>
/* Prevent fields with long IDs from overflowing. */
.field-id+dd {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.field-status-label {
  color: var(--indicator-color);
  font-weight: bold;
}
</style>
