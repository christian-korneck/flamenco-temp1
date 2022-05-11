<template>
  <div class="col col-1">
    <router-link :to="{ name: 'jobs' }"><span>J</span></router-link>
  </div>
  <div class="col col-2">
    <router-link :to="{ name: 'workers' }"><span>W</span></router-link>
  </div>
  <div class="col col-3">
    <router-link :to="{ name: 'settings' }"><span>S</span></router-link>
  </div>
  <footer>
    Make your choice.
  </footer>
</template>

<script>
export default {
  name: 'IndexView',
  components: {},
}
</script>

<style scoped>
.col a {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;

  transition: 500ms;
  background-color: var(--color-background-column);
}

.col a:hover {
  text-decoration: none;
  background-color: var(--color-accent-background);
}

.col a span {
  font-weight: bold;
  font-size: 10rem;
  text-align: center;
  vertical-align: middle;
}
</style>
