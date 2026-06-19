package store

import (
	"os"
	"sort"
	"sync"
	"time"

	"github.com/ghpages/mobagent/backend/internal/config"
	"github.com/ghpages/mobagent/backend/internal/events"
	"github.com/ghpages/mobagent/backend/internal/models"
)

type Store struct {
	mu                sync.RWMutex
	agents            map[string]*models.Agent
	tasks             map[string]*models.Task
	diffs             map[string][]models.FileDiff
	console           map[string][]string
	taskEvts          map[string][]models.AgentEvent
	workspaces        []config.Workspace
	accountWorkspaces map[string][]config.Workspace
}

func New(cfg *config.Config) *Store {
	s := &Store{
		agents:            make(map[string]*models.Agent),
		tasks:             make(map[string]*models.Task),
		diffs:             make(map[string][]models.FileDiff),
		console:           make(map[string][]string),
		taskEvts:          make(map[string][]models.AgentEvent),
		workspaces:        cfg.Workspaces,
		accountWorkspaces: make(map[string][]config.Workspace),
	}
	s.initAgents(cfg)
	s.seedDemo()
	return s
}

func NewTest() *Store {
	cfg := &config.Config{
		Workspaces: []config.Workspace{{Name: "ws", Path: "/tmp/ws"}},
		Agents: []config.AgentDef{
			{ID: "cursor-1", Type: models.AgentCursor, Name: "Cursor Agent", Workspace: "ws", Branch: "main"},
		},
	}
	s := New(cfg)
	s.agents["cursor-1"].AccountID = "demo"
	return s
}

func (s *Store) initAgents(cfg *config.Config) {
	now := time.Now().UTC()
	for _, def := range cfg.Agents {
		ws := def.Workspace
		if ws == "" && len(cfg.Workspaces) > 0 {
			ws = cfg.Workspaces[0].Name
		}
		s.agents[def.ID] = &models.Agent{
			ID: def.ID, AccountID: "demo", Name: def.Name, Type: def.Type,
			Status: "idle", Workspace: ws, Branch: def.Branch, Online: false,
			Progress: 0, CreatedAt: now, UpdatedAt: now,
		}
	}
}

func (s *Store) ListWorkspaces() []string {
	return s.ListWorkspacesForAccount("")
}

func (s *Store) ListWorkspacesForAccount(accountID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if accountID != "" {
		if ws, ok := s.accountWorkspaces[accountID]; ok && len(ws) > 0 {
			out := make([]string, len(ws))
			for i, w := range ws {
				out[i] = w.Name
			}
			return out
		}
	}
	out := make([]string, len(s.workspaces))
	for i, w := range s.workspaces {
		out[i] = w.Name
	}
	return out
}

func (s *Store) WorkspacePath(name string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, w := range s.workspaces {
		if w.Name == name {
			return w.Path, true
		}
	}
	return "", false
}

