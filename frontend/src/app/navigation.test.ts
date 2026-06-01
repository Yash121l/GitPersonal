import { describe, expect, it } from 'vitest'

import { getDefaultRepositoryRoute, getDefaultWorkspaceRoute, getRepositoryNavigation, getWorkspaceNavigationGroups } from '@/app/navigation'

describe('navigation config', () => {
  it('uses overview as the default authenticated workspace route', () => {
    expect(getDefaultWorkspaceRoute()).toEqual({ name: 'overview' })
  })

  it('exposes the GitHub-style repository sections', () => {
    const labels = getRepositoryNavigation().map((item) => item.label)

    expect(labels).toEqual(['Code', 'Issues', 'Pull Requests', 'Actions', 'Insights', 'Access', 'Automation', 'Activity', 'Settings'])
  })

  it('builds repository destinations from the configured default module', () => {
    expect(getDefaultRepositoryRoute('yash', 'forge')).toEqual({
      name: 'repository-code',
      params: { owner: 'yash', repo: 'forge' },
      query: undefined,
    })
  })

  it('keeps workspace navigation grouped for the shell', () => {
    const groups = getWorkspaceNavigationGroups()

    expect(groups.map((group) => group.label)).toEqual(['Workspace', 'Collaboration', 'Account'])
    expect(groups[0]?.items.map((item) => item.label)).toEqual(['Overview', 'Repositories', 'Explore'])
  })
})
