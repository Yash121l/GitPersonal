<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { computed, reactive, ref } from 'vue'

import EmptyState from '@/components/empty/EmptyState.vue'
import PageHeader from '@/components/app/PageHeader.vue'
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

const memberForms = reactive<Record<string, { username: string; role: string }>>({})
const errorMessage = ref('')

function formFor(slug: string) {
  if (!memberForms[slug]) {
    memberForms[slug] = { username: '', role: 'member' }
  }
  return memberForms[slug]
}

const addMember = useMutation({
  mutationFn: ({ slug, username, role }: { slug: string; username: string; role: string }) =>
    api.addOrganizationMember(slug, { username, role }),
  onSuccess: async (_, variables) => {
    formFor(variables.slug).username = ''
    formFor(variables.slug).role = 'member'
    errorMessage.value = ''
    await queryClient.invalidateQueries({ queryKey: ['organizations'] })
  },
})

const organizations = computed(() => organizationsQuery.data.value ?? [])

async function handleAddMember(slug: string) {
  const form = formFor(slug)
  errorMessage.value = ''
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
  <div class="space-y-6">
    <PageHeader
      eyebrow="Organizations"
      title="Shared ownership stays visible."
      description="Organization roles are first-class and flow through the browser, API, HTTP Git, and SSH Git. This screen keeps that model explicit so access remains understandable."
    />

    <div v-if="organizations.length" class="grid gap-4 xl:grid-cols-2">
      <Card v-for="organization in organizations" :key="organization.organization_slug" class="space-y-4">
        <div class="flex flex-wrap items-center gap-2">
          <Badge variant="accent">{{ organization.role }}</Badge>
          <Badge>{{ organization.organization_slug }}</Badge>
        </div>
        <div>
          <h3 class="text-2xl font-semibold text-zinc-50">
            {{ organization.organization_display_name }}
          </h3>
          <p class="mt-2 text-sm text-zinc-400">
            Add members from the browser when your role allows it. Forge will enforce the same permission rules on repository creation and Git access.
          </p>
        </div>

        <div class="grid gap-4 md:grid-cols-[1fr_160px_140px]">
          <div>
            <label class="field-label">Username</label>
            <Input v-model="formFor(organization.organization_slug).username" />
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

    <EmptyState
      v-else
      eyebrow="No Organizations"
      title="This account is not part of any organization yet."
      description="Create an organization from the repositories screen, or wait to be invited into a shared namespace."
    />

    <div
      v-if="errorMessage"
      class="rounded-md border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-300"
    >
      {{ errorMessage }}
    </div>
  </div>
</template>