func (s *Store) seedDemo() {
	if os.Getenv("MOBAGENT_SEED_DEMO") == "0" {
		return
	}
	now := time.Now().UTC()
	taskID := "task-demo-1"
	ag, ok := s.agents["cursor-1"]
	if !ok {
		return
	}

	ag.Status = "running"
	ag.Workspace = firstNonEmpty(ag.Workspace, "mobagent")
	ag.CurrentStage = "Reading project structure..."
	ag.CurrentFile = "src/auth/login.ts"
	ag.Progress = 67
	ag.ElapsedSec = 738
	ag.Tokens = 2400
	ag.CostUSD = 0.023
	ag.Branch = "feature/user-login"
	ag.CPU = 42
	ag.MemoryMB = 512
	ag.CurrentTask = "Implement user login function"
	ag.TaskID = taskID
	ag.UpdatedAt = now

	s.tasks[taskID] = &models.Task{
		ID: taskID, AccountID: "demo", SessionID: taskID, AgentID: ag.ID, AgentType: ag.Type,
		Title: "Implement user login function",
		Prompt: "Implement user login, support email/password, use JWT...",
		Workspace: ag.Workspace, Branch: ag.Branch, Priority: "medium",
		Status: "running", Progress: 67, Stage: "Editing",
		CurrentFile: "src/auth/login.ts", Tokens: 2400, CostUSD: 0.023, ElapsedSec: 738,
		CreatedAt: now.Add(-12 * time.Minute), UpdatedAt: now,
	}

	s.console[taskID] = []string{
		"[12:01:02] Analyzing project structure...",
		"[12:01:05] Detected TypeScript project",
		"[12:01:08] Reading src/auth/login.ts",
		"[12:01:12] Searching for auth middleware",
		"[12:01:18] Planning implementation",
		"[12:01:25] Editing src/auth/login.ts (+32 -4)",
		"[12:01:30] Editing src/auth/jwt.ts (+18 -0)",
		"[12:01:35] Creating src/types/user.d.ts (+6 -0)",
		"[12:02:10] Running npm run test",
		"[12:02:45] PASS src/auth/login.test.ts",
		"[12:02:46] PASS src/auth/jwt.test.ts",
		"[12:02:47] PASS src/types/user.test.ts",
		"[12:02:50] Test Suites: 3 passed, 3 total",
		"[12:02:50] Tests: 128 passed, 128 total",
		"[12:02:52] Reviewing changes...",
		"[12:02:58] Task completed successfully 🎉",
	}

	s.diffs[taskID] = []models.FileDiff{
		{
			Path: "src/auth/login.ts", Additions: 32, Deletions: 4, Status: "modified",
			Diff: "--- a/src/auth/login.ts\n+++ b/src/auth/login.ts\n@@ -12,8 +12,24 @@\n-  return res.status(401).json({ error: 'Invalid credentials' });\n+  return res.status(401).json({\n+    error: 'Invalid credentials',\n+    code: 'AUTH_INVALID',\n+  });\n+export async function loginUser(email: string, password: string) {\n+  const user = await findUserByEmail(email);\n+  if (!user || !verifyPassword(password, user.hash)) {\n+    throw new AuthError('AUTH_USER_NOT_FOUND');\n+  }\n+  return signJwt({ sub: user.id, email: user.email });\n+}\n",
		},
		{
			Path: "src/auth/jwt.ts", Additions: 18, Deletions: 0, Status: "modified",
			Diff: "--- a/src/auth/jwt.ts\n+++ b/src/auth/jwt.ts\n@@ -0,0 +1,18 @@\n+import jwt from 'jsonwebtoken';\n+export function signJwt(payload: object) {\n+  return jwt.sign(payload, process.env.JWT_SECRET!, { expiresIn: '7d' });\n+}\n",
		},
		{
			Path: "src/types/user.d.ts", Additions: 6, Deletions: 0, Status: "added",
			Diff: "--- /dev/null\n+++ b/src/types/user.d.ts\n@@ -0,0 +1,6 @@\n+export interface User {\n+  id: string;\n+  email: string;\n+  hash: string;\n+}\n",
		},
	}

	ws := ag.Workspace
	s.taskEvts[taskID] = []models.AgentEvent{
		{ID: "ev1", TaskID: taskID, SessionID: taskID, Agent: "cursor", Workspace: ws, Type: models.EventTaskStarted, Status: models.StatusCompleted, Timestamp: now.Add(-11 * time.Minute), Payload: map[string]interface{}{"title": "Implement user login function"}},
		{ID: "ev2", TaskID: taskID, SessionID: taskID, Agent: "cursor", Workspace: ws, Type: models.EventReadingProject, Status: models.StatusCompleted, Timestamp: now.Add(-10 * time.Minute), Payload: map[string]interface{}{"message": "Read 48 files"}},
		{ID: "ev3", TaskID: taskID, SessionID: taskID, Agent: "cursor", Workspace: ws, Type: models.EventSearching, Status: models.StatusCompleted, Timestamp: now.Add(-9 * time.Minute), Payload: map[string]interface{}{"query": "login related modules"}},
		{ID: "ev4", TaskID: taskID, SessionID: taskID, Agent: "cursor", Workspace: ws, Type: models.EventPlanning, Status: models.StatusCompleted, Timestamp: now.Add(-8 * time.Minute), Payload: map[string]interface{}{"message": "Generating implementation plan"}},
		{ID: "ev5", TaskID: taskID, SessionID: taskID, Agent: "cursor", Workspace: ws, Type: models.EventEditing, Status: models.StatusActive, Timestamp: now.Add(-5 * time.Minute), Payload: map[string]interface{}{
			"files": []map[string]interface{}{
				{"path": "src/auth/login.ts", "additions": 32, "deletions": 4},
				{"path": "src/auth/jwt.ts", "additions": 18, "deletions": 0},
				{"path": "src/types/user.d.ts", "additions": 6, "deletions": 0},
			},
		}},
		{ID: "ev6", TaskID: taskID, SessionID: taskID, Agent: "cursor", Workspace: ws, Type: models.EventRunningTests, Status: models.StatusActive, Timestamp: now.Add(-2 * time.Minute), Payload: map[string]interface{}{"message": "Executing unit tests"}},
		{ID: "ev7", TaskID: taskID, SessionID: taskID, Agent: "cursor", Workspace: ws, Type: models.EventReviewing, Status: models.StatusPending, Timestamp: now, Payload: map[string]interface{}{"message": "Code review in progress"}},
		{ID: "ev8", TaskID: taskID, SessionID: taskID, Agent: "cursor", Workspace: ws, Type: models.EventFinished, Status: models.StatusPending, Timestamp: now, Payload: map[string]interface{}{"message": "Task finished"}},
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func (s *Store) SeedEventsInto(engine *events.Engine) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for taskID, evts := range s.taskEvts {
		for _, ev := range evts {
			e := ev
			e.TaskID = taskID
			engine.Publish(e)
		}
	}
}

func (s *Store) ListAgents() []models.Agent {
	return s.ListAgentsForAccount("")
}

func (s *Store) ListAgentsForAccount(accountID string) []models.Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Agent, 0, len(s.agents))
	order := make([]string, 0, len(s.agents))
	for id, a := range s.agents {
		if accountID != "" && a.AccountID != accountID {
			continue
		}
		order = append(order, id)
	}
	sort.Strings(order)
	for _, id := range order {
		if a, ok := s.agents[id]; ok {
			if accountID != "" && a.AccountID != accountID {
				continue
			}
			out = append(out, *a)
		}
	}
	return out
}

