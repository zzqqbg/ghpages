<script setup lang="ts">
import type { AgentEvent } from '@/types'
import { computed } from 'vue'

const props = defineProps<{ event: AgentEvent; isLast?: boolean }>()

const label = computed(() => {
  const map: Record<string, string> = {
    TaskStarted: 'Task Started',
    ReadingProject: 'Reading Project',
    ReadingFile: 'Reading File',
    Searching: 'Searching',
    Planning: 'Planning',
    Editing: 'Editing',
    RunningTests: 'Running Tests',
    Reviewing: 'Reviewing',
    Finished: 'Completed',
    Failed: 'Failed',
    GitCommit: 'Git Commit',
    GitPush: 'Git Push',
  }
  return map[props.event.type] ?? props.event.type.replace(/([A-Z])/g, ' $1').trim()
})

const state = computed(() => {
  if (props.event.status === 'completed') return 'done'
  if (props.event.status === 'active') return 'active'
  if (props.event.status === 'failed') return 'failed'
  return 'pending'
})

const files = computed(() => {
  const raw = props.event.payload?.files
  return Array.isArray(raw) ? raw as Array<{ path: string; additions?: number; deletions?: number }> : []
})
</script>

<template>
  <li class="relative flex gap-3 pb-7 last:pb-0">
    <div class="relative z-10 flex flex-col items-center">
      <div
        class="flex h-8 w-8 items-center justify-center rounded-full border-2 text-sm"
        :class="{
          'border-success bg-success/15 text-success': state === 'done',
          'border-accent bg-accent/15 text-accent': state === 'active',
          'border-danger bg-danger/15 text-danger': state === 'failed',
          'border-white/15 bg-panel-2 text-muted': state === 'pending',
        }"
      >
        <span v-if="state === 'done'">✓</span>
        <span v-else-if="state === 'active'" class="spinner inline-block h-4 w-4 rounded-full border-2 border-accent border-t-transparent" />
        <span v-else-if="state === 'failed'">!</span>
        <span v-else class="h-2 w-2 rounded-full bg-white/20" />
      </div>
      <div v-if="!isLast" class="absolute top-8 h-[calc(100%+4px)] w-px bg-white/10" />
    </div>

    <div class="min-w-0 flex-1 pt-0.5">
      <p class="font-semibold">{{ label }}</p>
      <p v-if="event.payload?.message" class="mt-0.5 text-sm text-muted">{{ event.payload.message }}</p>
      <p v-else-if="event.payload?.query" class="mt-0.5 text-sm text-muted">{{ event.payload.query }}</p>
      <ul v-if="files.length" class="mt-2 space-y-1.5">
        <li
          v-for="f in files"
          :key="f.path"
          class="flex items-center justify-between rounded-lg bg-panel-2 px-2.5 py-1.5 font-mono text-xs"
        >
          <span class="truncate">{{ f.path }}</span>
          <span class="shrink-0 pl-2">
            <span class="text-success">+{{ f.additions ?? 0 }}</span>
            <span class="text-danger"> -{{ f.deletions ?? 0 }}</span>
          </span>
        </li>
      </ul>
      <p v-else-if="event.payload?.path" class="mt-1 truncate font-mono text-xs text-accent-2">{{ event.payload.path }}</p>
    </div>
  </li>
</template>
