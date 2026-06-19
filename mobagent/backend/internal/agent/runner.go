package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ghpages/mobagent/backend/internal/adapter"
	"github.com/ghpages/mobagent/backend/internal/events"
	"github.com/ghpages/mobagent/backend/internal/log"
	"github.com/ghpages/mobagent/backend/internal/models"
	"github.com/ghpages/mobagent/backend/internal/store"
	"github.com/ghpages/mobagent/backend/internal/worker"
	"go.uber.org/zap"
)

// Runner schedules tasks per agent. One agent runs one task at a time; extras queue (reuse adapter, minimal goroutines).
type Runner struct {
	store     *store.Store
	engine    *events.Engine
	registry  *adapter.Registry
	workerHub *worker.Hub
	mu        sync.Mutex
	slots     map[string]*agentSlot
	runs      map[string]*runState
}

type agentSlot struct {
	busy  bool
	queue []string
}

type runState struct {
	cancel context.CancelFunc
	paused bool
}

func NewRunner(st *store.Store, eng *events.Engine, reg *adapter.Registry, wh *worker.Hub) *Runner {
	return &Runner{
		store: st, engine: eng, registry: reg, workerHub: wh,
		slots: make(map[string]*agentSlot),
		runs:  make(map[string]*runState),
	}
}

func (r *Runner) usesWorker(accountID string, agentID string) bool {
	return r.workerHub != nil && r.workerHub.HasWorkerForAgent(accountID, agentID)
}

func (r *Runner) CreateTask(accountID string, req models.CreateTaskRequest) (*models.Task, error) {
	ag, ok := r.store.GetAgentForAccount(accountID, req.AgentID)
	if !ok {
		return nil, fmt.Errorf("agent not found")
	}
	if ag.WorkerID != "" && !r.usesWorker(accountID, req.AgentID) {
		return nil, fmt.Errorf("worker offline for agent %s", req.AgentID)
	}

	taskID := events.NewID()
	task := &models.Task{
		ID: taskID, AccountID: accountID, SessionID: taskID, AgentID: req.AgentID, AgentType: ag.Type,
		Title: truncate(req.Prompt, 80), Prompt: req.Prompt,
		Workspace: firstNonEmpty(req.Workspace, ag.Workspace),
		Branch:    firstNonEmpty(req.Branch, ag.Branch),
		Priority:  firstNonEmpty(req.Priority, "medium"),
		Status:    "running", Progress: 0, Stage: "Starting",
		WorkerID: ag.WorkerID,
	}
	if req.AgentType != "" {
		task.AgentType = req.AgentType
	}

	position := r.enqueue(req.AgentID, task)
	r.store.SaveTask(task)

	log.Logger().Info("task created",
		zap.String("task_id", task.ID),
		zap.String("agent_id", req.AgentID),
		zap.String("agent_type", string(task.AgentType)),
		zap.String("workspace", task.Workspace),
		zap.Int("queue_position", position),
	)

	r.engine.Publish(models.AgentEvent{
		ID: events.NewID(), TaskID: taskID, SessionID: taskID,
		Agent: string(ag.Type), Workspace: task.Workspace,
		Type: models.EventTaskStarted, Status: models.StatusActive,
		Timestamp: time.Now().UTC(),
		Payload: map[string]interface{}{
			"title": task.Title, "prompt": task.Prompt, "queuePosition": position,
		},
	})

	if position == 0 {
		r.markAgentRunning(ag, task)
		if r.usesWorker(accountID, req.AgentID) {
			if err := r.dispatchWorker(accountID, task); err != nil {
				task.Status = "failed"
				task.Stage = err.Error()
				r.store.SaveTask(task)
				r.releaseSlot(req.AgentID)
				return nil, err
			}
		} else {
			go r.start(task)
		}
	} else {
		task.Status = "queued"
		task.Stage = fmt.Sprintf("Queued #%d", position)
		task.QueuePosition = position
		r.store.SaveTask(task)
	}

	return task, nil
}

func (r *Runner) enqueue(agentID string, task *models.Task) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	slot := r.slot(agentID)
	if !slot.busy {
		slot.busy = true
		return 0
	}
	slot.queue = append(slot.queue, task.ID)
	return len(slot.queue)
}

func (r *Runner) slot(agentID string) *agentSlot {
	s, ok := r.slots[agentID]
	if !ok {
		s = &agentSlot{}
		r.slots[agentID] = s
	}
	return s
}

func (r *Runner) QueueLength(agentID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.slot(agentID).queue)
}

func (r *Runner) dispatchWorker(accountID string, task *models.Task) error {
	err := r.workerHub.DispatchTask(accountID, models.WorkerRunTask{
		TaskID: task.ID, SessionID: task.SessionID,
		AgentID: task.AgentID, AgentType: task.AgentType,
		Prompt: task.Prompt, Workspace: task.Workspace,
		Branch: task.Branch, Priority: task.Priority,
	})
	if err != nil {
		return err
	}
	task.Status = "running"
	task.Stage = "Dispatched"
	r.store.SaveTask(task)
	log.Logger().Info("task dispatched to worker",
		zap.String("task_id", task.ID),
		zap.String("account_id", accountID),
		zap.String("agent_id", task.AgentID),
	)
	return nil
}

