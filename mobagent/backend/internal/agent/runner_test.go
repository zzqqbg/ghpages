package agent_test

import (
	"testing"

	"github.com/ghpages/mobagent/backend/internal/adapter"
	"github.com/ghpages/mobagent/backend/internal/agent"
	"github.com/ghpages/mobagent/backend/internal/events"
	"github.com/ghpages/mobagent/backend/internal/models"
	"github.com/ghpages/mobagent/backend/internal/store"
)

func TestRunnerQueuesSecondTask(t *testing.T) {
	t.Parallel()
	st := store.NewTest()
	eng := events.NewEngine(nil)
	reg := adapter.NewRegistry(adapter.NewSimulated(models.AgentCursor, "Cursor"))
	r := agent.NewRunner(st, eng, reg, nil)

	req := models.CreateTaskRequest{
		AgentID: "cursor-1", Prompt: "task one", Workspace: "ws",
	}
	t1, err := r.CreateTask("demo", req)
	if err != nil {
		t.Fatal(err)
	}
	if t1.Status != "running" {
		t.Fatalf("first task should run, got %s", t1.Status)
	}

	t2, err := r.CreateTask("demo", models.CreateTaskRequest{
		AgentID: "cursor-1", Prompt: "task two", Workspace: "ws",
	})
	if err != nil {
		t.Fatal(err)
	}
	if t2.Status != "queued" {
		t.Fatalf("second task should queue, got %s", t2.Status)
	}
	if t2.QueuePosition != 1 {
		t.Fatalf("queue position want 1 got %d", t2.QueuePosition)
	}
	if r.QueueLength("cursor-1") != 1 {
		t.Fatalf("queue length want 1 got %d", r.QueueLength("cursor-1"))
	}
}
