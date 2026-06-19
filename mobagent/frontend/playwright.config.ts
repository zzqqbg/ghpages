import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: '../e2e',
  fullyParallel: true,
  retries: 0,
  use: {
    baseURL: 'http://127.0.0.1:5175',
    trace: 'on-first-retry',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: [
    {
      command: 'cd ../backend && DATA_DIR=../data WORKSPACES_ROOT=../data/workspaces PORT=8790 go run -buildvcs=false ./cmd/server/',
      url: 'http://127.0.0.1:8790/health',
      reuseExistingServer: !process.env.CI,
      timeout: 120000,
    },
    {
      command: 'pnpm dev --host 127.0.0.1 --port 5175',
      url: 'http://127.0.0.1:5175',
      reuseExistingServer: !process.env.CI,
      timeout: 120000,
    },
  ],
})
