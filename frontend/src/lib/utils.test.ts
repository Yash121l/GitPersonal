import { describe, expect, it } from 'vitest'

import { basename, dirname, formatBytes, titleCase } from './utils'

describe('utils', () => {
  it('formats byte counts predictably', () => {
    expect(formatBytes(0)).toBe('0 B')
    expect(formatBytes(512)).toBe('512 B')
    expect(formatBytes(1536)).toBe('1.5 KB')
    expect(formatBytes(10 * 1024 * 1024)).toBe('10 MB')
  })

  it('derives file paths cleanly', () => {
    expect(dirname('internal/server/ui.go')).toBe('internal/server')
    expect(dirname('README.md')).toBe('')
    expect(basename('internal/server/ui.go')).toBe('ui.go')
  })

  it('normalizes labels into title case', () => {
    expect(titleCase('repository.push')).toBe('Repository.push')
    expect(titleCase('ssh_keys')).toBe('Ssh Keys')
  })
})
