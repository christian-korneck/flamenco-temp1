import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'index',
      component: () => import('../views/IndexView.vue'),
    },
    {
      path: '/jobs/:jobID?',
      name: 'jobs',
      component: () => import('../views/JobsView.vue'),
      props: true,
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
    },
  ],
})

export default router