func (r *Runner) releaseSlot(agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	slot := r.slot(agentID)
	slot.busy = false
}

func (r *Runner) markAgentRunning(ag *models.Agent, task *models.Task) {
	ag.Status = "running"
	ag.TaskID = task.ID
	ag.CurrentTask = task.Title
	ag.Progress = task.Progress
	ag.Workspace = task.Workspace
	r.store.UpdateAgent(ag)
}

func (r *Runner) start(task *models.Task) {
	ctx, cancel := context.WithCancel(context.Background())
	r.mu.Lock()
	r.runs[task.ID] = &runState{cancel: cancel}
	r.mu.Unlock()

	defer func() {
		r.mu.Lock()
		delete(r.runs, task.ID)
		r.mu.Unlock()
		r.scheduleNext(task.AgentID)
	}()

	ad, ok := r.registry.Get(task.AgentType)
	if !ok {
		r.fail(task, "unsupported agent type")
		return
	}

	log.Logger().Info("agent run start",
		zap.String("task_id", task.ID),
		zap.String("agent_id", task.AgentID),
		zap.String("adapter", ad.Name()),
	)

	progress := 0
	emit := func(line string) {
		r.store.AppendConsole(task.ID, line)
		ev := r.engine.IngestRaw(task.ID, task.ID, string(task.AgentType), task.Workspace, line)
		if ev != nil {
			r.updateTaskFromEvent(task, ev)
			progress = min(100, progress+7)
			task.Progress = progress
			r.store.SaveTask(task)
			r.syncAgentProgress(task)
		}
		r.engine.Publish(models.AgentEvent{
			ID: events.NewID(), TaskID: task.ID, SessionID: task.ID,
			Agent: string(task.AgentType), Workspace: task.Workspace,
			Type: models.EventConsoleOutput, Status: models.StatusActive,
			Timestamp: time.Now().UTC(),
			Payload:   map[string]interface{}{"line": line},
		})
	}

	err := ad.Run(ctx, task, emit)
	if err != nil {
		if err == context.Canceled {
			task.Status = "stopped"
			r.store.SaveTask(task)
			log.Logger().Info("agent run stopped", zap.String("task_id", task.ID))
		} else {
			r.fail(task, err.Error())
			log.Logger().Error("agent run failed", zap.String("task_id", task.ID), zap.Error(err))
		}
		return
	}

	task.Status = "completed"
	task.Progress = 100
	task.Stage = "Finished"
	task.QueuePosition = 0
	r.store.SaveTask(task)
	r.engine.Publish(models.AgentEvent{
		ID: events.NewID(), TaskID: task.ID, SessionID: task.ID,
		Agent: string(task.AgentType), Workspace: task.Workspace,
		Type: models.EventFinished, Status: models.StatusCompleted,
		Timestamp: time.Now().UTC(),
		Payload:   map[string]interface{}{"message": "Task finished"},
	})
	log.Logger().Info("agent run completed", zap.String("task_id", task.ID))
}

func (r *Runner) scheduleNext(agentID string) {
	r.mu.Lock()
	slot := r.slot(agentID)
	if len(slot.queue) == 0 {
		slot.busy = false
		r.mu.Unlock()
		r.updateAgentIdle(agentID)
		return
	}
	nextID := slot.queue[0]
	slot.queue = slot.queue[1:]
	r.mu.Unlock()

	task, ok := r.store.GetTask(nextID)
	if !ok {
		r.scheduleNext(agentID)
		return
	}
	task.Status = "running"
	task.Stage = "Starting"
	task.QueuePosition = 0
	r.store.SaveTask(task)
	if ag, ok := r.store.GetAgent(agentID); ok {
		r.markAgentRunning(ag, task)
	}
	log.Logger().Info("dequeue task", zap.String("task_id", task.ID), zap.String("agent_id", agentID))
	if r.usesWorker(task.AccountID, agentID) {
		if err := r.dispatchWorker(task.AccountID, task); err != nil {
			task.Status = "failed"
			task.Stage = err.Error()
			r.store.SaveTask(task)
			r.scheduleNext(agentID)
		}
		return
	}
	go r.start(task)
}

func (r *Runner) OnWorkerTaskDone(agentID string) {
	r.scheduleNext(agentID)
}

func (r *Runner) syncAgentProgress(task *models.Task) {
	ag, ok := r.store.GetAgent(task.AgentID)
	if !ok || ag.TaskID != task.ID {
		return
	}
	ag.Progress = task.Progress
	ag.CurrentStage = task.Stage
	ag.CurrentFile = task.CurrentFile
	r.store.UpdateAgent(ag)
}

func (r *Runner) updateTaskFromEvent(task *models.Task, ev *models.AgentEvent) {
	task.Stage = string(ev.Type)
	if path, ok := ev.Payload["path"].(string); ok {
		task.CurrentFile = path
	}
	if ev.Type == models.EventDiffUpdated || ev.Type == models.EventEditing {
		r.maybeUpdateDiffs(task.ID, ev)
	}
}

