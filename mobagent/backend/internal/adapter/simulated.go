package adapter

import (
	"context"
	"time"

	"github.com/ghpages/mobagent/backend/internal/models"
)

type SimulatedAdapter struct {
	agentType models.AgentType
	label     string
}

func NewSimulated(t models.AgentType, label string) *SimulatedAdapter {
	return &SimulatedAdapter{agentType: t, label: label}
}

func (s *SimulatedAdapter) Type() models.AgentType { return s.agentType }
func (s *SimulatedAdapter) Name() string           { return s.label }

func (s *SimulatedAdapter) Run(ctx context.Context, task *models.Task, emit func(line string)) error {
	script := []struct {
		line string
		wait time.Duration
	}{
		{"Task started: " + task.Title, 200 * time.Millisecond},
		{"Reading package.json", 300 * time.Millisecond},
		{"Reading tsconfig.json", 200 * time.Millisecond},
		{"Reading project structure", 400 * time.Millisecond},
		{"Searching login related modules", 500 * time.Millisecond},
		{"Planning implementation", 800 * time.Millisecond},
		{"Editing src/auth/login.ts", 600 * time.Millisecond},
		{"Editing src/auth/jwt.ts", 500 * time.Millisecond},
		{"Creating src/types/user.d.ts", 400 * time.Millisecond},
		{"Running npm run test", 700 * time.Millisecond},
		{"PASS src/auth/login.test.ts", 100 * time.Millisecond},
		{"PASS src/auth/jwt.test.ts", 100 * time.Millisecond},
		{"Reviewing changes", 500 * time.Millisecond},
		{"git commit -m feat: add login", 300 * time.Millisecond},
		{"Task finished successfully", 200 * time.Millisecond},
	}
	for _, step := range script {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		emit(step.line)
		timer := time.NewTimer(step.wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return nil
}

func DefaultSimulatedAdapters() []Adapter {
	return []Adapter{
		NewSimulated(models.AgentCursor, "Cursor Agent"),
		NewSimulated(models.AgentClaudeCode, "Claude Code"),
		NewSimulated(models.AgentCodex, "Codex"),
		NewSimulated(models.AgentGeminiCLI, "Gemini CLI"),
		NewSimulated(models.AgentAider, "Aider"),
		NewSimulated(models.AgentOpenHands, "OpenHands"),
	}
}
