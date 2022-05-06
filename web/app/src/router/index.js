import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'jobs',
      // route level code-splitting
      // this generates a separate chunk (ViewName.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: () => import('../views/JobsView.vue')
    },
    {
      path: '/workers',
      name: 'workers',
      component: () => import('../views/WorkersView.vue')
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('../views/SettingsView.vue')
    }
  ]
})

export default router
