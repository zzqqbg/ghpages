<script setup lang="ts">
defineProps<{ name: 'home' | 'sessions' | 'console' | 'settings'; active?: boolean }>()
</script>

<template>
  <svg
    viewBox="0 0 24 24"
    fill="none"
    class="tab-icon"
    :class="active ? 'tab-icon--active' : 'tab-icon--idle'"
    aria-hidden="true"
  >
    <template v-if="name === 'home'">
      <path
        d="M5 11.2 12 5l7 6.2V19a1.2 1.2 0 0 1-1.2 1.2H14v-5.8h-4V20.2H6.2A1.2 1.2 0 0 1 5 19v-7.8Z"
        :fill="active ? 'currentColor' : 'none'"
        stroke="currentColor"
        stroke-width="1.5"
        stroke-linejoin="round"
      />
    </template>

    <template v-else-if="name === 'sessions'">
      <path d="M5 7h14M5 12h14M5 17h14" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
      <circle v-if="active" cx="18" cy="7" r="1.2" fill="currentColor" />
      <circle v-if="active" cx="14" cy="12" r="1.2" fill="currentColor" />
    </template>

    <template v-else-if="name === 'console'">
      <rect x="4" y="5.5" width="16" height="13" rx="2" stroke="currentColor" stroke-width="1.5" />
      <path d="M8 10.5 10.2 12.8 15.5 8.2" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
      <path d="M8 16h8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" opacity="0.55" />
    </template>

    <template v-else>
      <circle cx="12" cy="12" r="2.6" stroke="currentColor" stroke-width="1.5" :fill="active ? 'currentColor' : 'none'" />
      <path
        d="M12 4v2M12 18v2M4 12H2M22 12h-2M6.3 6.3 4.9 4.9M19.1 19.1l-1.4-1.4M17.7 6.3l1.4-1.4M6.3 17.7l-1.4 1.4"
        stroke="currentColor"
        stroke-width="1.5"
        stroke-linecap="round"
      />
    </template>
  </svg>
</template>

<style scoped>
.tab-icon {
  width: 1.375rem;
  height: 1.375rem;
  transition: color 0.15s ease;
}

.tab-icon--idle {
  color: color-mix(in srgb, white 38%, transparent);
}

.tab-icon--active {
  color: var(--color-accent);
}
</style>
