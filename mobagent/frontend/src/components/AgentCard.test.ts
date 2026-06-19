import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import AgentCard from '@/components/AgentCard.vue'

describe('AgentCard', () => {
  it('renders agent name and progress', () => {
    const wrapper = mount(AgentCard, {
      props: {
        agent: {
          id: '1', name: 'Cursor Agent', type: 'cursor', status: 'running',
          workspace: 'casino-platform', progress: 67, elapsedSec: 738,
          tokens: 2400, costUsd: 0.023, cpu: 42, memoryMb: 512,
          currentStage: 'Reading project structure...',
        },
      },
    })
    expect(wrapper.text()).toContain('Cursor Agent')
    expect(wrapper.text()).toContain('67%')
    expect(wrapper.text()).toContain('casino-platform')
  })
})
