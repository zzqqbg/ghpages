<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import AgentPicker from '@/components/AgentPicker.vue'
import FormSelect from '@/components/FormSelect.vue'

const store = useAppStore()
const router = useRouter()
const prompt = ref('Implement user login, support email/password, use JWT...')
const agentId = ref('')
const workspace = ref('')
const branch = ref('')
const priority = ref('medium')
const localError = ref('')
const maxLen = 500

const charCount = computed(() => prompt.value.length)
const submitting = computed(() => store.loading)

const workspaceOptions = computed(() =>
  store.workspaces.map((w) => ({ value: w, label: w })),
)

const selectedAgent = computed(() => store.agents.find((a) => a.id === agentId.value))

watch(
  () => store.agents,
  (list) => {
    if (!list.length) return
    if (!list.some((a) => a.id === agentId.value)) {
      agentId.value = list[0].id
    }
  },
  { immediate: true, deep: true },
)

watch(selectedAgent, (agent) => {
  if (!agent) return
  if (agent.workspace && store.workspaces.includes(agent.workspace)) {
    workspace.value = agent.workspace
  }
  if (agent.branch && !branch.value) {
    branch.value = agent.branch
  }
})

onMounted(async () => {
  await store.loadBootstrap()
  if (!workspace.value && store.workspaces[0]) {
    workspace.value = store.workspaces[0]
  }
})

async function submit() {
  localError.value = ''
  if (!prompt.value.trim()) {
    localError.value = 'Please enter a task description'
    return
  }
  if (!agentId.value) {
    localError.value = 'Please select an agent'
    return
  }
  if (!workspace.value) {
    localError.value = 'Please select a workspace'
    return
  }
  try {
    const task = await store.createTask({
      prompt: prompt.value.trim(),
      agentId: agentId.value,
      workspace: workspace.value,
      branch: branch.value,
      priority: priority.value,
    })
    window.Telegram?.WebApp?.HapticFeedback?.impactOccurred('medium')
    await router.push({
      name: 'agent',
      params: { id: agentId.value },
      query: { task: task.id },
    })
  } catch (e) {
    localError.value = e instanceof Error ? e.message : 'Failed to start task'
    window.Telegram?.WebApp?.showAlert?.(localError.value)
  }
}
</script>

<template>
  <div class="page-shell">
    <header class="mb-5 flex items-center justify-between">
      <h1 class="text-2xl font-bold">New Task</h1>
      <button type="button" class="glass flex h-10 w-10 items-center justify-center rounded-xl text-lg" @click="router.back()">✕</button>
    </header>

    <p v-if="store.loadError" class="mb-3 rounded-xl bg-warning/10 px-3 py-2 text-xs text-warning">
      {{ store.loadError }} — 请确认后端 API 已启动（./r dev）
    </p>

    <p v-if="localError" class="mb-4 rounded-xl border border-danger/30 bg-danger/10 px-3 py-2 text-sm text-danger">
      {{ localError }}
    </p>

    <form class="space-y-5" @submit.prevent="submit">
      <label class="block">
        <div class="mb-2 flex items-center justify-between">
          <span class="text-sm text-muted">Task Description</span>
          <span class="text-xs text-muted">{{ charCount }}/{{ maxLen }}</span>
        </div>
        <textarea
          v-model="prompt"
          :maxlength="maxLen"
          rows="5"
          class="glass w-full rounded-2xl p-3 text-sm outline-none focus:ring-2 focus:ring-accent/40"
        />
      </label>

      <div>
        <span class="mb-2 block text-sm text-muted">Agent</span>
        <AgentPicker v-model="agentId" :agents="store.agents" />
        <p v-if="selectedAgent" class="mt-2 text-xs text-muted">
          Selected: <span class="text-accent">{{ selectedAgent.name }}</span>
        </p>
      </div>

      <div>
        <span class="mb-2 block text-sm text-muted">Priority</span>
        <div class="flex gap-2">
          <button
            v-for="p in ['high', 'medium', 'low']"
            :key="p"
            type="button"
            class="min-h-11 flex-1 rounded-xl capitalize transition-colors"
            :class="priority === p ? (p === 'high' ? 'bg-danger' : p === 'low' ? 'bg-idle' : 'bg-accent') : 'glass text-muted'"
            @click="priority = p"
          >
            {{ p }}
          </button>
        </div>
      </div>

      <div>
        <span class="mb-2 block text-sm text-muted">Workspace</span>
        <FormSelect
          v-model="workspace"
          :options="workspaceOptions"
          title="Select Workspace"
          placeholder="Choose workspace"
        />
      </div>

      <label class="block">
        <span class="mb-2 block text-sm text-muted">Branch (Optional)</span>
        <input
          v-model="branch"
          placeholder="feature/user-login"
          class="glass min-h-12 w-full rounded-2xl px-3 text-sm outline-none focus:ring-2 focus:ring-accent/40"
        />
      </label>

      <button
        type="submit"
        class="min-h-14 w-full rounded-2xl bg-gradient-to-r from-accent to-accent-2 text-base font-semibold shadow-lg shadow-accent/25 disabled:opacity-60"
        :disabled="submitting || !store.agents.length || !store.workspaces.length"
      >
        {{ submitting ? 'Starting…' : 'Start Task' }}
      </button>
    </form>
  </div>
</template>
