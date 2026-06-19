<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import ContextPanel from '@/components/ContextPanel.vue'
import DetailTabs from '@/components/DetailTabs.vue'
import TimelineItem from '@/components/TimelineItem.vue'
import ConsoleView from '@/components/ConsoleView.vue'
import DiffPanel from '@/components/DiffPanel.vue'
import FilesView from '@/components/FilesView.vue'
import AskAgentBar from '@/components/AskAgentBar.vue'
import TaskActionsSheet from '@/components/TaskActionsSheet.vue'

const route = useRoute()
const router = useRouter()
const store = useAppStore()
const tab = ref('timeline')
const selectedDiff = ref('')
const showActions = ref(false)

const agent = computed(() => store.agents.find((a) => a.id === route.params.id))
const timelineEvents = computed(() =>
  store.events.filter((e) => e.type !== 'ConsoleOutput' && e.type !== 'Heartbeat'),
)

onMounted(async () => {
  await store.loadAgents()
  const taskId = (route.query.task as string) || agent.value?.taskId
  if (taskId) await store.selectTask(taskId)
  if (store.diffs[0]) selectedDiff.value = store.diffs[0].path
})

watch(() => store.diffs, (d) => {
  if (!selectedDiff.value && d[0]) selectedDiff.value = d[0].path
}, { deep: true })

async function handleAction(name: string) {
  showActions.value = false
  if (name === 'logs') {
    tab.value = 'console'
    return
  }
  await store.taskAction(name)
}
</script>

<template>
  <div class="page-shell page-shell--detail">
    <header class="mb-4 flex items-center gap-3">
      <button type="button" class="glass flex h-10 w-10 items-center justify-center rounded-xl" @click="router.back()">←</button>
      <div class="min-w-0 flex-1">
        <h1 class="truncate text-lg font-bold">{{ agent?.name ?? 'Agent' }}</h1>
        <p class="truncate text-sm text-muted">{{ agent?.workspace }}</p>
      </div>
      <span
        class="rounded-full border px-2.5 py-1 text-[11px] font-medium capitalize"
        :class="agent?.status === 'running' ? 'border-success/30 bg-success/15 text-success' : 'border-white/10 bg-white/5 text-muted'"
      >
        {{ agent?.status ?? 'idle' }}
      </span>
    </header>

    <p v-if="store.lastNotice" class="mb-3 rounded-xl border border-accent/25 bg-accent/10 px-3 py-2 text-sm text-accent-2">
      {{ store.lastNotice }}
    </p>

    <ContextPanel :agent="agent" :task="store.activeTask" />
    <DetailTabs v-model="tab" />

    <div class="min-h-0 flex-1">
      <ul v-if="tab === 'timeline'" class="pb-4">
        <TimelineItem
          v-for="(ev, i) in timelineEvents"
          :key="ev.id"
          :event="ev"
          :is-last="i === timelineEvents.length - 1"
        />
        <p v-if="!timelineEvents.length" class="py-8 text-center text-sm text-muted">No events yet</p>
      </ul>
      <ConsoleView v-else-if="tab === 'console'" :lines="store.consoleLines" />
      <FilesView v-else-if="tab === 'files'" :files="store.files" />
      <DiffPanel v-else :diffs="store.diffs" :selected="selectedDiff" @select="selectedDiff = $event" />
    </div>

    <AskAgentBar @actions="showActions = true" />
    <TaskActionsSheet :open="showActions" @close="showActions = false" @action="handleAction" />
  </div>
</template>

<style scoped>
.page-shell--detail {
  min-height: 100%;
  padding-bottom: 1rem;
}
</style>