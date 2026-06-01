<script setup lang="ts">
import { GitBranchPlus } from '@lucide/vue'
import { RouterLink } from 'vue-router'

import PageHeader from '@/components/app/PageHeader.vue'
import WorkspaceCreationPanel from '@/components/app/WorkspaceCreationPanel.vue'
import Button from '@/components/ui/Button.vue'
</script>

<template>
  <div class="section-stack">
    <PageHeader
      eyebrow="Create"
      title="New repository"
      description="Create personal repositories, organization namespaces, and shared repositories from a dedicated workflow."
    >
      <template #actions>
        <Button :as="RouterLink" :to="{ name: 'repositories' }" variant="secondary">
          <GitBranchPlus class="size-4" />
          Browse repositories
        </Button>
      </template>
    </PageHeader>

    <WorkspaceCreationPanel />
  </div>
</template>
