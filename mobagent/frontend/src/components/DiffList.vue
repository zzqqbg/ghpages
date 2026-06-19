<script setup lang="ts">
import type { FileDiff } from '@/types'

defineProps<{ diffs: FileDiff[]; selected?: string }>()
defineEmits<{ select: [path: string] }>()
</script>

<template>
  <div class="space-y-2">
    <button
      v-for="d in diffs"
      :key="d.path"
      class="glass flex w-full items-center justify-between rounded-xl px-3 py-3 text-left transition-colors duration-200"
      :class="selected === d.path ? 'ring-2 ring-accent/50' : ''"
      @click="$emit('select', d.path)"
    >
      <span class="truncate font-mono text-sm">{{ d.path }}</span>
      <span class="shrink-0 text-xs">
        <span class="text-success">+{{ d.additions }}</span>
        <span class="text-danger"> -{{ d.deletions }}</span>
      </span>
    </button>
    <p v-if="!diffs.length" class="py-8 text-center text-sm text-muted">No file changes yet</p>
  </div>
</template>
