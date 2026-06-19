import { createRouter, createWebHistory } from 'vue-router'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: () => import('@/pages/HomePage.vue') },
    { path: '/sessions', name: 'sessions', component: () => import('@/pages/SessionsPage.vue') },
    { path: '/console', name: 'console', component: () => import('@/pages/ConsolePage.vue') },
    { path: '/settings', name: 'settings', component: () => import('@/pages/SettingsPage.vue') },
    { path: '/agent/:id', name: 'agent', component: () => import('@/pages/AgentDetailPage.vue') },
    { path: '/task/new', name: 'new-task', component: () => import('@/pages/NewTaskPage.vue') },
  ],
})
