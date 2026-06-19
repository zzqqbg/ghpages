<script setup lang="ts">
defineProps<{ title?: string; showSearch?: boolean }>()
</script>

<template>
  <header class="mb-5 flex items-center justify-between gap-3">
    <div class="flex min-w-0 items-center gap-3">
      <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-accent to-accent-2 text-lg">🤖</div>
      <div class="min-w-0">
        <p class="text-xs text-muted">Telegram Mini App</p>
        <h1 class="truncate text-lg font-bold">{{ title ?? 'MobAgent bot' }}</h1>
      </div>
    </div>
    <div class="flex items-center gap-2">
      <button v-if="showSearch !== false" type="button" class="glass flex h-10 w-10 items-center justify-center rounded-xl text-muted" aria-label="Search">🔍</button>
      <div class="flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-accent/80 to-purple/80 text-sm font-semibold">U</div>
    </div>
  </header>
</template>
