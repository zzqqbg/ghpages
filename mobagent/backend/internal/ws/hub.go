package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ghpages/mobagent/backend/internal/events"
	"github.com/ghpages/mobagent/backend/internal/models"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	heartbeatEvery = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Hub struct {
	mu       sync.RWMutex
	clients  map[*Client]struct{}
	taskSubs map[string]map[*Client]struct{}
	engine   *events.Engine
}

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	sessionID string
	taskID    string
	lastEvent string
}

func NewHub(engine *events.Engine) *Hub {
	return &Hub{
		clients:  make(map[*Client]struct{}),
		taskSubs: make(map[string]map[*Client]struct{}),
		engine:   engine,
	}
}

func (h *Hub) BroadcastTask(taskID string, ev models.AgentEvent) {
	data, err := json.Marshal(models.WSMessage{Type: "event", Payload: ev})
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.taskSubs[taskID] {
		select {
		case c.send <- data:
		default:
		}
	}
	for c := range h.clients {
		if c.taskID == "" {
			select {
			case c.send <- data:
			default:
			}
		}
	}
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &Client{hub: h, conn: conn, send: make(chan []byte, 256)}
	h.register(c)
	go c.writePump()
	go c.readPump()
}

func (h *Hub) register(c *Client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) unregister(c *Client) {
	h.mu.Lock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send)
	}
	if c.taskID != "" {
		if subs, ok := h.taskSubs[c.taskID]; ok {
			delete(subs, c)
		}
	}
	h.mu.Unlock()
	c.conn.Close()
}

func (h *Hub) subscribeTask(c *Client, taskID string) {
	h.mu.Lock()
	if c.taskID != "" {
		if subs, ok := h.taskSubs[c.taskID]; ok {
			delete(subs, c)
		}
	}
	c.taskID = taskID
	if h.taskSubs[taskID] == nil {
		h.taskSubs[taskID] = make(map[*Client]struct{})
	}
	h.taskSubs[taskID][c] = struct{}{}
	h.mu.Unlock()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister(c)
	}()
	c.conn.SetReadLimit(512 * 1024)
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
		case "hello":
			c.handleHello(msg.Payload)
		case "subscribe":
			if m, ok := msg.Payload.(map[string]interface{}); ok {
				if tid, ok := m["taskId"].(string); ok {
					c.hub.subscribeTask(c, tid)
					c.sendResume(tid)
				}
			}
		case "ping":
			c.sendJSON(models.WSMessage{Type: "pong", Payload: map[string]int64{"ts": time.Now().Unix()}})
		}
	}
}

func (c *Client) handleHello(payload interface{}) {
	raw, _ := json.Marshal(payload)
	var hello models.WSClientHello
	_ = json.Unmarshal(raw, &hello)
	if hello.SessionID == "" {
		hello.SessionID = events.NewID()
	}
	c.sessionID = hello.SessionID
	c.lastEvent = hello.LastEventID
	if hello.TaskID != "" {
		c.hub.subscribeTask(c, hello.TaskID)
		c.sendResume(hello.TaskID)
	}
	c.sendJSON(models.WSMessage{
		Type: "hello_ack",
		Payload: models.WSHelloAck{SessionID: hello.SessionID, ServerTS: time.Now().Unix()},
	})
}

func (c *Client) sendResume(taskID string) {
	events := c.hub.engine.EventsSince(taskID, c.lastEvent)
	if len(events) == 0 {
		return
	}
	last := events[len(events)-1].ID
	c.sendJSON(models.WSMessage{
		Type: "resume",
		Payload: models.WSResumeAck{TaskID: taskID, Events: events, LastEvent: last},
	})
}

func (c *Client) sendJSON(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	select {
	case c.send <- data:
	default:
		log.Printf("ws: client send buffer full")
	}
}

func (c *Client) writePump() {
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

func (h *Hub) StartHeartbeat(stop <-chan struct{}) {
	ticker := time.NewTicker(heartbeatEvery)
	go func() {
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				data, _ := json.Marshal(models.WSMessage{
					Type: "heartbeat",
					Payload: map[string]int64{"ts": time.Now().Unix()},
				})
				h.mu.RLock()
				for c := range h.clients {
					select {
					case c.send <- data:
					default:
					}
				}
				h.mu.RUnlock()
			}
		}
	}()
}
