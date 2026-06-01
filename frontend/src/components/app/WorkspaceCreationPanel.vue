<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { BookCopy, FolderGit2, Landmark } from '@lucide/vue'
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'

import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import Input from '@/components/ui/Input.vue'
import Select from '@/components/ui/Select.vue'
import Textarea from '@/components/ui/Textarea.vue'
import { api, type Repository } from '@/lib/api'

const router = useRouter()
const queryClient = useQueryClient()

const emit = defineEmits<{
  created: [repository: Repository]
}>()

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

const submitAttempted = reactive({
  personal: false,
  organization: false,
  organizationRepo: false,
})

const personalServerError = ref('')
const organizationServerError = ref('')
const organizationRepoServerError = ref('')

const organizations = computed(() => organizationsQuery.data.value ?? [])

const personalNameError = computed(() =>
  submitAttempted.personal && personalRepoForm.name.trim() === '' ? 'Repository name is required.' : '',
)
const personalBranchError = computed(() =>
  submitAttempted.personal && personalRepoForm.default_branch.trim() === '' ? 'Default branch is required.' : '',
)
const organizationSlugError = computed(() =>
  submitAttempted.organization && orgForm.slug.trim() === '' ? 'Organization slug is required.' : '',
)
const organizationRepoOwnerError = computed(() =>
  submitAttempted.organizationRepo && orgRepoForm.owner.trim() === '' ? 'Choose an organization.' : '',
)
const organizationRepoNameError = computed(() =>
  submitAttempted.organizationRepo && orgRepoForm.name.trim() === '' ? 'Repository name is required.' : '',
)
const organizationRepoBranchError = computed(() =>
  submitAttempted.organizationRepo && orgRepoForm.default_branch.trim() === '' ? 'Default branch is required.' : '',
)

const personalRepoValid = computed(() => personalRepoForm.name.trim() !== '' && personalRepoForm.default_branch.trim() !== '')
const organizationValid = computed(() => orgForm.slug.trim() !== '')
const organizationRepoValid = computed(
  () =>
    orgRepoForm.owner.trim() !== '' &&
    orgRepoForm.name.trim() !== '' &&
    orgRepoForm.default_branch.trim() !== '',
)

const createRepository = useMutation({
  mutationFn: api.createRepository,
  onSuccess: async (repository) => {
    personalRepoForm.name = ''
    personalRepoForm.description = ''
    personalRepoForm.default_branch = 'main'
    orgRepoForm.name = ''
    orgRepoForm.description = ''
    orgRepoForm.default_branch = 'main'
    personalServerError.value = ''
    organizationRepoServerError.value = ''
    submitAttempted.personal = false
    submitAttempted.organizationRepo = false
    emit('created', repository)
    await queryClient.invalidateQueries({ queryKey: ['repositories'] })
    await router.push({ name: 'repository-code', params: { owner: repository.owner, repo: repository.name } })
  },
})

const createOrganization = useMutation({
  mutationFn: api.createOrganization,
  onSuccess: async () => {
    orgForm.slug = ''
    orgForm.display_name = ''
    orgForm.description = ''
    organizationServerError.value = ''
    submitAttempted.organization = false
    await queryClient.invalidateQueries({ queryKey: ['organizations'] })
  },
})

async function handleCreatePersonalRepository() {
  submitAttempted.personal = true
  personalServerError.value = ''
  if (!personalRepoValid.value) {
    return
  }

  try {
    await createRepository.mutateAsync({ ...personalRepoForm })
  } catch (error) {
    personalServerError.value = error instanceof Error ? error.message : 'Unable to create repository.'
  }
}

async function handleCreateOrganization() {
  submitAttempted.organization = true
  organizationServerError.value = ''
  if (!organizationValid.value) {
    return
  }

  try {
    await createOrganization.mutateAsync({ ...orgForm })
  } catch (error) {
    organizationServerError.value = error instanceof Error ? error.message : 'Unable to create organization.'
  }
}

async function handleCreateOrganizationRepository() {
  submitAttempted.organizationRepo = true
  organizationRepoServerError.value = ''
  if (!organizationRepoValid.value) {
    return
  }

  try {
    await createRepository.mutateAsync({
      ...orgRepoForm,
      owner_type: 'organization',
    })
  } catch (error) {
    organizationRepoServerError.value = error instanceof Error ? error.message : 'Unable to create organization repository.'
  }
}
</script>

