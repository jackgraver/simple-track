import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { authStatus } from '~/composables/auth/session'
import { resolveAuthSession } from '~/composables/auth/useAuth'

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
      {
        path: 'logging/cardio',
        name: 'logging-cardio',
        component: () => import('~/pages/gym/logging/exercise/[id].vue'),
      },
      {
        path: 'logging/mobility/:slot',
        name: 'logging-mobility',
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

router.beforeEach(async (to) => {
  if (to.name === 'signin') {
    if (authStatus.value === 'unknown') await resolveAuthSession()
    if (authStatus.value === 'authenticated') {
      return { name: 'gym' }
    }
    return true
  }
  if (authStatus.value === 'unknown') await resolveAuthSession()
  if (authStatus.value !== 'authenticated') {
    return { name: 'signin', query: { redirect: to.fullPath } }
  }
  return true
})