func (s *Store) GetAgentForAccount(accountID, id string) (*models.Agent, bool) {
	a, ok := s.GetAgent(id)
	if !ok {
		return nil, false
	}
	if accountID != "" && a.AccountID != accountID {
		return nil, false
	}
	return a, ok
}

func (s *Store) GetAgent(id string) (*models.Agent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.agents[id]
	if !ok {
		return nil, false
	}
	cp := *a
	return &cp, true
}

func (s *Store) UpdateAgent(a *models.Agent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a.UpdatedAt = time.Now().UTC()
	s.agents[a.ID] = a
}

func (s *Store) ListTasks() []models.Task {
	return s.ListTasksForAccount("")
}

func (s *Store) ListTasksForAccount(accountID string) []models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		if accountID != "" && t.AccountID != accountID {
			continue
		}
		out = append(out, *t)
	}
	return out
}

func (s *Store) GetTaskForAccount(accountID, id string) (*models.Task, bool) {
	t, ok := s.GetTask(id)
	if !ok {
		return nil, false
	}
	if accountID != "" && t.AccountID != accountID {
		return nil, false
	}
	return t, ok
}

func (s *Store) SyncWorkerCatalog(accountID, workerID string, agents []models.WorkerAgentDecl, workspaces []models.WorkerWorkspaceDecl) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	for id, a := range s.agents {
		if a.AccountID == accountID && a.WorkerID == workerID {
			delete(s.agents, id)
		}
	}
	wsList := make([]config.Workspace, 0, len(workspaces))
	for _, w := range workspaces {
		wsList = append(wsList, config.Workspace{Name: w.Name, Path: w.Path})
	}
	if len(wsList) > 0 {
		s.accountWorkspaces[accountID] = wsList
	}
	for _, def := range agents {
		ws := def.Workspace
		if ws == "" && len(wsList) > 0 {
			ws = wsList[0].Name
		}
		s.agents[def.ID] = &models.Agent{
			ID: def.ID, AccountID: accountID, WorkerID: workerID, Online: true,
			Name: def.Name, Type: def.Type, Status: "idle",
			Workspace: ws, Branch: def.Branch,
			Progress: 0, CreatedAt: now, UpdatedAt: now,
		}
	}
}

func (s *Store) MarkWorkerOffline(accountID, workerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range s.agents {
		if a.AccountID == accountID && a.WorkerID == workerID {
			a.Online = false
			if a.Status == "running" {
				a.Status = "offline"
			} else if a.Status != "stopped" {
				a.Status = "offline"
			}
			a.UpdatedAt = time.Now().UTC()
		}
	}
}

func (s *Store) GetTask(id string) (*models.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	if !ok {
		return nil, false
	}
	cp := *t
	return &cp, true
}

func (s *Store) SaveTask(t *models.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	t.UpdatedAt = now
	s.tasks[t.ID] = t
}

func (s *Store) AppendConsole(taskID, line string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.console[taskID] = append(s.console[taskID], line)
	if len(s.console[taskID]) > 10000 {
		s.console[taskID] = s.console[taskID][len(s.console[taskID])-10000:]
	}
}

func (s *Store) ConsoleLines(taskID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, len(s.console[taskID]))
	copy(out, s.console[taskID])
	return out
}

func (s *Store) SetDiffs(taskID string, diffs []models.FileDiff) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.diffs[taskID] = diffs
}

func (s *Store) GetDiffs(taskID string) []models.FileDiff {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.FileDiff, len(s.diffs[taskID]))
	copy(out, s.diffs[taskID])
	return out
}

func (s *Store) GetDiff(taskID, path string) (*models.FileDiff, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.diffs[taskID] {
		if d.Path == path {
			cp := d
			return &cp, true
		}
	}
	return nil, false
}

func (s *Store) DemoTaskID() string { return "task-demo-1" }
