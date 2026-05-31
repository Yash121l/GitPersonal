import {
  Activity,
  FolderGit2,
  Gauge,
  KeyRound,
  Landmark,
  Settings2,
  ShieldCheck,
  Zap,
} from '@lucide/vue'
import type { Component } from 'vue'
import type { LocationQuery, RouteLocationRaw, RouteLocationNormalizedLoaded } from 'vue-router'

import { bootstrap } from '@/app/bootstrap'
import type { FeatureFlag } from '@/app/features'

export type AppRouteName =
  | 'overview'
  | 'repositories'
  | 'organizations'
  | 'keys'
  | 'repository-code'
  | 'repository-access'
  | 'repository-automation'
  | 'repository-activity'
  | 'repository-settings'

interface BaseNavigationItem {
  id: string
  label: string
  description: string
  icon: Component
  feature: FeatureFlag
  matchRouteNames: string[]
}

export interface WorkspaceNavigationItem extends BaseNavigationItem {
  group: 'workspace' | 'collaboration'
  to: RouteLocationRaw
  routeName: AppRouteName
}

export interface RepositoryContext {
  owner: string
  repo: string
  currentQuery?: LocationQuery
}

export interface RepositoryNavigationItem extends BaseNavigationItem {
  to: (context: RepositoryContext) => RouteLocationRaw
  routeName: AppRouteName
}

const workspaceItems: WorkspaceNavigationItem[] = [
  {
    id: 'workspace-overview',
    label: 'Overview',
    description: 'Developer dashboard and quick links.',
    icon: Gauge,
    feature: 'workspaceOverview',
    routeName: 'overview',
    to: { name: 'overview' },
    group: 'workspace',
    matchRouteNames: ['overview'],
  },
  {
    id: 'workspace-repositories',
    label: 'Repositories',
    description: 'Browse and create repositories.',
    icon: FolderGit2,
    feature: 'repositories',
    routeName: 'repositories',
    to: { name: 'repositories' },
    group: 'workspace',
    matchRouteNames: ['repositories', 'repository-code', 'repository-access', 'repository-automation', 'repository-activity', 'repository-settings'],
  },
  {
    id: 'workspace-organizations',
    label: 'Organizations',
    description: 'Manage shared namespaces and memberships.',
    icon: Landmark,
    feature: 'organizations',
    routeName: 'organizations',
    to: { name: 'organizations' },
    group: 'collaboration',
    matchRouteNames: ['organizations'],
  },
  {
    id: 'workspace-keys',
    label: 'SSH Keys',
    description: 'Manage developer keys for Git transport.',
    icon: KeyRound,
    feature: 'sshKeys',
    routeName: 'keys',
    to: { name: 'keys' },
    group: 'collaboration',
    matchRouteNames: ['keys'],
  },
]

const repositoryItems: RepositoryNavigationItem[] = [
  {
    id: 'repo-code',
    label: 'Code',
    description: 'Tree browser, branches, and blob previews.',
    icon: FolderGit2,
    feature: 'repositoryCode',
    routeName: 'repository-code',
    matchRouteNames: ['repository-code'],
    to: ({ owner, repo, currentQuery }) => ({
      name: 'repository-code',
      params: { owner, repo },
      query: currentQuery,
    }),
  },
  {
    id: 'repo-access',
    label: 'Access',
    description: 'Clone endpoints, visibility, and collaborators.',
    icon: ShieldCheck,
    feature: 'repositoryAccess',
    routeName: 'repository-access',
    matchRouteNames: ['repository-access'],
    to: ({ owner, repo }) => ({
      name: 'repository-access',
      params: { owner, repo },
    }),
  },
  {
    id: 'repo-automation',
    label: 'Automation',
    description: 'Webhooks and event integrations.',
    icon: Zap,
    feature: 'repositoryAutomation',
    routeName: 'repository-automation',
    matchRouteNames: ['repository-automation'],
    to: ({ owner, repo }) => ({
      name: 'repository-automation',
      params: { owner, repo },
    }),
  },
  {
    id: 'repo-activity',
    label: 'Activity',
    description: 'Health, branches, and repository metadata.',
    icon: Activity,
    feature: 'repositoryActivity',
    routeName: 'repository-activity',
    matchRouteNames: ['repository-activity'],
    to: ({ owner, repo }) => ({
      name: 'repository-activity',
      params: { owner, repo },
    }),
  },
  {
    id: 'repo-settings',
    label: 'Settings',
    description: 'Future repository controls and guardrails.',
    icon: Settings2,
    feature: 'repositorySettings',
    routeName: 'repository-settings',
    matchRouteNames: ['repository-settings'],
    to: ({ owner, repo }) => ({
      name: 'repository-settings',
      params: { owner, repo },
    }),
  },
]

export function isFeatureEnabled(feature: FeatureFlag) {
  return bootstrap.features[feature]
}

export function getWorkspaceNavigation() {
  return workspaceItems.filter((item) => isFeatureEnabled(item.feature))
}

export function getWorkspaceNavigationGroups() {
  const items = getWorkspaceNavigation()
  return [
    {
      id: 'workspace',
      label: 'Workspace',
      items: items.filter((item) => item.group === 'workspace'),
    },
    {
      id: 'collaboration',
      label: 'Collaboration',
      items: items.filter((item) => item.group === 'collaboration'),
    },
  ].filter((group) => group.items.length > 0)
}

export function getRepositoryNavigation() {
  return repositoryItems.filter((item) => isFeatureEnabled(item.feature))
}

export function getDefaultWorkspaceRoute(): RouteLocationRaw {
  const firstItem = getWorkspaceNavigation()[0]
  return firstItem?.to ?? { name: 'repositories' }
}

export function getDefaultRepositoryRoute(owner: string, repo: string, currentQuery?: LocationQuery): RouteLocationRaw {
  const firstItem = getRepositoryNavigation()[0]
  return firstItem ? firstItem.to({ owner, repo, currentQuery }) : { name: 'repositories' }
}

export function isNavigationItemActive(item: Pick<BaseNavigationItem, 'matchRouteNames'>, route: RouteLocationNormalizedLoaded) {
  return item.matchRouteNames.includes(String(route.name ?? ''))
}
