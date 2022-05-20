<script setup>
import { onMounted, watch } from 'vue'
import { TabulatorFull as Tabulator } from 'tabulator-tables';
import { useTaskLog } from '@/stores/tasklog'
import { useTasks } from '@/stores/tasks'

const taskLog = useTaskLog();
const tasks = useTasks();

const tabOptions = {
  columns: [
    {
      title: 'Log Lines',
      field: 'line',
      sorter: 'string',
      widthGrow: 100,
      resizable: true,
    },
  ],
  headerVisible: false,
  layout: "fitDataStretch",
  resizableColumnFit: true,
  height: "calc(20vh - 3rem)", // Must be set in order for the virtual DOM to function correctly.
  data: taskLog.history,
  placeholder: "Task log will appear here",
  selectable: false,
};

let tabulator = null;
onMounted(() => {
  tabulator = new Tabulator('#task_log_list', tabOptions);
  tabulator.on("tableBuilt", _scrollToBottom);
  tabulator.on("tableBuilt", _subscribeToPinia);
  console.log("Task log list: mounted on task ID", tasks.activeTaskID);
});

tasks.$subscribe((_, state) => {
  console.log("Task log list: new task ID", state.activeTaskID);
});


function _scrollToBottom() {
  if (taskLog.empty) return;
  tabulator.scrollToRow(taskLog.lastID, "bottom", false);
}
function _subscribeToPinia() {
  taskLog.$subscribe(() => {
    tabulator.setData(taskLog.history)
      .then(_scrollToBottom)
  })
}
</script>

<template>
    <div id="task_log_list"></div>
</template>
