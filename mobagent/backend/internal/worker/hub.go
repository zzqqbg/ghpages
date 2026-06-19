package worker

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/ghpages/mobagent/backend/internal/auth"
	"github.com/ghpages/mobagent/backend/internal/events"
	"github.com/ghpages/mobagent/backend/internal/log"
	"github.com/ghpages/mobagent/backend/internal/models"
	"github.com/ghpages/mobagent/backend/internal/store"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 90 * time.Second
	pingPeriod = (pongWait * 9) / 10
	sendBuf    = 4096
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Hub struct {
	mu       sync.RWMutex
	accounts map[string]map[string]*Conn // accountID -> workerID -> conn
	store    *store.Store
	engine   *events.Engine
	auth     *auth.Store
	onDone   func(taskID, status, errMsg string)
}

type Conn struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	accountID string
	workerID  string
	agents    map[string]struct{}
}

func NewHub(st *store.Store, eng *events.Engine, authStore *auth.Store) *Hub {
	return &Hub{
		accounts: make(map[string]map[string]*Conn),
		store:    st,
		engine:   eng,
		auth:     authStore,
	}
}

func (h *Hub) SetTaskDoneHandler(fn func(taskID, status, errMsg string)) {
	h.onDone = fn
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	token := auth.ExtractToken(
		r.Header.Get("Authorization"),
		r.URL.Query().Get("token"),
		r.Header.Get("X-MobAgent-Token"),
	)
	acct, ok := h.auth.ValidateToken(token)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &Conn{
		hub:       h,
		conn:      conn,
		send:      make(chan []byte, sendBuf),
		accountID: acct.ID,
		agents:    make(map[string]struct{}),
	}
	go c.writePump()
	go c.readPump()
	log.Logger().Info("worker ws connected", zap.String("account_id", acct.ID))
}

func (h *Hub) register(c *Conn) {
	h.mu.Lock()
	if h.accounts[c.accountID] == nil {
		h.accounts[c.accountID] = make(map[string]*Conn)
	}
	if old, ok := h.accounts[c.accountID][c.workerID]; ok && old != c {
		close(old.send)
		old.conn.Close()
	}
	h.accounts[c.accountID][c.workerID] = c
	h.mu.Unlock()
}

func (h *Hub) unregister(c *Conn) {
	h.mu.Lock()
	if workers, ok := h.accounts[c.accountID]; ok {
		if cur, ok := workers[c.workerID]; ok && cur == c {
			delete(workers, c.workerID)
		}
		if len(workers) == 0 {
			delete(h.accounts, c.accountID)
		}
	}
	h.mu.Unlock()
	h.store.MarkWorkerOffline(c.accountID, c.workerID)
	close(c.send)
	c.conn.Close()
	log.Logger().Info("worker ws disconnected",
		zap.String("account_id", c.accountID),
		zap.String("worker_id", c.workerID),
	)
}

func (h *Hub) HasWorkerForAgent(accountID, agentID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.accounts[accountID] {
		if _, ok := c.agents[agentID]; ok {
			return true
		}
	}
	return false
}

func (h *Hub) DispatchTask(accountID string, req models.WorkerRunTask) error {
	h.mu.RLock()
	var target *Conn
	for _, c := range h.accounts[accountID] {
		if _, ok := c.agents[req.AgentID]; ok {
			target = c
			break
		}
	}
	h.mu.RUnlock()
	if target == nil {
		return errNoWorker
	}
	return target.sendJSON(models.WSMessage{Type: "run_task", Payload: req})
}

func (h *Hub) CancelTask(accountID, taskID string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.accounts[accountID] {
		_ = c.sendJSON(models.WSMessage{Type: "cancel_task", Payload: models.WorkerCancelTask{TaskID: taskID}})
	}
}

func (c *Conn) readPump() {
	defer func() { c.hub.unregister(c) }()
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var msg models.WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}
		switch msg.Type {
		case "register":
			c.handleRegister(msg.Payload)
		case "event":
			c.handleEvent(msg.Payload)
		case "console":
			c.handleConsole(msg.Payload)
		case "task_done":
			c.handleTaskDone(msg.Payload)
		case "ping":
			c.sendJSON(models.WSMessage{Type: "pong", Payload: map[string]int64{"ts": time.Now().Unix()}})
		}
	}
}

