<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'

const store = useAppStore()
const router = useRouter()

const actions = [
  { key: 'new', label: 'New Task', icon: '➕', color: 'bg-accent text-white', route: '/task/new' },
  { key: 'resume', label: 'Resume', icon: '▶', color: 'glass text-warning', action: 'resume' },
  { key: 'pause', label: 'Pause', icon: '⏸', color: 'glass text-success', action: 'pause' },
  { key: 'stop', label: 'Stop', icon: '■', color: 'glass text-danger', action: 'stop' },
  { key: 'restart', label: 'Restart', icon: '↺', color: 'glass text-purple', action: 'restart' },
  { key: 'settings', label: 'Settings', icon: '⚙', color: 'glass text-muted', route: '/settings' },
]

async function handle(item: (typeof actions)[number]) {
  if (item.route) {
    router.push(item.route)
    return
  }
  if (item.action) {
    if (!store.activeTaskId && store.demoTaskId) {
      await store.selectTask(store.demoTaskId)
    }
    await store.taskAction(item.action)
  }
}
</script>

<template>
  <section>
    <h2 class="section-title">Quick Actions</h2>
    <div class="grid grid-cols-3 gap-2">
      <button
        v-for="item in actions"
        :key="item.key"
        type="button"
        class="clickable flex min-h-[4.5rem] flex-col items-center justify-center gap-1 rounded-2xl text-xs transition-opacity duration-150 hover:opacity-95 active:opacity-80"
        :class="item.color"
        @click="handle(item)"
      >
        <span class="text-lg">{{ item.icon }}</span>
        <span>{{ item.label }}</span>
      </button>
    </div>
  </section>
</template>
