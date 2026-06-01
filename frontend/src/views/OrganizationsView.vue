<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { computed, reactive, ref } from 'vue'

import EmptyState from '@/components/empty/EmptyState.vue'
import PageHeader from '@/components/app/PageHeader.vue'
import CardSkeletonGrid from '@/components/state/CardSkeletonGrid.vue'
import ViewState from '@/components/state/ViewState.vue'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import Input from '@/components/ui/Input.vue'
import Select from '@/components/ui/Select.vue'
import { api } from '@/lib/api'

const queryClient = useQueryClient()
const organizationsQuery = useQuery({
  queryKey: ['organizations'],
  queryFn: () => api.listOrganizations(),
})

const memberForms = reactive<Record<string, { username: string; role: string; attempted: boolean }>>({})
const errorMessage = ref('')

function formFor(slug: string) {
  if (!memberForms[slug]) {
    memberForms[slug] = { username: '', role: 'member', attempted: false }
  }
  return memberForms[slug]
}

const addMember = useMutation({
  mutationFn: ({ slug, username, role }: { slug: string; username: string; role: string }) =>
    api.addOrganizationMember(slug, { username, role }),
  onSuccess: async (_, variables) => {
    formFor(variables.slug).username = ''
    formFor(variables.slug).role = 'member'
    formFor(variables.slug).attempted = false
    errorMessage.value = ''
    await queryClient.invalidateQueries({ queryKey: ['organizations'] })
  },
})

const organizations = computed(() => organizationsQuery.data.value ?? [])

function usernameError(slug: string) {
  const form = formFor(slug)
  return form.attempted && form.username.trim() === '' ? 'Username is required.' : ''
}

async function handleAddMember(slug: string) {
  const form = formFor(slug)
  form.attempted = true
  errorMessage.value = ''
  if (form.username.trim() === '') {
    return
  }

  try {
    await addMember.mutateAsync({
      slug,
      username: form.username.trim(),
      role: form.role,
    })
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Unable to add member.'
  }
}
</script>

<template>
  <div class="section-stack">
    <PageHeader
      eyebrow="Organizations"
      title="Organizations"
      description="Manage shared namespaces and member roles."
    />

    <ViewState
      :loading="organizationsQuery.isLoading.value"
      :empty="organizations.length === 0"
      empty-eyebrow="No Organizations"
      empty-title="This account is not part of any organization yet."
      empty-description="Create an organization from the repositories screen, or wait to be invited into a shared namespace."
    >
      <template #loading>
        <CardSkeletonGrid :count="4" wrapper-class="grid gap-6 xl:grid-cols-2" item-class="h-48" />
      </template>

      <template #empty>
        <EmptyState
          eyebrow="No Organizations"
          title="This account is not part of any organization yet."
          description="Create an organization from the repositories screen, or wait to be invited into a shared namespace."
        />
      </template>

      <div class="grid gap-6 xl:grid-cols-2">
        <Card v-for="organization in organizations" :key="organization.organization_slug" class="space-y-5">
          <div class="flex flex-wrap items-center gap-2">
            <Badge variant="accent">{{ organization.role }}</Badge>
            <Badge>{{ organization.organization_slug }}</Badge>
          </div>

          <div class="space-y-2">
            <h3 class="text-lg font-semibold text-zinc-50">{{ organization.organization_display_name }}</h3>
            <p class="text-sm leading-6 text-zinc-400">
              Add members and assign their role in this organization.
            </p>
          </div>

          <div class="grid gap-4 md:grid-cols-[1fr_160px_140px]">
            <div>
              <label class="field-label">Username</label>
              <Input v-model="formFor(organization.organization_slug).username" placeholder="teammate" />
              <p v-if="usernameError(organization.organization_slug)" class="mt-1 text-xs text-red-400">
                {{ usernameError(organization.organization_slug) }}
              </p>
            </div>
            <div>
              <label class="field-label">Role</label>
              <Select v-model="formFor(organization.organization_slug).role">
                <option value="member">Member</option>
                <option value="maintainer">Maintainer</option>
                <option value="owner">Owner</option>
              </Select>
            </div>
            <div class="flex items-end">
              <Button class="w-full" variant="secondary" @click="handleAddMember(organization.organization_slug)">
                Add Member
              </Button>
            </div>
          </div>
        </Card>
      </div>
    </ViewState>

    <div v-if="errorMessage" class="rounded-md border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-300">
      {{ errorMessage }}
    </div>
  </div>
</template>
