package models

import "time"

type EventType string

const (
	EventTaskStarted           EventType = "TaskStarted"
	EventPlanning              EventType = "Planning"
	EventReadingProject        EventType = "ReadingProject"
	EventReadingFile           EventType = "ReadingFile"
	EventSearching             EventType = "Searching"
	EventEditing               EventType = "Editing"
	EventCreatingFile          EventType = "CreatingFile"
	EventUpdatingFile          EventType = "UpdatingFile"
	EventRemovingFile          EventType = "RemovingFile"
	EventRunningCommand        EventType = "RunningCommand"
	EventInstallingDeps        EventType = "InstallingDependencies"
	EventBuilding              EventType = "Building"
	EventRunningTests          EventType = "RunningTests"
	EventReviewing             EventType = "Reviewing"
	EventGitCommit             EventType = "GitCommit"
	EventGitPush               EventType = "GitPush"
	EventNeedUserInput         EventType = "NeedUserInput"
	EventWarning               EventType = "Warning"
	EventError                 EventType = "Error"
	EventFinished              EventType = "Finished"
	EventFailed                EventType = "Failed"
	EventDiffUpdated           EventType = "DiffUpdated"
	EventConsoleOutput         EventType = "ConsoleOutput"
	EventCostUpdated           EventType = "CostUpdated"
	EventResourceUpdated       EventType = "ResourceUpdated"
	EventHeartbeat             EventType = "Heartbeat"
)

type EventStatus string

const (
	StatusPending   EventStatus = "pending"
	StatusActive    EventStatus = "active"
	StatusCompleted EventStatus = "completed"
	StatusFailed    EventStatus = "failed"
)

type AgentEvent struct {
	ID        string                 `json:"id"`
	TaskID    string                 `json:"taskId"`
	SessionID string                 `json:"sessionId"`
	Timestamp time.Time              `json:"timestamp"`
	Agent     string                 `json:"agent"`
	Workspace string                 `json:"workspace"`
	Type      EventType              `json:"type"`
	Status    EventStatus            `json:"status"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

type WSClientHello struct {
	SessionID   string `json:"sessionId"`
	LastEventID string `json:"lastEventId,omitempty"`
	TaskID      string `json:"taskId,omitempty"`
}

type WSHelloAck struct {
	SessionID string `json:"sessionId"`
	ServerTS  int64  `json:"serverTs"`
}

type WSResumeAck struct {
	TaskID    string       `json:"taskId"`
	Events    []AgentEvent `json:"events"`
	LastEvent string       `json:"lastEventId"`
}
