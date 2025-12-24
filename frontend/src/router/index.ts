import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('~/pages/home/home.vue'),
  },
  {
    path: '/diet',
    name: 'diet',
    component: () => import('~/pages/diet.vue'),
  },
  {
    path: '/gym',
    name: 'gym',
    component: () => import('~/pages/gym.vue'),
  },
  {
    path: '/settings',
    name: 'settings',
    component: () => import('~/pages/settings.vue'),
  },
  {
    path: '/progression',
    name: 'progression',
    component: () => import('~/pages/progression.vue'),
  },
  {
    path: '/manageplans',
    name: 'manageplans',
    component: () => import('~/pages/manageplans.vue'),
  },
  {
    path: '/logmeal',
    name: 'logmeal',
    component: () => import('~/pages/logmeal/logmeal.vue'),
  },
  {
    path: '/mealplan',
    name: 'mealplan',
    component: () => import('~/pages/mealplan.vue'),
  },
  {
    path: '/liveworkout',
    name: 'liveworkout',
    component: () => import('~/pages/liveworkout/index.vue'),
  },
  {
    path: '/liveworkout/log/:id(\\d+)',
    name: 'liveworkout-log',
    component: () => import('~/pages/liveworkout/log/[id].vue'),
  },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

