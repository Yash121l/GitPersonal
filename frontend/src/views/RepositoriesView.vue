<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { BookCopy, FolderGit2, Plus } from '@lucide/vue'
import { computed, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'

import EmptyState from '@/components/empty/EmptyState.vue'
import PageHeader from '@/components/app/PageHeader.vue'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import Input from '@/components/ui/Input.vue'
import Select from '@/components/ui/Select.vue'
import Textarea from '@/components/ui/Textarea.vue'
import { api } from '@/lib/api'
import { formatDate } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const queryClient = useQueryClient()

const repositoriesQuery = useQuery({
  queryKey: ['repositories'],
  queryFn: () => api.listRepositories(),
})

const organizationsQuery = useQuery({
  queryKey: ['organizations'],
  queryFn: () => api.listOrganizations(),
})

const personalRepoForm = reactive({
  name: '',
  description: '',
  visibility: 'private',
  default_branch: 'main',
})

const orgForm = reactive({
  slug: '',
  display_name: '',
  description: '',
})

const orgRepoForm = reactive({
  owner: '',
  name: '',
  description: '',
  visibility: 'private',
  default_branch: 'main',
})

const formError = ref('')

const organizations = computed(() => organizationsQuery.data.value ?? [])
const repositories = computed(() => repositoriesQuery.data.value ?? [])

const createRepository = useMutation({
  mutationFn: api.createRepository,
  onSuccess: async () => {
    formError.value = ''
    personalRepoForm.name = ''
    personalRepoForm.description = ''
    orgRepoForm.name = ''
    orgRepoForm.description = ''
    await queryClient.invalidateQueries({ queryKey: ['repositories'] })
  },
})

const createOrganization = useMutation({
  mutationFn: api.createOrganization,
  onSuccess: async () => {
    formError.value = ''
    orgForm.slug = ''
    orgForm.display_name = ''
    orgForm.description = ''
    await queryClient.invalidateQueries({ queryKey: ['organizations'] })
  },
})

async function handleCreatePersonalRepository() {
  formError.value = ''
  try {
    await createRepository.mutateAsync({ ...personalRepoForm })
  } catch (error) {
    formError.value = error instanceof Error ? error.message : 'Unable to create repository.'
  }
}

async function handleCreateOrganization() {
  formError.value = ''
  try {
    await createOrganization.mutateAsync({ ...orgForm })
  } catch (error) {
    formError.value = error instanceof Error ? error.message : 'Unable to create organization.'
  }
}

async function handleCreateOrganizationRepository() {
  formError.value = ''
  try {
    await createRepository.mutateAsync({
      ...orgRepoForm,
      owner_type: 'organization',
    })
  } catch (error) {
    formError.value = error instanceof Error ? error.message : 'Unable to create organization repository.'
  }
}
</script>

<template>
  <div class="section-stack">
    <PageHeader
      eyebrow="Repositories"
      title="Repositories"
      description="Browse visible repositories and create personal or organization-owned repos."
    />

    <div class="grid gap-3 xl:grid-cols-[1.12fr_0.88fr]">
      <Card class="space-y-4">
        <div class="panel-header">
          <div>
            <p class="eyebrow">Visible Repositories</p>
            <h3 class="mt-1 text-lg font-semibold text-zinc-50">
              {{ repositories.length }} repository{{ repositories.length === 1 ? '' : 'ies' }}
            </h3>
          </div>
          <Badge variant="accent">{{ authStore.currentUser?.role }}</Badge>
        </div>

        <div v-if="repositoriesQuery.isLoading.value" class="space-y-3">
          <div v-for="index in 4" :key="index" class="h-24 animate-pulse rounded-xl bg-zinc-900" />
        </div>
        <div v-else-if="repositories.length" class="grid gap-3">
          <RouterLink
            v-for="repository in repositories"
            :key="`${repository.owner}/${repository.name}`"
            :to="{ name: 'repository-code', params: { owner: repository.owner, repo: repository.name } }"
            class="group rounded-lg border border-zinc-800 bg-black/30 p-4 transition hover:border-zinc-700 hover:bg-zinc-950"
          >
            <div class="flex flex-wrap items-center gap-2">
              <Badge>{{ repository.owner_type }}</Badge>
              <Badge variant="accent">{{ repository.visibility }}</Badge>
            </div>
            <h3 class="mt-2.5 font-mono text-base font-semibold text-zinc-50 group-hover:text-white">
              {{ repository.owner }}/{{ repository.name }}
            </h3>
            <p class="mt-1.5 text-sm text-zinc-400">
              {{ repository.description || 'No description yet.' }}
            </p>
            <div class="meta-list mt-3">
              <span>Default branch: {{ repository.default_branch }}</span>
              <span>Updated {{ formatDate(repository.updated_at) }}</span>
            </div>
          </RouterLink>
        </div>
        <EmptyState
          v-else
          eyebrow="No Repositories"
          title="Nothing is visible to this account yet."
          description="Create your first personal repository or create an organization repository from the forms on the right."
        >
          <Button @click="personalRepoForm.name = 'forge-app'">
            <Plus class="size-4" />
            Seed a Name
          </Button>
        </EmptyState>
      </Card>

      <div class="space-y-3">
        <Card class="space-y-4">
          <div class="panel-header">
            <div>
              <p class="eyebrow">Create Personal Repository</p>
              <h3 class="mt-1 text-lg font-semibold text-zinc-50">Personal repository</h3>
            </div>
          </div>
          <div>
            <label class="field-label">Name</label>
            <Input v-model="personalRepoForm.name" />
          </div>
          <div>
            <label class="field-label">Description</label>
            <Textarea v-model="personalRepoForm.description" />
          </div>
          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <label class="field-label">Visibility</label>
              <Select v-model="personalRepoForm.visibility">
                <option value="private">Private</option>
                <option value="public">Public</option>
              </Select>
            </div>
            <div>
              <label class="field-label">Default branch</label>
              <Input v-model="personalRepoForm.default_branch" />
            </div>
          </div>
          <Button :disabled="createRepository.isPending.value" @click="handleCreatePersonalRepository">
            <FolderGit2 class="size-4" />
            {{ createRepository.isPending.value ? 'Creating...' : 'Create Repository' }}
          </Button>
        </Card>

        <Card class="space-y-4">
          <div class="panel-header">
            <div>
              <p class="eyebrow">Create Organization</p>
              <h3 class="mt-1 text-lg font-semibold text-zinc-50">Organization namespace</h3>
            </div>
          </div>
          <div>
            <label class="field-label">Slug</label>
            <Input v-model="orgForm.slug" />
          </div>
          <div>
            <label class="field-label">Display name</label>
            <Input v-model="orgForm.display_name" />
          </div>
          <div>
            <label class="field-label">Description</label>
            <Textarea v-model="orgForm.description" />
          </div>
          <Button :disabled="createOrganization.isPending.value" variant="secondary" @click="handleCreateOrganization">
            {{ createOrganization.isPending.value ? 'Creating...' : 'Create Organization' }}
          </Button>
        </Card>

        <Card class="space-y-4">
          <div class="panel-header">
            <div>
              <p class="eyebrow">Create Organization Repository</p>
              <h3 class="mt-1 text-lg font-semibold text-zinc-50">Organization repository</h3>
            </div>
          </div>
          <div v-if="organizations.length === 0" class="rounded-lg border border-zinc-800 bg-black/30 p-4 text-sm text-zinc-400">
            You need an organization membership before creating an organization-owned repository.
          </div>
          <template v-else>
            <div>
              <label class="field-label">Organization</label>
              <Select v-model="orgRepoForm.owner">
                <option disabled value="">Choose an organization</option>
                <option
                  v-for="organization in organizations"
                  :key="organization.organization_slug"
                  :value="organization.organization_slug"
                >
                  {{ organization.organization_slug }} ({{ organization.role }})
                </option>
              </Select>
            </div>
            <div>
              <label class="field-label">Name</label>
              <Input v-model="orgRepoForm.name" />
            </div>
            <div>
              <label class="field-label">Description</label>
              <Textarea v-model="orgRepoForm.description" />
            </div>
            <Button :disabled="createRepository.isPending.value" variant="secondary" @click="handleCreateOrganizationRepository">
              <BookCopy class="size-4" />
              Create Shared Repository
            </Button>
          </template>
          <div
            v-if="formError"
            class="rounded-md border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-300"
          >
            {{ formError }}
          </div>
        </Card>
      </div>
    </div>
  </div>
</template>
