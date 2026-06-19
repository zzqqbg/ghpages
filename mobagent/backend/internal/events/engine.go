package events

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/ghpages/mobagent/backend/internal/models"
	"github.com/ghpages/mobagent/backend/internal/parser"
)

type BroadcastFunc func(taskID string, ev models.AgentEvent)

type Engine struct {
	mu        sync.RWMutex
	byTask    map[string][]models.AgentEvent
	ids       map[string]struct{}
	maxPerTask int
	onBroadcast BroadcastFunc
}

func NewEngine(onBroadcast BroadcastFunc) *Engine {
	return &Engine{
		byTask: make(map[string][]models.AgentEvent),
		ids:    make(map[string]struct{}),
		maxPerTask: 100000,
		onBroadcast: onBroadcast,
	}
}

func NewID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (e *Engine) IngestRaw(taskID, sessionID, agent, workspace, line string) *models.AgentEvent {
	p := parser.NewUniversalParser(agent, taskID, sessionID, workspace)
	ev, ok := p.Parse(parser.RawLine{Text: line, Timestamp: time.Now().UTC()})
	if !ok {
		return nil
	}
	return e.Publish(*ev)
}

func (e *Engine) Publish(ev models.AgentEvent) *models.AgentEvent {
	if ev.ID == "" {
		ev.ID = NewID()
	}
	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now().UTC()
	}

	e.mu.Lock()
	if _, dup := e.ids[ev.ID]; dup {
		e.mu.Unlock()
		return nil
	}
	e.ids[ev.ID] = struct{}{}
	list := e.byTask[ev.TaskID]
	if len(list) >= e.maxPerTask {
		list = list[1:]
	}
	list = append(list, ev)
	e.byTask[ev.TaskID] = list
	e.mu.Unlock()

	if e.onBroadcast != nil {
		e.onBroadcast(ev.TaskID, ev)
	}
	return &ev
}

func (e *Engine) MergeSimilar(in models.AgentEvent) *models.AgentEvent {
	e.mu.RLock()
	list := e.byTask[in.TaskID]
	e.mu.RUnlock()
	if len(list) == 0 {
		return e.Publish(in)
	}
	last := list[len(list)-1]
	if last.Type == in.Type && last.Status == in.Status {
		in.ID = last.ID
		in.Timestamp = time.Now().UTC()
		e.mu.Lock()
		list[len(list)-1] = in
		e.byTask[in.TaskID] = list
		e.mu.Unlock()
		if e.onBroadcast != nil {
			e.onBroadcast(in.TaskID, in)
		}
		return &in
	}
	return e.Publish(in)
}

func (e *Engine) EventsSince(taskID, lastEventID string) []models.AgentEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()
	list := e.byTask[taskID]
	if lastEventID == "" {
		out := make([]models.AgentEvent, len(list))
		copy(out, list)
		return out
	}
	start := 0
	for i, ev := range list {
		if ev.ID == lastEventID {
			start = i + 1
			break
		}
	}
	if start >= len(list) {
		return nil
	}
	out := make([]models.AgentEvent, len(list)-start)
	copy(out, list[start:])
	return out
}

func (e *Engine) AllEvents(taskID string) []models.AgentEvent {
	return e.EventsSince(taskID, "")
}

func (e *Engine) LastEventID(taskID string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	list := e.byTask[taskID]
	if len(list) == 0 {
		return ""
	}
	return list[len(list)-1].ID
}

func (e *Engine) Count(taskID string) int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.byTask[taskID])
}

func (e *Engine) Heartbeat(taskID, sessionID, agent, workspace string) {
	e.Publish(models.AgentEvent{
		ID: NewID(), TaskID: taskID, SessionID: sessionID, Agent: agent, Workspace: workspace,
		Type: models.EventHeartbeat, Status: models.StatusActive,
		Timestamp: time.Now().UTC(),
		Payload:   map[string]interface{}{"ts": time.Now().Unix()},
	})
}
