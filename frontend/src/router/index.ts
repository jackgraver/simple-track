import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: { name: 'gym' },
  },
  {
    path: '/home',
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
    component: () => import('~/pages/gym/gym.vue'),
    children: [
      {
        path: 'plans',
        name: 'gym-plans',
        component: () => import('~/pages/gym/plans/index.vue'),
      },
      {
        path: 'logging',
        name: 'logging',
        component: () => import('~/pages/gym/logging/index.vue'),
      },
      {
        path: 'logging/:id(\\d+)',
        name: 'logging-exercise',
        component: () => import('~/pages/gym/logging/exercise/[id].vue'),
      },
    ],
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
    redirect: { name: 'gym-plans' },
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
    path: '/signin',
    name: 'signin',
    component: () => import('~/pages/signin.vue'),
  },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

