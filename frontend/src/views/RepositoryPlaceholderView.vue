<script setup lang="ts">
import { Activity, GitPullRequest, ShieldCheck, Tag, Workflow } from '@lucide/vue'
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import PageHeader from '@/components/app/PageHeader.vue'
import Badge from '@/components/ui/Badge.vue'
import Card from '@/components/ui/Card.vue'
import { useRepositoryWorkspace } from '@/composables/useRepositoryWorkspace'

const route = useRoute()
const workspace = useRepositoryWorkspace()

const section = computed(() => {
  const path = route.path.toLowerCase()
  if (path.includes('/issues')) return { label: 'Issues', icon: Activity }
  if (path.includes('/pull')) return { label: 'Pull Requests', icon: GitPullRequest }
  if (path.includes('/actions')) return { label: 'Actions', icon: Workflow }
  if (path.includes('/security')) return { label: 'Security', icon: ShieldCheck }
  if (path.includes('/releases') || path.includes('/tags')) return { label: 'Releases', icon: Tag }
  return { label: 'Repository', icon: Activity }
})
</script>

<template>
  <div class="section-stack">
    <PageHeader :eyebrow="section.label" :title="section.label" :description="`${workspace.owner.value}/${workspace.repo.value}`">
      <template #actions>
        <Badge>{{ workspace.repositoryQuery.data.value?.repository.visibility }}</Badge>
      </template>
    </PageHeader>

    <Card class="space-y-5">
      <div class="flex size-10 items-center justify-center rounded-md border border-zinc-800 bg-zinc-900 text-zinc-300">
        <component :is="section.icon" class="size-5" />
      </div>
      <div class="space-y-2">
        <h3 class="text-lg font-semibold text-zinc-50">{{ section.label }} is routed.</h3>
        <p class="max-w-3xl text-sm leading-6 text-zinc-400">
          This screen is ready in the GitHub-style sitemap. The current backend exposes repositories, code browsing, access,
          organizations, SSH keys, and webhooks; richer {{ section.label.toLowerCase() }} data can attach here without changing navigation.
        </p>
      </div>
    </Card>
  </div>
</template>
