<script setup lang="ts">
import PageHeader from '@/components/app/PageHeader.vue'
import Card from '@/components/ui/Card.vue'
</script>

<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Settings"
      title="Reserved for future repository controls."
      description="This route exists behind a feature flag so upcoming repository configuration surfaces can be added without changing the routing model."
    />

    <Card class="space-y-3">
      <p class="eyebrow">Feature Flagged</p>
      <h3 class="text-2xl font-semibold text-zinc-50">No repository settings are exposed yet.</h3>
      <p class="text-sm leading-6 text-zinc-400">
        When this feature is enabled, repository retention, branch protection, and other configuration panels can plug into the existing shell.
      </p>
    </Card>
  </div>
</template>
