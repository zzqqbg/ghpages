<script setup lang="ts">
import type { Agent } from '@/types'
import { formatCost, formatDuration, formatTokens } from '@/lib/api'

defineProps<{ agent: Agent }>()
defineEmits<{ click: [] }>()

const agentIcon: Record<string, string> = {
  cursor: '⚡',
  'claude-code': '🧠',
  codex: '🤖',
  'gemini-cli': '✨',
  aider: '🔧',
  openhands: '🛠',
}

const statusStyle: Record<string, string> = {
  running: 'bg-success/15 text-success border-success/30',
  idle: 'bg-idle/15 text-idle border-idle/30',
  stopped: 'bg-danger/15 text-danger border-danger/30',
}
</script>

<template>
  <article
    class="glass-strong clickable rounded-2xl p-4 transition-opacity duration-150 hover:opacity-95 active:opacity-80"
    @click="$emit('click')"
  >
    <div class="flex items-start gap-3">
      <div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-panel-3 text-xl">
        {{ agentIcon[agent.type] ?? '⚡' }}
      </div>
      <div class="min-w-0 flex-1">
        <div class="flex items-start justify-between gap-2">
          <div class="min-w-0">
            <h3 class="truncate font-semibold">{{ agent.name }}</h3>
            <p class="truncate text-sm text-muted">{{ agent.workspace }}</p>
          </div>
          <span
            class="shrink-0 rounded-full border px-2.5 py-0.5 text-[11px] font-medium capitalize"
            :class="statusStyle[agent.status] ?? statusStyle.idle"
          >
            {{ agent.status }}
          </span>
        </div>
        <p v-if="agent.currentTask" class="mt-2 truncate text-sm text-white/90">{{ agent.currentTask }}</p>
        <p v-if="agent.status === 'running' && agent.currentStage" class="mt-1 truncate text-xs text-accent-2">
          {{ agent.currentStage }}
        </p>
      </div>
    </div>

    <div v-if="agent.status === 'running'" class="mt-4 space-y-2">
      <div class="flex items-center justify-between text-xs text-muted">
        <span>Progress</span>
        <span>{{ agent.progress }}%</span>
      </div>
      <div class="h-2 overflow-hidden rounded-full bg-white/8">
        <div
          class="h-full rounded-full bg-gradient-to-r from-accent to-accent-2 transition-all duration-300"
          :style="{ width: `${agent.progress}%` }"
        />
      </div>
      <div class="flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted">
        <span>⏱ {{ formatDuration(agent.elapsedSec) }}</span>
        <span>🪙 {{ formatTokens(agent.tokens) }}</span>
        <span>💲 {{ formatCost(agent.costUsd) }}</span>
      </div>
    </div>
  </article>
</template>
