package ws_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ghpages/mobagent/backend/internal/events"
	"github.com/ghpages/mobagent/backend/internal/models"
	"github.com/ghpages/mobagent/backend/internal/ws"
	"github.com/gorilla/websocket"
)

func TestWebSocketHelloAndEvent(t *testing.T) {
	t.Parallel()
	var hub *ws.Hub
	eng := events.NewEngine(func(taskID string, ev models.AgentEvent) {
		if hub != nil {
			hub.BroadcastTask(taskID, ev)
		}
	})
	hub = ws.NewHub(eng)

	srv := httptest.NewServer(http.HandlerFunc(hub.HandleWS))
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	hello, _ := json.Marshal(models.WSMessage{
		Type: "hello",
		Payload: models.WSClientHello{SessionID: "sess-1", TaskID: "task-1"},
	})
	if err := conn.WriteMessage(websocket.TextMessage, hello); err != nil {
		t.Fatal(err)
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(msg), "hello_ack") {
		t.Fatalf("expected hello_ack, got %s", msg)
	}

	eng.Publish(models.AgentEvent{
		ID: "ev1", TaskID: "task-1", SessionID: "sess-1",
		Type: models.EventPlanning, Status: models.StatusActive,
		Timestamp: time.Now().UTC(),
	})

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err = conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(msg), "Planning") {
		t.Fatalf("expected event broadcast, got %s", msg)
	}
}
