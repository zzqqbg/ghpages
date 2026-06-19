import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import AgentPicker from '@/components/AgentPicker.vue'

describe('AgentPicker', () => {
  it('emits selected agent id on click', async () => {
    const wrapper = mount(AgentPicker, {
      props: {
        modelValue: 'cursor-1',
        agents: [
          { id: 'cursor-1', name: 'Cursor Agent', type: 'cursor', status: 'running', workspace: 'a', progress: 0, elapsedSec: 0, tokens: 0, costUsd: 0, cpu: 0, memoryMb: 0 },
          { id: 'claude-1', name: 'Claude Code', type: 'claude-code', status: 'idle', workspace: 'b', progress: 0, elapsedSec: 0, tokens: 0, costUsd: 0, cpu: 0, memoryMb: 0 },
        ],
      },
    })
    const buttons = wrapper.findAll('.agent-picker__item')
    expect(buttons.length).toBe(2)
    await buttons[1].trigger('click')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['claude-1'])
  })
})
