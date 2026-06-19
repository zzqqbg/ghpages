package models

type WorkerAgentDecl struct {
	ID        string    `json:"id"`
	Type      AgentType `json:"type"`
	Name      string    `json:"name"`
	Workspace string    `json:"workspace"`
	Branch    string    `json:"branch,omitempty"`
}

type WorkerWorkspaceDecl struct {
	Name string `json:"name"`
	Path string `json:"path,omitempty"`
}

type WorkerRegister struct {
	WorkerID   string                `json:"workerId"`
	Hostname   string                `json:"hostname,omitempty"`
	Version    string                `json:"version,omitempty"`
	Agents     []WorkerAgentDecl     `json:"agents"`
	Workspaces []WorkerWorkspaceDecl `json:"workspaces"`
}

type WorkerAuthOK struct {
	AccountID string `json:"accountId"`
	WorkerID  string `json:"workerId"`
	ServerTS  int64  `json:"serverTs"`
}

type WorkerRegisterOK struct {
	ServerTS int64 `json:"serverTs"`
}

type WorkerRunTask struct {
	TaskID    string    `json:"taskId"`
	SessionID string    `json:"sessionId"`
	AgentID   string    `json:"agentId"`
	AgentType AgentType `json:"agentType"`
	Prompt    string    `json:"prompt"`
	Workspace string    `json:"workspace"`
	Branch    string    `json:"branch,omitempty"`
	Priority  string    `json:"priority,omitempty"`
}

type WorkerCancelTask struct {
	TaskID string `json:"taskId"`
}

type WorkerTaskEvent struct {
	Event AgentEvent `json:"event"`
}

type WorkerConsoleLine struct {
	TaskID    string `json:"taskId"`
	SessionID string `json:"sessionId"`
	Line      string `json:"line"`
}

type WorkerTaskDone struct {
	TaskID    string `json:"taskId"`
	SessionID string `json:"sessionId"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

type WorkerTaskAck struct {
	TaskID string `json:"taskId"`
}
