<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import AppHeader from '@/components/AppHeader.vue'
import QuickActions from '@/components/QuickActions.vue'
import AgentCard from '@/components/AgentCard.vue'
import LiveActivityPanel from '@/components/LiveActivityPanel.vue'

const store = useAppStore()
const router = useRouter()

onMounted(async () => {
  store.connectWS()
  await store.refresh()
  if (!store.activeTaskId && store.demoTaskId) {
    await store.selectTask(store.demoTaskId)
  }
})

function openAgent(id: string, taskId?: string) {
  router.push({ name: 'agent', params: { id }, query: taskId ? { task: taskId } : {} })
}
</script>

<template>
  <div class="page-shell">
    <AppHeader title="MobAgent bot" />

    <div v-if="store.offline" class="mb-3 rounded-xl bg-warning/10 px-3 py-2 text-center text-xs text-warning">
      Reconnecting…
    </div>

    <section class="mb-4 shrink-0">
      <h2 class="section-title">Agents</h2>
      <div class="space-y-3">
        <AgentCard
          v-for="agent in store.agents"
          :key="agent.id"
          :agent="agent"
          @click="openAgent(agent.id, agent.taskId)"
        />
      </div>
    </section>

    <LiveActivityPanel />

    <section class="shrink-0 pb-1">
      <QuickActions />
    </section>
  </div>
</template>
