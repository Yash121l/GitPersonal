import { useQuery } from '@tanstack/vue-query'
import { computed, inject, provide, type InjectionKey } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { api } from '@/lib/api'
import { dirname } from '@/lib/utils'

function createRepositoryWorkspace() {
  const route = useRoute()
  const router = useRouter()

  const owner = computed(() => String(route.params.owner))
  const repo = computed(() => String(route.params.repo))
  const selectedRef = computed(() => (typeof route.query.ref === 'string' ? route.query.ref : undefined))
  const selectedPath = computed(() => (typeof route.query.path === 'string' ? route.query.path : ''))
  const selectedFile = computed(() => (typeof route.query.file === 'string' ? route.query.file : ''))

  const repositoryQuery = useQuery({
    queryKey: ['repository', owner, repo],
    queryFn: () => api.repository(owner.value, repo.value),
  })

  const branchesQuery = useQuery({
    queryKey: ['branches', owner, repo],
    queryFn: () => api.branches(owner.value, repo.value),
  })

  const currentBranch = computed(
    () => selectedRef.value || repositoryQuery.data.value?.repository.default_branch || 'main',
  )

  const branchModel = computed({
    get: () => currentBranch.value,
    set: (value: string) => {
      void handleBranchChange(value)
    },
  })

  const breadcrumbs = computed(() => {
    if (!selectedPath.value) {
      return []
    }

    return selectedPath.value.split('/').map((segment, index, parts) => ({
      label: segment,
      value: parts.slice(0, index + 1).join('/'),
    }))
  })

  async function updateQuery(nextQuery: Record<string, string | undefined>) {
    await router.replace({
      query: Object.fromEntries(Object.entries(nextQuery).filter(([, value]) => value)),
    })
  }

  async function handleBranchChange(branch: string) {
    await updateQuery({
      ref: branch,
      path: selectedPath.value || undefined,
      file: undefined,
    })
  }

  async function openDirectory(pathValue = '') {
    await updateQuery({
      ref: selectedRef.value,
      path: pathValue || undefined,
      file: undefined,
    })
  }

  async function openFile(filePath: string) {
    await updateQuery({
      ref: selectedRef.value,
      path: dirname(filePath) || undefined,
      file: filePath,
    })
  }

  async function handleTreeSelect(entry: { path: string; type: string }) {
    if (entry.type === 'tree') {
      await openDirectory(entry.path)
      return
    }

    await openFile(entry.path)
  }

  return {
    owner,
    repo,
    selectedRef,
    selectedPath,
    selectedFile,
    repositoryQuery,
    branchesQuery,
    currentBranch,
    branchModel,
    breadcrumbs,
    updateQuery,
    handleBranchChange,
    openDirectory,
    openFile,
    handleTreeSelect,
  }
}

type RepositoryWorkspace = ReturnType<typeof createRepositoryWorkspace>

const repositoryWorkspaceKey: InjectionKey<RepositoryWorkspace> = Symbol('repository-workspace')

export function provideRepositoryWorkspace() {
  const workspace = createRepositoryWorkspace()
  provide(repositoryWorkspaceKey, workspace)
  return workspace
}

export function useRepositoryWorkspace() {
  const workspace = inject(repositoryWorkspaceKey)
  if (!workspace) {
    throw new Error('Repository workspace context is not available.')
  }
  return workspace
}
