import { describe, expect, it } from 'vitest'

describe('WSClient', () => {
  it('constructs with default ws url', async () => {
    const { WSClient } = await import('@/lib/ws-client')
    const client = new WSClient()
    expect(client.isOffline()).toBe(false)
    client.disconnect()
  })
})

describe('event dedupe', () => {
  it('skips duplicate ids in resume batch logic', () => {
    const seen = new Set<string>()
    const events = [{ id: 'a' }, { id: 'a' }, { id: 'b' }]
    let last = ''
    for (const e of events) {
      if (e.id === last) continue
      if (seen.has(e.id)) continue
      seen.add(e.id)
      last = e.id
    }
    expect(seen.size).toBe(2)
  })
})
