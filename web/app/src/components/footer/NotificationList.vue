<script setup>
import { onMounted } from 'vue'
import { TabulatorFull as Tabulator } from 'tabulator-tables';
import { useNotifs } from '@/stores/notifications'
import * as datetime from "@/datetime";

const notifs = useNotifs();

const tabOptions = {
  columns: [
    {
      title: 'Time', field: 'time',
      sorter: 'alphanum', sorterParams: { alignEmptyValues: "top" },
      formatter(cell) {
        const cellValue = cell.getData().time;
        return datetime.shortened(cellValue);
      },
      widthGrow: 1,
      resizable: true,
    },
    {
      title: 'Message',
      field: 'msg',
      sorter: 'string',
      widthGrow: 100,
      resizable: true,
    },
  ],
  initialSort: [
    { column: "time", dir: "asc" },
  ],
  headerVisible: false,
  layout: "fitDataStretch",
  resizableColumnFit: true,
  height: "calc(25vh - 3rem)", // Must be set in order for the virtual DOM to function correctly.
  data: notifs.history,
  placeholder: "Notification history will appear here",
  selectable: false,
};

let tabulator = null;
onMounted(() => {
  tabulator = new Tabulator('#notification_list', tabOptions);
  tabulator.on("tableBuilt", _scrollToBottom);
  tabulator.on("tableBuilt", _subscribeToPinia);
});

function _scrollToBottom() {
  if (notifs.empty) return;
  tabulator.scrollToRow(notifs.lastID, "bottom", false);
}
function _subscribeToPinia() {
  notifs.$subscribe(() => {
    tabulator.setData(notifs.history)
      .then(_scrollToBottom)
  })
}
</script>

<template>
    <div id="notification_list"></div>
</template>
