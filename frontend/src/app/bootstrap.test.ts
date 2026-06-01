import { describe, expect, it } from 'vitest'

import { resolveBootstrap } from '@/app/bootstrap'

describe('resolveBootstrap', () => {
  it('normalizes the base path and merges feature defaults', () => {
    const resolved = resolveBootstrap({
      basePath: 'app',
      productName: 'Forge Enterprise',
      features: {
        repositorySettings: true,
        sshKeys: false,
      },
    })

    expect(resolved.basePath).toBe('/app/')
    expect(resolved.productName).toBe('Forge Enterprise')
    expect(resolved.features.repositorySettings).toBe(true)
    expect(resolved.features.sshKeys).toBe(false)
    expect(resolved.features.repositories).toBe(true)
  })

  it('falls back to safe defaults for empty payloads', () => {
    const resolved = resolveBootstrap({})

    expect(resolved.basePath).toBe('/app/')
    expect(resolved.productName).toBe('Forge')
    expect(resolved.features.workspaceOverview).toBe(true)
    expect(resolved.features.repositorySettings).toBe(true)
  })
})
