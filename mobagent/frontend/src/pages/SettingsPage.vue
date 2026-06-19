<script setup lang="ts">
import { onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import AppHeader from '@/components/AppHeader.vue'

const store = useAppStore()

onMounted(() => store.refresh())

const features = [
  { icon: '⚡', title: 'Real-time Streaming', desc: 'Millisecond-level agent feedback via WebSocket' },
  { icon: '🧠', title: 'Smart Event Parsing', desc: 'Terminal output → structured timeline events' },
  { icon: '📝', title: 'Live Diff Preview', desc: 'Realtime file change tracking' },
  { icon: '🤖', title: 'Multi-Agent Control', desc: 'Cursor, Claude Code, Codex and more' },
  { icon: '🔒', title: 'Secure & Lightweight', desc: 'End-to-end ready, native Telegram feel' },
]
</script>

<template>
  <div class="page-shell">
    <AppHeader title="Settings" :show-search="false" />

    <section class="flex flex-1 flex-col gap-3">
      <div v-for="f in features" :key="f.title" class="glass rounded-2xl p-4">
        <div class="flex gap-3">
          <span class="text-2xl">{{ f.icon }}</span>
          <div>
            <h3 class="font-semibold">{{ f.title }}</h3>
            <p class="mt-1 text-sm text-muted">{{ f.desc }}</p>
          </div>
        </div>
      </div>

      <section class="glass mt-auto rounded-2xl p-4 text-sm">
        <p><span class="text-muted">Connection</span> · {{ store.offline ? 'Offline' : 'Connected' }}</p>
        <p class="mt-2"><span class="text-muted">Agents</span> · {{ store.agents.length }}</p>
        <p class="mt-2"><span class="text-muted">Active Task</span> · {{ store.activeTask?.title ?? 'None' }}</p>
        <p class="mt-2"><span class="text-muted">Heartbeat</span> · 10s</p>
      </section>
    </section>
  </div>
</template>
