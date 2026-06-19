<script setup lang="ts">
import { onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import AppHeader from '@/components/AppHeader.vue'
import ConsoleView from '@/components/ConsoleView.vue'
import ContextPanel from '@/components/ContextPanel.vue'

const store = useAppStore()

onMounted(async () => {
  await store.refresh()
  const taskId = store.activeTaskId || store.demoTaskId
  if (taskId) await store.selectTask(taskId)
})
</script>

<template>
  <div class="page-shell page-shell--fill">
    <AppHeader title="Console" :show-search="false" />
    <ContextPanel :agent="store.agents.find(a => a.id === store.activeTask?.agentId)" :task="store.activeTask" />
    <ConsoleView class="console-grow" :lines="store.consoleLines" placeholder="Ask Agent or type command..." />
  </div>
</template>

<style scoped>
.page-shell--fill {
  min-height: 100%;
}

.console-grow {
  flex: 1;
  min-height: 16rem;
}
</style>