func (c *Conn) handleRegister(payload interface{}) {
	raw, _ := json.Marshal(payload)
	var reg models.WorkerRegister
	if json.Unmarshal(raw, &reg) != nil {
		return
	}
	if reg.WorkerID == "" {
		reg.WorkerID = events.NewID()
	}
	c.workerID = reg.WorkerID
	c.agents = make(map[string]struct{}, len(reg.Agents))
	for _, a := range reg.Agents {
		c.agents[a.ID] = struct{}{}
	}
	c.hub.store.SyncWorkerCatalog(c.accountID, c.workerID, reg.Agents, reg.Workspaces)
	c.hub.register(c)
	c.sendJSON(models.WSMessage{
		Type: "auth_ok",
		Payload: models.WorkerAuthOK{
			AccountID: c.accountID,
			WorkerID:  c.workerID,
			ServerTS:  time.Now().Unix(),
		},
	})
	c.sendJSON(models.WSMessage{
		Type:    "register_ok",
		Payload: models.WorkerRegisterOK{ServerTS: time.Now().Unix()},
	})
	log.Logger().Info("worker registered",
		zap.String("account_id", c.accountID),
		zap.String("worker_id", c.workerID),
		zap.Int("agents", len(reg.Agents)),
		zap.Int("workspaces", len(reg.Workspaces)),
	)
}

func (c *Conn) handleEvent(payload interface{}) {
	raw, _ := json.Marshal(payload)
	var wrap models.WorkerTaskEvent
	if json.Unmarshal(raw, &wrap) != nil || wrap.Event.TaskID == "" {
		return
	}
	ev := wrap.Event
	if ev.SessionID == "" {
		ev.SessionID = ev.TaskID
	}
	c.hub.engine.Publish(ev)
	c.syncTaskFromEvent(&ev)
}

func (c *Conn) handleConsole(payload interface{}) {
	raw, _ := json.Marshal(payload)
	var line models.WorkerConsoleLine
	if json.Unmarshal(raw, &line) != nil || line.TaskID == "" {
		return
	}
	c.hub.store.AppendConsole(line.TaskID, line.Line)
	c.hub.engine.Publish(models.AgentEvent{
		ID: events.NewID(), TaskID: line.TaskID, SessionID: line.SessionID,
		Agent: "worker", Workspace: "",
		Type: models.EventConsoleOutput, Status: models.StatusActive,
		Timestamp: time.Now().UTC(),
		Payload:   map[string]interface{}{"line": line.Line},
	})
}

func (c *Conn) handleTaskDone(payload interface{}) {
	raw, _ := json.Marshal(payload)
	var done models.WorkerTaskDone
	if json.Unmarshal(raw, &done) != nil {
		return
	}
	if task, ok := c.hub.store.GetTask(done.TaskID); ok {
		task.Status = done.Status
		if done.Error != "" {
			task.Stage = done.Error
		} else {
			task.Stage = done.Status
		}
		if done.Status == "completed" {
			task.Progress = 100
		}
		c.hub.store.SaveTask(task)
		if ag, ok := c.hub.store.GetAgent(task.AgentID); ok {
			if done.Status == "completed" || done.Status == "failed" || done.Status == "stopped" {
				ag.Status = "idle"
				ag.Progress = task.Progress
				ag.TaskID = ""
				ag.CurrentTask = ""
			}
			c.hub.store.UpdateAgent(ag)
		}
	}
	if c.hub.onDone != nil {
		c.hub.onDone(done.TaskID, done.Status, done.Error)
	}
}

func (c *Conn) syncTaskFromEvent(ev *models.AgentEvent) {
	task, ok := c.hub.store.GetTask(ev.TaskID)
	if !ok {
		return
	}
	task.Stage = string(ev.Type)
	if path, ok := ev.Payload["path"].(string); ok {
		task.CurrentFile = path
	}
	if ev.Type == models.EventFinished {
		task.Status = "completed"
		task.Progress = 100
	} else if ev.Type == models.EventFailed {
		task.Status = "failed"
	} else if task.Status == "queued" || task.Status == "starting" {
		task.Status = "running"
	}
	c.hub.store.SaveTask(task)
	if ag, ok := c.hub.store.GetAgent(task.AgentID); ok && ag.TaskID == task.ID {
		ag.Status = "running"
		ag.Progress = task.Progress
		ag.CurrentStage = task.Stage
		ag.CurrentFile = task.CurrentFile
		c.hub.store.UpdateAgent(ag)
	}
}

func (c *Conn) sendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	select {
	case c.send <- data:
		return nil
	default:
		log.Logger().Warn("worker send buffer full", zap.String("worker_id", c.workerID))
		return errSendFull
	}
}

func (c *Conn) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
