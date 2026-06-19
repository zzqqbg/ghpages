import { test, expect } from '@playwright/test'

test('home shows agents and quick actions', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByText('Cursor Agent')).toBeVisible()
  await expect(page.getByText('Quick Actions')).toBeVisible()
  await expect(page.getByRole('button', { name: 'New Task' })).toBeVisible()
})

test('sessions page lists demo task', async ({ page }) => {
  await page.goto('/sessions')
  await expect(page.getByText('Implement user login function')).toBeVisible()
})

test('console page shows log output', async ({ page }) => {
  await page.goto('/console')
  await expect(page.getByText('Analyzing project structure')).toBeVisible()
})

test('agent detail has four tabs', async ({ page }) => {
  await page.goto('/agent/cursor-1?task=task-demo-1')
  await expect(page.getByRole('button', { name: 'Timeline' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Console' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Files' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Diff' })).toBeVisible()
})

test('new task page selects agent inline', async ({ page }) => {
  await page.goto('/task/new')
  await page.getByRole('option', { name: /Claude Code/i }).click()
  await expect(page.getByText('Selected: Claude Code')).toBeVisible()
})
