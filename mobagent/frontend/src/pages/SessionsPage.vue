<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import AppHeader from '@/components/AppHeader.vue'
import { formatCost, formatDuration } from '@/lib/api'

const store = useAppStore()
const router = useRouter()

const sessions = computed(() =>
  [...store.tasks].sort((a, b) => new Date(b.updatedAt ?? 0).getTime() - new Date(a.updatedAt ?? 0).getTime()),
)

onMounted(async () => {
  await store.refresh()
})

function openSession(taskId: string, agentId: string) {
  router.push({ name: 'agent', params: { id: agentId }, query: { task: taskId } })
}

const statusColor: Record<string, string> = {
  running: 'text-success',
  completed: 'text-accent-2',
  paused: 'text-warning',
  stopped: 'text-muted',
  failed: 'text-danger',
}
</script>

<template>
  <div class="page-shell">
    <AppHeader title="Sessions" />

    <section class="flex flex-1 flex-col gap-3">
      <article
        v-for="task in sessions"
        :key="task.id"
        class="glass-strong clickable rounded-2xl p-4 transition-opacity duration-150 hover:opacity-95 active:opacity-80"
        @click="openSession(task.id, task.agentId)"
      >
        <div class="flex items-start justify-between gap-2">
          <div class="min-w-0">
            <h3 class="truncate font-semibold">{{ task.title }}</h3>
            <p class="truncate text-sm text-muted">{{ task.workspace }} · {{ task.branch || 'main' }}</p>
          </div>
          <span class="shrink-0 text-xs capitalize" :class="statusColor[task.status] ?? 'text-muted'">{{ task.status }}</span>
        </div>
        <div class="mt-3 flex flex-wrap gap-3 text-xs text-muted">
          <span>{{ task.progress }}%</span>
          <span>{{ formatDuration(task.elapsedSec) }}</span>
          <span>{{ formatCost(task.costUsd) }}</span>
        </div>
        <div class="mt-2 h-1.5 overflow-hidden rounded-full bg-white/8">
          <div class="h-full rounded-full bg-accent transition-all" :style="{ width: `${task.progress}%` }" />
        </div>
      </article>

      <div v-if="!sessions.length" class="glass flex flex-1 flex-col items-center justify-center rounded-2xl py-12">
        <p class="text-muted">No sessions yet</p>
        <button type="button" class="mt-4 rounded-xl bg-accent px-4 py-2 text-sm" @click="router.push('/task/new')">Create Task</button>
      </div>

      <div v-else class="glass mt-auto rounded-2xl p-4 text-sm text-muted">
        <p class="font-medium text-white/85">Session tips</p>
        <p class="mt-2 leading-relaxed">Tap a session to open Timeline, Console, Files and Diff for that task.</p>
      </div>
    </section>
  </div>
</template>
