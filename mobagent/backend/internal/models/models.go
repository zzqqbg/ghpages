package models

import "time"

type AgentType string

const (
	AgentCursor     AgentType = "cursor"
	AgentClaudeCode AgentType = "claude-code"
	AgentCodex      AgentType = "codex"
	AgentGeminiCLI  AgentType = "gemini-cli"
	AgentAider      AgentType = "aider"
	AgentOpenHands  AgentType = "openhands"
)

type Agent struct {
	ID           string    `json:"id"`
	AccountID    string    `json:"accountId,omitempty"`
	WorkerID     string    `json:"workerId,omitempty"`
	Online       bool      `json:"online,omitempty"`
	Name         string    `json:"name"`
	Type         AgentType `json:"type"`
	Status       string    `json:"status"`
	Workspace    string    `json:"workspace"`
	CurrentStage string    `json:"currentStage,omitempty"`
	CurrentFile  string    `json:"currentFile,omitempty"`
	Progress     int       `json:"progress"`
	ElapsedSec   int       `json:"elapsedSec"`
	Tokens       int       `json:"tokens"`
	CostUSD      float64   `json:"costUsd"`
	Branch       string    `json:"branch,omitempty"`
	CPU          float64   `json:"cpu"`
	MemoryMB     float64   `json:"memoryMb"`
	CurrentTask  string    `json:"currentTask,omitempty"`
	TaskID       string    `json:"taskId,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Task struct {
	ID          string    `json:"id"`
	AccountID   string    `json:"accountId,omitempty"`
	WorkerID    string    `json:"workerId,omitempty"`
	SessionID   string    `json:"sessionId,omitempty"`
	AgentID     string    `json:"agentId"`
	AgentType   AgentType `json:"agentType"`
	Title       string    `json:"title"`
	Prompt      string    `json:"prompt"`
	Workspace   string    `json:"workspace"`
	Branch      string    `json:"branch,omitempty"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	Progress    int       `json:"progress"`
	Stage       string    `json:"stage,omitempty"`
	CurrentFile string    `json:"currentFile,omitempty"`
	Tokens      int       `json:"tokens"`
	CostUSD     float64   `json:"costUsd"`
	ElapsedSec  int       `json:"elapsedSec"`
	QueuePosition int     `json:"queuePosition,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type FileDiff struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Status    string `json:"status"`
	Diff      string `json:"diff,omitempty"`
}

type CreateTaskRequest struct {
	Prompt    string    `json:"prompt" binding:"required"`
	AgentID   string    `json:"agentId" binding:"required"`
	Workspace string    `json:"workspace"`
	Branch    string    `json:"branch"`
	Priority  string    `json:"priority"`
	AgentType AgentType `json:"agentType"`
}

type TaskAction string

const (
	ActionPause   TaskAction = "pause"
	ActionResume  TaskAction = "resume"
	ActionStop    TaskAction = "stop"
	ActionRestart TaskAction = "restart"
	ActionRetry   TaskAction = "retry"
	ActionExplain TaskAction = "explain"
	ActionReview  TaskAction = "review"
	ActionContinue TaskAction = "continue"
)

type ResourceSnapshot struct {
	CPU      float64 `json:"cpu"`
	MemoryMB float64 `json:"memoryMb"`
	Tokens   int     `json:"tokens"`
	CostUSD  float64 `json:"costUsd"`
}
