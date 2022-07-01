import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'index',
      redirect: { name: 'jobs' },
    },
    {
      path: '/jobs/:jobID?/:taskID?',
      name: 'jobs',
      component: () => import('../views/JobsView.vue'),
      props: true,
    },
    {
      path: '/workers/:workerID?',
      name: 'workers',
      component: () => import('../views/WorkersView.vue'),
      props: true,
    },
    {
      path: '/last-rendered',
      name: 'last-rendered',
      component: () => import('../views/LastRenderedView.vue'),
    },
  ],
})

export default router
