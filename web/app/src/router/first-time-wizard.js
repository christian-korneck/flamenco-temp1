import { createRouter, createWebHistory } from "vue-router";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "index",
      component: () => import("../views/FirstTimeWizardView.vue"),
    },
    {
      path: "/:pathMatch(.*)*",
      name: "redirect-to-index",
      redirect: '/',
    },
  ],
});

export default router;
