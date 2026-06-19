package events_test

import (
	"sync"
	"testing"

	"github.com/ghpages/mobagent/backend/internal/events"
	"github.com/ghpages/mobagent/backend/internal/models"
)

func TestEnginePublishDedupeID(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	var count int
	eng := events.NewEngine(func(string, models.AgentEvent) {
		mu.Lock()
		count++
		mu.Unlock()
	})
	ev := models.AgentEvent{ID: "same-id", TaskID: "t1", Type: models.EventPlanning}
	eng.Publish(ev)
	eng.Publish(ev)
	if count != 1 {
		t.Fatalf("expected 1 broadcast, got %d", count)
	}
}

func TestEngineEventsSince(t *testing.T) {
	t.Parallel()
	eng := events.NewEngine(nil)
	for i := 0; i < 5; i++ {
		eng.Publish(models.AgentEvent{
			ID: string(rune('a' + i)), TaskID: "t1", Type: models.EventConsoleOutput,
		})
	}
	got := eng.EventsSince("t1", "b")
	if len(got) != 3 {
		t.Fatalf("expected 3 events after b, got %d", len(got))
	}
}

func TestEngineIngestRaw(t *testing.T) {
	t.Parallel()
	eng := events.NewEngine(nil)
	ev := eng.IngestRaw("t1", "s1", "cursor", "/tmp", "Planning implementation")
	if ev == nil || ev.Type != models.EventPlanning {
		t.Fatalf("unexpected event: %+v", ev)
	}
}

func TestEngineStress100k(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	eng := events.NewEngine(nil)
	for i := 0; i < 100000; i++ {
		eng.Publish(models.AgentEvent{
			ID: events.NewID(), TaskID: "stress", Type: models.EventConsoleOutput,
			Payload: map[string]interface{}{"i": i},
		})
	}
	if eng.Count("stress") != 100000 {
		t.Fatalf("expected 100000 events, got %d", eng.Count("stress"))
	}
}