func (r *Runner) maybeUpdateDiffs(taskID string, ev *models.AgentEvent) {
	path, _ := ev.Payload["path"].(string)
	if path == "" {
		return
	}
	diff := models.FileDiff{
		Path: path, Additions: 32, Deletions: 4, Status: "modified",
		Diff: fmt.Sprintf("--- a/%s\n+++ b/%s\n+// updated by agent\n", path, path),
	}
	r.store.SetDiffs(taskID, append(r.store.GetDiffs(taskID), diff))
	r.engine.Publish(models.AgentEvent{
		ID: events.NewID(), TaskID: taskID, SessionID: taskID,
		Type: models.EventDiffUpdated, Status: models.StatusActive,
		Timestamp: time.Now().UTC(),
		Payload: map[string]interface{}{
			"files": []models.FileDiff{diff},
		},
	})
}

func (r *Runner) Action(taskID string, action models.TaskAction) error {
	task, ok := r.store.GetTask(taskID)
	if !ok {
		return fmt.Errorf("task not found")
	}
	log.Logger().Info("task action", zap.String("task_id", taskID), zap.String("action", string(action)))

	switch action {
	case models.ActionPause:
		r.mu.Lock()
		if rs, ok := r.runs[taskID]; ok {
			rs.paused = true
		}
		r.mu.Unlock()
		task.Status = "paused"
		r.store.SaveTask(task)
	case models.ActionResume, models.ActionContinue:
		if task.Status == "queued" {
			return fmt.Errorf("task is queued; wait until it starts")
		}
		task.Status = "running"
		r.store.SaveTask(task)
	case models.ActionStop:
		if task.Status == "queued" {
			r.removeFromQueue(task.AgentID, taskID)
			task.Status = "stopped"
			r.store.SaveTask(task)
			return nil
		}
		if r.usesWorker(task.AccountID, task.AgentID) && r.workerHub != nil {
			r.workerHub.CancelTask(task.AccountID, taskID)
		}
		r.mu.Lock()
		if rs, ok := r.runs[taskID]; ok {
			rs.cancel()
		}
		r.mu.Unlock()
		task.Status = "stopped"
		r.store.SaveTask(task)
	case models.ActionRestart, models.ActionRetry:
		_ = r.Action(taskID, models.ActionStop)
		time.Sleep(100 * time.Millisecond)
		_, err := r.CreateTask(task.AccountID, models.CreateTaskRequest{
			Prompt: task.Prompt, AgentID: task.AgentID, Workspace: task.Workspace,
			Branch: task.Branch, Priority: task.Priority, AgentType: task.AgentType,
		})
		return err
	case models.ActionExplain:
		r.engine.Publish(models.AgentEvent{
			ID: events.NewID(), TaskID: taskID, SessionID: taskID,
			Agent: string(task.AgentType), Workspace: task.Workspace,
			Type: models.EventPlanning, Status: models.StatusActive,
			Timestamp: time.Now().UTC(),
			Payload: map[string]interface{}{
				"message": fmt.Sprintf("Currently at %s (%d%%)", task.Stage, task.Progress),
			},
		})
	case models.ActionReview:
		r.engine.Publish(models.AgentEvent{
			ID: events.NewID(), TaskID: taskID, SessionID: taskID,
			Type: models.EventReviewing, Status: models.StatusActive,
			Timestamp: time.Now().UTC(),
			Payload:   map[string]interface{}{"message": "Review started"},
		})
	default:
		return fmt.Errorf("unknown action")
	}
	return nil
}

func (r *Runner) removeFromQueue(agentID, taskID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	slot := r.slot(agentID)
	out := slot.queue[:0]
	for _, id := range slot.queue {
		if id != taskID {
			out = append(out, id)
		}
	}
	slot.queue = out
	if !r.isRunning(agentID) && len(slot.queue) == 0 {
		slot.busy = false
		go r.updateAgentIdle(agentID)
	}
}

func (r *Runner) isRunning(agentID string) bool {
	for id, rs := range r.runs {
		t, ok := r.store.GetTask(id)
		if ok && t.AgentID == agentID && rs != nil {
			return true
		}
	}
	return false
}

func (r *Runner) fail(task *models.Task, msg string) {
	task.Status = "failed"
	task.Stage = "Failed"
	r.store.SaveTask(task)
	r.engine.Publish(models.AgentEvent{
		ID: events.NewID(), TaskID: task.ID, SessionID: task.ID,
		Agent: string(task.AgentType), Workspace: task.Workspace,
		Type: models.EventFailed, Status: models.StatusFailed,
		Timestamp: time.Now().UTC(),
		Payload:   map[string]interface{}{"error": msg},
	})
}

func (r *Runner) updateAgentIdle(agentID string) {
	ag, ok := r.store.GetAgent(agentID)
	if !ok {
		return
	}
	if r.QueueLength(agentID) > 0 {
		return
	}
	ag.Status = "idle"
	ag.Progress = 0
	ag.TaskID = ""
	ag.CurrentTask = ""
	r.store.UpdateAgent(ag)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