<template>
  <div class="grid gap-6 xl:grid-cols-[1fr_1fr]">
    <Card class="space-y-4">
      <div class="flex items-start gap-3 border-b border-zinc-800 pb-4">
        <div class="flex size-9 items-center justify-center rounded-md border border-zinc-800 bg-zinc-900 text-zinc-300">
          <FolderGit2 class="size-4" />
        </div>
        <div>
          <p class="eyebrow">Personal</p>
          <h3 class="mt-1 text-lg font-semibold text-zinc-50">Repository</h3>
          <p class="mt-1 text-sm leading-6 text-zinc-400">Create a codebase in your user namespace.</p>
        </div>
      </div>

      <div>
        <label class="field-label">Name</label>
        <Input v-model="personalRepoForm.name" placeholder="forge-web" />
        <p v-if="personalNameError" class="mt-1 text-xs text-red-400">{{ personalNameError }}</p>
      </div>
      <div>
        <label class="field-label">Description</label>
        <Textarea v-model="personalRepoForm.description" placeholder="Internal Git workspace UI refresh" />
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
          <p v-if="personalBranchError" class="mt-1 text-xs text-red-400">{{ personalBranchError }}</p>
        </div>
      </div>
      <p v-if="personalServerError" class="text-sm text-red-400">{{ personalServerError }}</p>
      <Button :disabled="createRepository.isPending.value" @click="handleCreatePersonalRepository">
        <FolderGit2 class="size-4" />
        {{ createRepository.isPending.value ? 'Creating...' : 'Create repository' }}
      </Button>
    </Card>

    <div class="space-y-6">
      <Card class="space-y-4">
        <div class="flex items-start gap-3 border-b border-zinc-800 pb-4">
          <div class="flex size-9 items-center justify-center rounded-md border border-zinc-800 bg-zinc-900 text-zinc-300">
            <Landmark class="size-4" />
          </div>
          <div>
            <p class="eyebrow">Organization</p>
            <h3 class="mt-1 text-lg font-semibold text-zinc-50">Namespace</h3>
            <p class="mt-1 text-sm leading-6 text-zinc-400">Create a shared owner for team repositories.</p>
          </div>
        </div>
        <div>
          <label class="field-label">Slug</label>
          <Input v-model="orgForm.slug" placeholder="platform" />
          <p v-if="organizationSlugError" class="mt-1 text-xs text-red-400">{{ organizationSlugError }}</p>
        </div>
        <div>
          <label class="field-label">Display name</label>
          <Input v-model="orgForm.display_name" placeholder="Platform Team" />
        </div>
        <div>
          <label class="field-label">Description</label>
          <Textarea v-model="orgForm.description" placeholder="Shared ownership for platform repositories" />
        </div>
        <p v-if="organizationServerError" class="text-sm text-red-400">{{ organizationServerError }}</p>
        <Button :disabled="createOrganization.isPending.value" variant="secondary" @click="handleCreateOrganization">
          {{ createOrganization.isPending.value ? 'Creating...' : 'Create organization' }}
        </Button>
      </Card>

      <Card class="space-y-4">
        <div class="flex items-start gap-3 border-b border-zinc-800 pb-4">
          <div class="flex size-9 items-center justify-center rounded-md border border-zinc-800 bg-zinc-900 text-zinc-300">
            <BookCopy class="size-4" />
          </div>
          <div>
            <p class="eyebrow">Organization</p>
            <h3 class="mt-1 text-lg font-semibold text-zinc-50">Repository</h3>
            <p class="mt-1 text-sm leading-6 text-zinc-400">Create a repository under an organization owner.</p>
          </div>
        </div>

        <div v-if="organizations.length === 0" class="rounded-md border border-zinc-800 bg-zinc-900/60 p-4 text-sm text-zinc-400">
          Create or join an organization before adding an organization-owned repository.
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
            <p v-if="organizationRepoOwnerError" class="mt-1 text-xs text-red-400">{{ organizationRepoOwnerError }}</p>
          </div>
          <div>
            <label class="field-label">Name</label>
            <Input v-model="orgRepoForm.name" placeholder="ops-console" />
            <p v-if="organizationRepoNameError" class="mt-1 text-xs text-red-400">{{ organizationRepoNameError }}</p>
          </div>
          <div>
            <label class="field-label">Description</label>
            <Textarea v-model="orgRepoForm.description" placeholder="Shared repo for deployment workflows" />
          </div>
          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <label class="field-label">Visibility</label>
              <Select v-model="orgRepoForm.visibility">
                <option value="private">Private</option>
                <option value="public">Public</option>
              </Select>
            </div>
            <div>
              <label class="field-label">Default branch</label>
              <Input v-model="orgRepoForm.default_branch" />
              <p v-if="organizationRepoBranchError" class="mt-1 text-xs text-red-400">{{ organizationRepoBranchError }}</p>
            </div>
          </div>
          <p v-if="organizationRepoServerError" class="text-sm text-red-400">{{ organizationRepoServerError }}</p>
          <Button :disabled="createRepository.isPending.value" variant="secondary" @click="handleCreateOrganizationRepository">
            <BookCopy class="size-4" />
            Create shared repository
          </Button>
        </template>
      </Card>
    </div>
  </div>
</template>
