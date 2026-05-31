import { createRouter, createWebHistory } from 'vue-router'

import { bootstrap } from '@/app/bootstrap'
import type { FeatureFlag } from '@/app/features'
import { getDefaultRepositoryRoute, getDefaultWorkspaceRoute, isFeatureEnabled } from '@/app/navigation'
import { pinia } from '@/app/pinia'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(bootstrap.basePath),
  scrollBehavior() {
    return { top: 0 }
  },
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: {
        guestOnly: true,
        title: 'Forge | Sign In',
      },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/RegisterView.vue'),
      meta: {
        guestOnly: true,
        title: 'Forge | Create Account',
      },
    },
    {
      path: '/',
      component: () => import('@/views/AppLayoutView.vue'),
      meta: {
        requiresAuth: true,
      },
      children: [
        {
          path: '',
          redirect: () => getDefaultWorkspaceRoute(),
        },
        {
          path: 'overview',
          name: 'overview',
          component: () => import('@/views/OverviewView.vue'),
          meta: {
            title: 'Forge | Overview',
            feature: 'workspaceOverview',
          },
        },
        {
          path: 'repos',
          name: 'repositories',
          component: () => import('@/views/RepositoriesView.vue'),
          meta: {
            title: 'Forge | Repositories',
            feature: 'repositories',
          },
        },
        {
          path: 'repos/:owner/:repo',
          component: () => import('@/views/RepositoryLayoutView.vue'),
          meta: {
            title: 'Forge | Repository',
            feature: 'repositories',
          },
          children: [
            {
              path: '',
              redirect: (to) => getDefaultRepositoryRoute(String(to.params.owner), String(to.params.repo), to.query),
            },
            {
              path: 'code',
              name: 'repository-code',
              component: () => import('@/views/RepositoryCodeView.vue'),
              meta: {
                title: 'Forge | Repository Code',
                feature: 'repositoryCode',
              },
            },
            {
              path: 'access',
              name: 'repository-access',
              component: () => import('@/views/RepositoryAccessView.vue'),
              meta: {
                title: 'Forge | Repository Access',
                feature: 'repositoryAccess',
              },
            },
            {
              path: 'automation',
              name: 'repository-automation',
              component: () => import('@/views/RepositoryAutomationView.vue'),
              meta: {
                title: 'Forge | Repository Automation',
                feature: 'repositoryAutomation',
              },
            },
            {
              path: 'activity',
              name: 'repository-activity',
              component: () => import('@/views/RepositoryActivityView.vue'),
              meta: {
                title: 'Forge | Repository Activity',
                feature: 'repositoryActivity',
              },
            },
            {
              path: 'settings',
              name: 'repository-settings',
              component: () => import('@/views/RepositorySettingsView.vue'),
              meta: {
                title: 'Forge | Repository Settings',
                feature: 'repositorySettings',
              },
            },
          ],
        },
        {
          path: 'orgs',
          name: 'organizations',
          component: () => import('@/views/OrganizationsView.vue'),
          meta: {
            title: 'Forge | Organizations',
            feature: 'organizations',
          },
        },
        {
          path: 'keys',
          name: 'keys',
          component: () => import('@/views/KeysView.vue'),
          meta: {
            title: 'Forge | SSH Keys',
            feature: 'sshKeys',
          },
        },
      ],
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: () => getDefaultWorkspaceRoute(),
    },
  ],
})

router.beforeEach(async (to) => {
  const authStore = useAuthStore(pinia)
  const requiresAuth = to.matched.some((record) => Boolean(record.meta.requiresAuth))
  const guestOnly = to.matched.some((record) => Boolean(record.meta.guestOnly))

  if ((requiresAuth || guestOnly) && !authStore.initialized) {
    await authStore.ensureLoaded()
  }

  if (requiresAuth && !authStore.isAuthenticated) {
    return {
      name: 'login',
      query: to.fullPath ? { redirect: to.fullPath } : undefined,
    }
  }

  const disabledFeature = to.matched.find((record) => {
    if (typeof record.meta.feature !== 'string') {
      return false
    }
    return !isFeatureEnabled(record.meta.feature as FeatureFlag)
  })

  if (disabledFeature) {
    if (typeof to.params.owner === 'string' && typeof to.params.repo === 'string') {
      return getDefaultRepositoryRoute(String(to.params.owner), String(to.params.repo), to.query)
    }
    return getDefaultWorkspaceRoute()
  }

  if (guestOnly && authStore.isAuthenticated) {
    return getDefaultWorkspaceRoute()
  }

  if (typeof to.meta.title === 'string') {
    document.title = to.meta.title
  }

  return true
})

export default router
