import { defaultFeatureFlags, type FeatureFlag } from '@/app/features'

export interface ForgeBootstrapPayload {
  basePath?: string
  productName?: string
  features?: Partial<Record<FeatureFlag, boolean>>
}

export interface ForgeBootstrap {
  basePath: string
  productName: string
  features: Record<FeatureFlag, boolean>
}

declare global {
  interface Window {
    __FORGE_BOOTSTRAP__?: ForgeBootstrapPayload
  }
}

function normalizeBasePath(value?: string) {
  if (typeof value !== 'string') {
    return '/app/'
  }

  const trimmed = value.trim()
  if (trimmed === '') {
    return '/app/'
  }

  const withLeadingSlash = trimmed.startsWith('/') ? trimmed : `/${trimmed}`
  return withLeadingSlash.endsWith('/') ? withLeadingSlash : `${withLeadingSlash}/`
}

function resolveFeatureFlags(features?: Partial<Record<FeatureFlag, boolean>>) {
  return Object.entries(defaultFeatureFlags).reduce(
    (resolved, [key, defaultValue]) => {
      const nextValue = features?.[key as FeatureFlag]
      resolved[key as FeatureFlag] = typeof nextValue === 'boolean' ? nextValue : defaultValue
      return resolved
    },
    {} as Record<FeatureFlag, boolean>,
  )
}

export function resolveBootstrap(payload?: ForgeBootstrapPayload): ForgeBootstrap {
  return {
    basePath: normalizeBasePath(payload?.basePath),
    productName: typeof payload?.productName === 'string' && payload.productName.trim() !== '' ? payload.productName : 'Forge',
    features: resolveFeatureFlags(payload?.features),
  }
}

export const bootstrap = resolveBootstrap(typeof window !== 'undefined' ? window.__FORGE_BOOTSTRAP__ : undefined)
