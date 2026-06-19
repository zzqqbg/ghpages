<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import TabIcon from '@/components/TabIcon.vue'

const route = useRoute()
const tabs = [
  { to: '/', label: 'Home', icon: 'home' as const },
  { to: '/sessions', label: 'Sessions', icon: 'sessions' as const },
  { to: '/console', label: 'Console', icon: 'console' as const },
  { to: '/settings', label: 'Settings', icon: 'settings' as const },
]

const hidden = computed(() => ['agent', 'new-task'].includes(String(route.name ?? '')))

function isActive(path: string) {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}
</script>

<template>
  <nav v-if="!hidden" class="tab-bar" aria-label="Main navigation">
    <RouterLink
      v-for="tab in tabs"
      :key="tab.to"
      :to="tab.to"
      class="tab-bar__item"
      :class="{ 'is-active': isActive(tab.to) }"
      :aria-current="isActive(tab.to) ? 'page' : undefined"
    >
      <TabIcon :name="tab.icon" :active="isActive(tab.to)" />
      <span class="tab-bar__label">{{ tab.label }}</span>
    </RouterLink>
  </nav>
</template>

<style scoped>
.tab-bar {
  flex-shrink: 0;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  align-items: stretch;
  min-height: 3.5rem;
  padding: 0.375rem 0.5rem calc(0.375rem + env(safe-area-inset-bottom, 0px));
  background: color-mix(in srgb, var(--color-panel) 98%, var(--color-surface));
  border-top: 1px solid color-mix(in srgb, white 7%, transparent);
}

.tab-bar__item {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.1875rem;
  padding: 0.375rem 0.25rem;
  border-radius: 0.625rem;
  text-decoration: none;
  cursor: pointer;
  user-select: none;
  -webkit-tap-highlight-color: transparent;
  transition: opacity 0.15s ease;
}

.tab-bar__item:hover {
  opacity: 0.92;
}

.tab-bar__item:active {
  opacity: 0.72;
}

.tab-bar__item.is-active::after {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  width: 1.125rem;
  height: 2px;
  border-radius: 999px;
  background: var(--color-accent);
  transform: translateX(-50%);
}

.tab-bar__label {
  font-size: 0.625rem;
  font-weight: 500;
  line-height: 1;
  letter-spacing: 0.02em;
  color: color-mix(in srgb, white 34%, transparent);
  transition: color 0.15s ease;
}

.tab-bar__item.is-active .tab-bar__label {
  color: var(--color-accent);
  font-weight: 600;
}
</style>
