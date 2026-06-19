package workerclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/ghpages/mobagent/backend/internal/adapter"
	"github.com/ghpages/mobagent/backend/internal/config"
	"github.com/ghpages/mobagent/backend/internal/events"
	"github.com/ghpages/mobagent/backend/internal/log"
	"github.com/ghpages/mobagent/backend/internal/models"
	"github.com/ghpages/mobagent/backend/internal/parser"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Client struct {
	cloudURL  string
	token     string
	workerID  string
	hostname  string
	reg       *adapter.Registry
	localCfg  *config.Config
	conn      *websocket.Conn
	send      chan []byte
	mu        sync.Mutex
	running   map[string]context.CancelFunc
}

func New(cloudURL, token string, reg *adapter.Registry, localCfg *config.Config) *Client {
	host, _ := os.Hostname()
	return &Client{
		cloudURL: cloudURL,
		token:    token,
		hostname: host,
		reg:      reg,
		localCfg: localCfg,
		send:     make(chan []byte, 4096),
		running:  make(map[string]context.CancelFunc),
	}
}

func (c *Client) Run(ctx context.Context) error {
	backoff := time.Second
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := c.connectOnce(ctx)
		if err != nil {
			log.Logger().Warn("worker disconnected", zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func (c *Client) connectOnce(ctx context.Context) error {
	u, err := url.Parse(c.cloudURL)
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("token", c.token)
	u.RawQuery = q.Encode()

	header := http.Header{}
	header.Set("Authorization", "Bearer "+c.token)

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), header)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	defer c.closeConn()

	if err := c.register(); err != nil {
		return err
	}

	readDone := make(chan struct{})
	go func() {
		c.readPump()
		close(readDone)
	}()
	go c.writePump()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-readDone:
		return fmt.Errorf("connection closed")
	}
}

func (c *Client) register() error {
	agents := make([]models.WorkerAgentDecl, 0, len(c.localCfg.Agents))
	for _, a := range c.localCfg.Agents {
		agents = append(agents, models.WorkerAgentDecl{
			ID: a.ID, Type: a.Type, Name: a.Name, Workspace: a.Workspace, Branch: a.Branch,
		})
	}
	workspaces := make([]models.WorkerWorkspaceDecl, 0, len(c.localCfg.Workspaces))
	for _, w := range c.localCfg.Workspaces {
		workspaces = append(workspaces, models.WorkerWorkspaceDecl{Name: w.Name, Path: w.Path})
	}
	payload := models.WorkerRegister{
		WorkerID: c.workerID, Hostname: c.hostname, Version: "1.0.0",
		Agents: agents, Workspaces: workspaces,
	}
	return c.sendJSON(models.WSMessage{Type: "register", Payload: payload})
}

func (c *Client) readPump() {
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		var msg models.WSMessage
		if json.Unmarshal(message, &msg) != nil {
			continue
		}
		switch msg.Type {
		case "auth_ok":
			raw, _ := json.Marshal(msg.Payload)
			var ack models.WorkerAuthOK
			if json.Unmarshal(raw, &ack) == nil && ack.WorkerID != "" {
				c.workerID = ack.WorkerID
			}
		case "run_task":
			raw, _ := json.Marshal(msg.Payload)
			var task models.WorkerRunTask
			if json.Unmarshal(raw, &task) == nil {
				go c.executeTask(task)
			}
		case "cancel_task":
			raw, _ := json.Marshal(msg.Payload)
			var cancel models.WorkerCancelTask
			if json.Unmarshal(raw, &cancel) == nil {
				c.cancelTask(cancel.TaskID)
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			c.mu.Lock()
			conn := c.conn
			c.mu.Unlock()
			if conn == nil {
				return
			}
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.sendJSON(models.WSMessage{Type: "ping"})
		}
	}
}

func (c *Client) executeTask(task models.WorkerRunTask) {
	sessionID := task.SessionID
	if sessionID == "" {
		sessionID = task.TaskID
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.mu.Lock()
	c.running[task.TaskID] = cancel
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		delete(c.running, task.TaskID)
		c.mu.Unlock()
		cancel()
	}()

	ad, ok := c.reg.Get(task.AgentType)
	if !ok {
		c.sendTaskDone(task.TaskID, sessionID, "failed", "unsupported agent type")
		return
	}

	log.Logger().Info("worker run task",
		zap.String("task_id", task.TaskID),
		zap.String("session_id", sessionID),
		zap.String("agent", ad.Name()),
	)

	p := parser.NewUniversalParser(string(task.AgentType), task.TaskID, sessionID, task.Workspace)
	emit := func(line string) {
		c.sendJSON(models.WSMessage{
			Type: "console",
			Payload: models.WorkerConsoleLine{
				TaskID: task.TaskID, SessionID: sessionID, Line: line,
			},
		})
		if ev, ok := p.Parse(parser.RawLine{Text: line, Timestamp: time.Now().UTC()}); ok {
			ev.ID = events.NewID()
			c.sendJSON(models.WSMessage{
				Type:    "event",
				Payload: models.WorkerTaskEvent{Event: *ev},
			})
		}
	}

	runTask := &models.Task{
		ID: task.TaskID, SessionID: sessionID, AgentID: task.AgentID, AgentType: task.AgentType,
		Prompt: task.Prompt, Workspace: task.Workspace, Branch: task.Branch, Title: task.Prompt,
	}

	err := ad.Run(ctx, runTask, emit)
	if err != nil {
		if err == context.Canceled {
			c.sendTaskDone(task.TaskID, sessionID, "stopped", "")
			return
		}
		c.sendTaskDone(task.TaskID, sessionID, "failed", err.Error())
		return
	}
	c.sendJSON(models.WSMessage{
		Type: "event",
		Payload: models.WorkerTaskEvent{Event: models.AgentEvent{
			ID: events.NewID(), TaskID: task.TaskID, SessionID: sessionID,
			Agent: string(task.AgentType), Workspace: task.Workspace,
			Type: models.EventFinished, Status: models.StatusCompleted,
			Timestamp: time.Now().UTC(),
			Payload:   map[string]interface{}{"message": "Task finished"},
		}},
	})
	c.sendTaskDone(task.TaskID, sessionID, "completed", "")
}

func (c *Client) cancelTask(taskID string) {
	c.mu.Lock()
	cancel, ok := c.running[taskID]
	c.mu.Unlock()
	if ok {
		cancel()
	}
}

func (c *Client) sendTaskDone(taskID, sessionID, status, errMsg string) {
	_ = c.sendJSON(models.WSMessage{
		Type: "task_done",
		Payload: models.WorkerTaskDone{
			TaskID: taskID, SessionID: sessionID, Status: status, Error: errMsg,
		},
	})
}

func (c *Client) sendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	select {
	case c.send <- data:
		return nil
	default:
		return fmt.Errorf("send buffer full")
	}
}

func (c *Client) closeConn() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}
