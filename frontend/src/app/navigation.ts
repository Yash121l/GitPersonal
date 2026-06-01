import {
  Activity,
  Compass,
  FolderGit2,
  Gauge,
  KeyRound,
  Landmark,
  Search,
  Settings2,
  ShieldCheck,
  UserRound,
  Zap,
} from '@lucide/vue'
import type { Component } from 'vue'
import type { LocationQuery, RouteLocationRaw, RouteLocationNormalizedLoaded } from 'vue-router'

import { bootstrap } from '@/app/bootstrap'
import type { FeatureFlag } from '@/app/features'

export type AppRouteName =
  | 'overview'
  | 'explore'
  | 'repositories'
  | 'new-repository'
  | 'new-organization'
  | 'organizations'
  | 'keys'
  | 'account-settings'
  | 'profile'
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
  group: 'workspace' | 'collaboration' | 'account'
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
    description: 'Inventory, recent repositories, and workspace status.',
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
    description: 'Browse codebases and create new repositories.',
    icon: FolderGit2,
    feature: 'repositories',
    routeName: 'repositories',
    to: { name: 'repositories' },
    group: 'workspace',
    matchRouteNames: [
      'repositories',
      'new-repository',
      'new-organization',
      'repository-code',
      'repository-access',
      'repository-automation',
      'repository-activity',
      'repository-settings',
    ],
  },
  {
    id: 'workspace-explore',
    label: 'Explore',
    description: 'Search users and public repositories.',
    icon: Compass,
    feature: 'repositories',
    routeName: 'explore',
    to: { name: 'explore' },
    group: 'workspace',
    matchRouteNames: ['explore'],
  },
  {
    id: 'workspace-organizations',
    label: 'Organizations',
    description: 'Shared namespaces, members, and roles.',
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
    description: 'SSH identities for clone and push access.',
    icon: KeyRound,
    feature: 'sshKeys',
    routeName: 'keys',
    to: { name: 'keys' },
    group: 'collaboration',
    matchRouteNames: ['keys'],
  },
  {
    id: 'workspace-profile',
    label: 'Profile',
    description: 'Public account page and owned repositories.',
    icon: UserRound,
    feature: 'workspaceOverview',
    routeName: 'profile',
    to: { name: 'profile' },
    group: 'account',
    matchRouteNames: ['profile'],
  },
  {
    id: 'workspace-settings',
    label: 'Settings',
    description: 'Account profile, SSH keys, and preferences.',
    icon: Settings2,
    feature: 'workspaceOverview',
    routeName: 'account-settings',
    to: { name: 'account-settings' },
    group: 'account',
    matchRouteNames: ['account-settings'],
  },
]

const repositoryItems: RepositoryNavigationItem[] = [
  {
    id: 'repo-code',
    label: 'Code',
    description: 'Tree browser, branch selection, and file previews.',
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
    id: 'repo-issues',
    label: 'Issues',
    description: 'Issue tracking route scaffold.',
    icon: Search,
    feature: 'repositoryAccess',
    routeName: 'repository-access',
    matchRouteNames: ['repository-issues'],
    to: ({ owner, repo }) => ({
      name: 'repository-issues',
      params: { owner, repo },
    }),
  },
  {
    id: 'repo-pulls',
    label: 'Pull Requests',
    description: 'Pull request route scaffold.',
    icon: Activity,
    feature: 'repositoryAccess',
    routeName: 'repository-access',
    matchRouteNames: ['repository-pulls'],
    to: ({ owner, repo }) => ({
      name: 'repository-pulls',
      params: { owner, repo },
    }),
  },
  {
    id: 'repo-actions',
    label: 'Actions',
    description: 'CI workflow and webhook automation.',
    icon: Zap,
    feature: 'repositoryAutomation',
    routeName: 'repository-automation',
    matchRouteNames: ['repository-actions', 'repository-automation'],
    to: ({ owner, repo }) => ({
      name: 'repository-actions',
      params: { owner, repo },
    }),
  },
  {
    id: 'repo-insights',
    label: 'Insights',
    description: 'Repository activity and health signals.',
    icon: Activity,
    feature: 'repositoryActivity',
    routeName: 'repository-activity',
    matchRouteNames: ['repository-insights', 'repository-activity'],
    to: ({ owner, repo }) => ({
      name: 'repository-insights',
      params: { owner, repo },
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
    description: 'Webhooks and downstream integrations.',
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
    description: 'Repository metadata, branches, and health signals.',
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
    description: 'Guardrails and future repository controls.',
    icon: Settings2,
    feature: 'repositorySettings',
    routeName: 'repository-settings',
    matchRouteNames: ['repository-settings', 'repository-settings-access', 'repository-settings-branches', 'repository-settings-hooks', 'repository-settings-keys', 'repository-settings-pages', 'repository-settings-secrets'],
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
    {
      id: 'account',
      label: 'Account',
      items: items.filter((item) => item.group === 'account'),
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
