<template>
  <div class="flamenco-jobs-table">
    <JobsTabulator v-model="jobs" :options="options" />
  </div>
</template>

<script>
import { TabulatorComponent } from "vue-tabulator";

export default {
  props: ["apiClient"],
  components: {
    JobsTabulator: TabulatorComponent,
  },
  data: () => {
    const options = {
      columns: [
        { title: 'Name', field: 'name', sorter: 'string', width: 200, editor: true },
        { title: 'Age', field: 'age', sorter: 'int', width: 200, editor: true }
      ]
    };

    return {
      jobs: [
        // To be loaded from an API call instead.
        { name: 'Teste', age: 13 }
      ],
      options: options,
    };
  },
};
</script>

<style lang='scss'>
@import "~vue-tabulator/dist/scss/bootstrap/tabulator_bootstrap4";

.flamenco-jobs-table {
  min-height: 50%;
  border: thick solid fuchsia;
}
</style>
