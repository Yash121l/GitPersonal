import { createRouter, createWebHistory } from 'vue-router'
import type { RouteLocationGeneric } from 'vue-router'

import { bootstrap } from '@/app/bootstrap'
import type { FeatureFlag } from '@/app/features'
import { getDefaultRepositoryRoute, getDefaultWorkspaceRoute, isFeatureEnabled } from '@/app/navigation'
import { pinia } from '@/app/pinia'
import { useAuthStore } from '@/stores/auth'

const repositoryRoutes = [
  {
    path: '',
    redirect: (to: RouteLocationGeneric) => getDefaultRepositoryRoute(String(to.params.owner), String(to.params.repo), to.query),
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
    path: 'issues',
    name: 'repository-issues',
    component: () => import('@/views/RepositoryPlaceholderView.vue'),
    meta: {
      title: 'Forge | Issues',
      feature: 'repositoryAccess',
    },
  },
  {
    path: 'issues/new',
    name: 'repository-issues-new',
    component: () => import('@/views/RepositoryPlaceholderView.vue'),
    meta: {
      title: 'Forge | New Issue',
      feature: 'repositoryAccess',
    },
  },
  {
    path: 'issues/:number',
    name: 'repository-issue-detail',
    component: () => import('@/views/RepositoryPlaceholderView.vue'),
    meta: {
      title: 'Forge | Issue',
      feature: 'repositoryAccess',
    },
  },
  {
    path: 'pulls',
    name: 'repository-pulls',
    component: () => import('@/views/RepositoryPlaceholderView.vue'),
    meta: {
      title: 'Forge | Pull Requests',
      feature: 'repositoryAccess',
    },
  },
  {
    path: 'pulls/new',
    name: 'repository-pulls-new',
    component: () => import('@/views/RepositoryPlaceholderView.vue'),
    meta: {
      title: 'Forge | New Pull Request',
      feature: 'repositoryAccess',
    },
  },
  {
    path: 'pull/:number/:view(files|commits)?',
    name: 'repository-pull-detail',
    component: () => import('@/views/RepositoryPlaceholderView.vue'),
    meta: {
      title: 'Forge | Pull Request',
      feature: 'repositoryAccess',
    },
  },
  {
    path: 'actions/:rest(.*)*',
    name: 'repository-actions',
    component: () => import('@/views/RepositoryAutomationView.vue'),
    meta: {
      title: 'Forge | Actions',
      feature: 'repositoryAutomation',
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
    path: 'insights/:section?',
    name: 'repository-insights',
    component: () => import('@/views/RepositoryActivityView.vue'),
    meta: {
      title: 'Forge | Insights',
      feature: 'repositoryActivity',
    },
  },
  {
    path: 'projects',
    name: 'repository-projects',
    component: () => import('@/views/RepositoryPlaceholderView.vue'),
    meta: {
      title: 'Forge | Projects',
      feature: 'repositories',
    },
  },
  {
    path: 'wiki',
    name: 'repository-wiki',
    component: () => import('@/views/RepositoryPlaceholderView.vue'),
    meta: {
      title: 'Forge | Wiki',
      feature: 'repositories',
    },
  },
  {
    path: 'security/:section?',
    name: 'repository-security',
    component: () => import('@/views/RepositoryPlaceholderView.vue'),
    meta: {
      title: 'Forge | Security',
      feature: 'repositoryAccess',
    },
  },
  {
    path: 'releases/:section?',
    name: 'repository-releases',
    component: () => import('@/views/RepositoryPlaceholderView.vue'),
    meta: {
      title: 'Forge | Releases',
      feature: 'repositories',
    },
  },
  {
    path: 'tags',
    name: 'repository-tags',
    component: () => import('@/views/RepositoryPlaceholderView.vue'),
    meta: {
      title: 'Forge | Tags',
      feature: 'repositories',
    },
  },
  {
    path: 'packages',
    name: 'repository-packages',
    component: () => import('@/views/RepositoryPlaceholderView.vue'),
    meta: {
      title: 'Forge | Packages',
      feature: 'repositories',
    },
  },
  {
    path: 'discussions',
    name: 'repository-discussions',
    component: () => import('@/views/RepositoryPlaceholderView.vue'),
    meta: {
      title: 'Forge | Discussions',
      feature: 'repositories',
    },
  },
  {
    path: 'settings',
    name: 'repository-settings',
    component: () => import('@/views/RepositorySettingsView.vue'),
    meta: {
      title: 'Forge | Repository Settings',
      feature: 'repositoryAccess',
    },
  },
  {
    path: 'settings/:section(.*)*',
    name: 'repository-settings-section',
    component: () => import('@/views/RepositorySettingsView.vue'),
    meta: {
      title: 'Forge | Repository Settings',
      feature: 'repositoryAccess',
    },
  },
]

const legacyRepositoryRedirects = [
  { path: 'repos/:owner/:repo', section: undefined },
  { path: 'repos/:owner/:repo/code', section: 'code' },
  { path: 'repos/:owner/:repo/access', section: 'access' },
  { path: 'repos/:owner/:repo/automation', section: 'automation' },
  { path: 'repos/:owner/:repo/activity', section: 'activity' },
  { path: 'repos/:owner/:repo/settings', section: 'settings' },
].map(({ path, section }) => ({
  path,
  redirect: (to: RouteLocationGeneric) => ({
    name: section ? `repository-${section}` : 'repository-code',
    params: {
      owner: to.params.owner,
      repo: to.params.repo,
    },
    query: to.query,
  }),
}))

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
      alias: '/signup',
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
          path: 'explore/:section?',
          name: 'explore',
          component: () => import('@/views/ExploreView.vue'),
          meta: {
            title: 'Forge | Explore',
            feature: 'repositories',
          },
        },
        {
          path: 'new',
          name: 'new-repository',
          component: () => import('@/views/NewRepositoryView.vue'),
          meta: {
            title: 'Forge | New Repository',
            feature: 'repositories',
          },
        },
        {
          path: 'organizations/new',
          name: 'new-organization',
          component: () => import('@/views/NewRepositoryView.vue'),
          meta: {
            title: 'Forge | New Organization',
            feature: 'organizations',
          },
        },
        {
          path: 'settings',
          name: 'account-settings',
          component: () => import('@/views/AccountSettingsView.vue'),
          meta: {
            title: 'Forge | Settings',
            feature: 'workspaceOverview',
          },
        },
        {
          path: 'settings/:section',
          name: 'account-settings-section',
          component: () => import('@/views/AccountSettingsView.vue'),
          meta: {
            title: 'Forge | Settings',
            feature: 'workspaceOverview',
          },
        },
        {
          path: 'marketplace',
          redirect: { name: 'explore' },
        },
        {
          path: 'codespaces',
          component: () => import('@/views/ExploreView.vue'),
          meta: {
            title: 'Forge | Codespaces',
            feature: 'repositories',
          },
        },
        ...legacyRepositoryRedirects,
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
        {
          path: ':owner/:repo',
          component: () => import('@/views/RepositoryLayoutView.vue'),
          meta: {
            title: 'Forge | Repository',
            feature: 'repositories',
          },
          children: repositoryRoutes,
        },
        {
          path: ':username',
          name: 'profile',
          component: () => import('@/views/ProfileView.vue'),
          meta: {
            title: 'Forge | Profile',
            feature: 'workspaceOverview',
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
