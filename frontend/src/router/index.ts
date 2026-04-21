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
    component: () => import('~/pages/home/index.vue'),
  },
  {
    path: '/diet',
    name: 'diet',
    component: () => import('~/pages/diet/index.vue'),
    children: [
      {
        path: 'log',
        name: 'diet-log',
        component: () => import('~/pages/diet/logmeal/logmeal.vue'),
      },
      {
        path: 'edit-planned',
        name: 'diet-edit-planned',
        component: () => import('~/pages/diet/edit-planned/index.vue'),
      },
    ]
  },
  {
    path: '/gym',
    name: 'gym',
    component: () => import('~/pages/gym/index.vue'),
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
      {
        path: 'progression',
        name: 'progression',
        component: () => import('~/pages/gym/progression/progression.vue'),
      },
      {
        path: 'weight',
        name: 'gym-weight',
        component: () => import('~/pages/gym/weight/index.vue'),
      },
      {
        path: 'steps',
        name: 'gym-steps',
        component: () => import('~/pages/gym/steps/index.vue'),
      },
    ],
  },
  {
    path: '/logmeal',
    name: 'logmeal',
    component: () => import('~/pages/diet/logmeal/logmeal.vue'),
  },

  {
    path: '/auth/signin',
    name: 'signin',
    component: () => import('~/pages/auth/signin/index.vue'),
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

