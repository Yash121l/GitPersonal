export interface User {
  id: number
  username: string
  role: string
  created_at: string
}

export interface Repository {
  id: number
  owner: string
  owner_type: string
  name: string
  description: string
  visibility: string
  default_branch: string
  archived: boolean
  size_bytes: number
  last_indexed_at?: string | null
  last_maintained_at?: string | null
  created_at: string
  updated_at: string
}

export interface OrganizationMembership {
  organization_id: number
  organization_slug: string
  organization_display_name: string
  user_id: number
  username?: string
  role: string
  created_at: string
}

export interface SSHKey {
  id: number
  user_id: number
  name: string
  public_key: string
  fingerprint_sha256: string
  created_at: string
  last_used_at?: string | null
}

export interface RepositoryWebhook {
  id: number
  repository_id: number
  url: string
  events: string[]
  created_at: string
  last_delivery_at?: string | null
  last_delivery_status?: number
  last_delivery_error?: string
  success_count: number
  failure_count: number
}

export interface RepositoryBranch {
  name: string
}

export interface RepositoryTreeEntry {
  path: string
  name: string
  type: 'tree' | 'blob'
  mode: string
  object_id: string
  size_bytes: number
}

export interface RepositoryBlob {
  path: string
  size_bytes: number
  content: string
  truncated: boolean
  is_binary: boolean
  language: string
}

export interface RepositoryDetailResponse {
  repository: Repository
  http_clone_url: string
  ssh_clone_url?: string
}

export interface RepositoryTreeResponse {
  ref: string
  path: string
  entries: RepositoryTreeEntry[] | null
}

export interface RepositoryBlobResponse {
  ref: string
  blob: RepositoryBlob
}

export interface RepositoryBranchListResponse {
  branches: RepositoryBranch[] | null
}

type UnauthorizedHandler = () => void | Promise<void>

let unauthorizedHandler: UnauthorizedHandler | null = null

export class ApiError extends Error {
  status: number
  requestId?: string

  constructor(message: string, status: number, requestId?: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.requestId = requestId
  }
}

export function setUnauthorizedHandler(handler: UnauthorizedHandler | null) {
  unauthorizedHandler = handler
}

async function apiFetch<T>(input: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers)
  headers.set('Accept', 'application/json')

  if (init?.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const response = await fetch(input, {
    ...init,
    headers,
    credentials: 'same-origin',
  })

  const text = await response.text()
  const payload = text ? safeParseJSON(text) : null

  if (!response.ok) {
    if (response.status === 401) {
      await unauthorizedHandler?.()
    }
    const message =
      (isRecord(payload) && typeof payload.error === 'string' && payload.error) ||
      `Request failed with status ${response.status}`
    const requestId = isRecord(payload) && typeof payload.request_id === 'string' ? payload.request_id : undefined
    throw new ApiError(message, response.status, requestId)
  }

  return (payload ?? {}) as T
}

function safeParseJSON(text: string) {
  try {
    return JSON.parse(text)
  } catch {
    return null
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function ensureArray<T>(value: T[] | null | undefined) {
  return Array.isArray(value) ? value : []
}

export const api = {
  me() {
    return apiFetch<{ user: User }>('/api/v1/me')
  },
  login(payload: { username: string; password: string }) {
    return apiFetch<{ user: User }>('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  register(payload: { username: string; password: string }) {
    return apiFetch<{ user: User }>('/api/v1/auth/register', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  logout() {
    return apiFetch<{ status: string }>('/api/v1/auth/logout', { method: 'POST' })
  },
  async listRepositories() {
    const payload = await apiFetch<{ repositories: Repository[] | null }>('/api/v1/repos')
    return ensureArray(payload.repositories)
  },
  createRepository(payload: {
    owner?: string
    owner_type?: string
    name: string
    description: string
    visibility: string
    default_branch: string
  }) {
    return apiFetch<Repository>('/api/v1/repos', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  repository(owner: string, repo: string) {
    return apiFetch<RepositoryDetailResponse>(`/api/v1/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}`)
  },
  async branches(owner: string, repo: string) {
    const payload = await apiFetch<RepositoryBranchListResponse>(
      `/api/v1/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/branches`,
    )
    return ensureArray(payload.branches)
  },
  tree(owner: string, repo: string, params: { ref?: string; path?: string }) {
    const query = new URLSearchParams()
    if (params.ref) {
      query.set('ref', params.ref)
    }
    if (params.path) {
      query.set('path', params.path)
    }
    return apiFetch<RepositoryTreeResponse>(
      `/api/v1/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/tree?${query.toString()}`,
    ).then((payload) => ({ ...payload, entries: ensureArray(payload.entries) }))
  },
  blob(owner: string, repo: string, params: { ref?: string; path: string }) {
    const query = new URLSearchParams({ path: params.path })
    if (params.ref) {
      query.set('ref', params.ref)
    }
    return apiFetch<RepositoryBlobResponse>(
      `/api/v1/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/blob?${query.toString()}`,
    )
  },
  async listOrganizations() {
    const payload = await apiFetch<{ organizations: OrganizationMembership[] | null }>('/api/v1/orgs')
    return ensureArray(payload.organizations)
  },
  createOrganization(payload: { slug: string; display_name?: string; description?: string }) {
    return apiFetch<{ id: number; slug: string; display_name: string; description: string }>('/api/v1/orgs', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  addOrganizationMember(org: string, payload: { username: string; role: string }) {
    return apiFetch<OrganizationMembership>(`/api/v1/orgs/${encodeURIComponent(org)}/members`, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  async listKeys() {
    const payload = await apiFetch<{ keys: SSHKey[] | null }>('/api/v1/keys')
    return ensureArray(payload.keys)
  },
  createKey(payload: { name: string; public_key: string }) {
    return apiFetch<SSHKey>('/api/v1/keys', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  addCollaborator(owner: string, repo: string, payload: { username: string; role: string }) {
    return apiFetch(`/api/v1/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/collaborators`, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  async listWebhooks(owner: string, repo: string) {
    const payload = await apiFetch<{ webhooks: RepositoryWebhook[] | null }>(
      `/api/v1/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/webhooks`,
    )
    return ensureArray(payload.webhooks)
  },
  createWebhook(owner: string, repo: string, payload: { url: string; secret?: string; events: string[] }) {
    return apiFetch<RepositoryWebhook>(`/api/v1/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/webhooks`, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  deleteWebhook(owner: string, repo: string, webhookId: number) {
    return apiFetch<void>(`/api/v1/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/webhooks/${webhookId}`, {
      method: 'DELETE',
    })
  },
}
