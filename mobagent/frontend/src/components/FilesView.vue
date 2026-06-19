<script setup lang="ts">
import type { TaskFile } from '@/types'

defineProps<{ files: TaskFile[] }>()

const statusIcon: Record<string, string> = {
  modified: '✏️',
  added: '➕',
  removed: '🗑',
  read: '📄',
}
</script>

<template>
  <div class="space-y-2">
    <button
      v-for="f in files"
      :key="f.path"
      type="button"
      class="glass flex w-full items-center gap-3 rounded-xl px-3 py-3 text-left"
    >
      <span>{{ statusIcon[f.status] ?? '📄' }}</span>
      <div class="min-w-0 flex-1">
        <p class="truncate font-mono text-sm">{{ f.path }}</p>
        <p class="text-xs capitalize text-muted">{{ f.status }}</p>
      </div>
      <span v-if="f.additions != null" class="shrink-0 font-mono text-xs">
        <span class="text-success">+{{ f.additions }}</span>
        <span class="text-danger"> -{{ f.deletions ?? 0 }}</span>
      </span>
    </button>
    <p v-if="!files.length" class="py-8 text-center text-sm text-muted">No files scanned yet</p>
  </div>
</template>
