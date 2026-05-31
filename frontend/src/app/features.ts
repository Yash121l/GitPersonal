export const defaultFeatureFlags = {
  workspaceOverview: true,
  repositories: true,
  organizations: true,
  sshKeys: true,
  repositoryCode: true,
  repositoryAccess: true,
  repositoryAutomation: true,
  repositoryActivity: true,
  repositorySettings: false,
} as const

export type FeatureFlag = keyof typeof defaultFeatureFlags

export interface FeatureDefinition {
  key: FeatureFlag
  label: string
  description: string
  scope: 'workspace' | 'repository'
}

export const featureCatalog: Record<FeatureFlag, FeatureDefinition> = {
  workspaceOverview: {
    key: 'workspaceOverview',
    label: 'Overview',
    description: 'Developer dashboard with system-wide inventory and quick links.',
    scope: 'workspace',
  },
  repositories: {
    key: 'repositories',
    label: 'Repositories',
    description: 'Repository inventory and creation workflows.',
    scope: 'workspace',
  },
  organizations: {
    key: 'organizations',
    label: 'Organizations',
    description: 'Shared namespaces, memberships, and collaboration boundaries.',
    scope: 'workspace',
  },
  sshKeys: {
    key: 'sshKeys',
    label: 'SSH Keys',
    description: 'Developer identity and SSH transport management.',
    scope: 'workspace',
  },
  repositoryCode: {
    key: 'repositoryCode',
    label: 'Code',
    description: 'Branch-aware repository browser with bounded blob previews.',
    scope: 'repository',
  },
  repositoryAccess: {
    key: 'repositoryAccess',
    label: 'Access',
    description: 'Clone endpoints, visibility, and collaborator controls.',
    scope: 'repository',
  },
  repositoryAutomation: {
    key: 'repositoryAutomation',
    label: 'Automation',
    description: 'Webhook endpoints and automation controls.',
    scope: 'repository',
  },
  repositoryActivity: {
    key: 'repositoryActivity',
    label: 'Activity',
    description: 'Repository health and operational metadata.',
    scope: 'repository',
  },
  repositorySettings: {
    key: 'repositorySettings',
    label: 'Settings',
    description: 'Reserved slot for future repository configuration surfaces.',
    scope: 'repository',
  },
}
