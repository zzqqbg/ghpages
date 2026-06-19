package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ghpages/mobagent/backend/internal/adapter"
	"github.com/ghpages/mobagent/backend/internal/models"
)

func TestDiscoverWorkspaces(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, name := range []string{"alpha", "beta", ".hidden"} {
		if name == ".hidden" {
			_ = os.Mkdir(filepath.Join(root, name), 0o755)
			continue
		}
		if err := os.Mkdir(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	got := discoverWorkspaces(root)
	if len(got) != 2 {
		t.Fatalf("want 2 workspaces, got %d", len(got))
	}
	if got[0].Name != "alpha" || got[1].Name != "beta" {
		t.Fatalf("unexpected order/names: %+v", got)
	}
}

func TestDefaultAgentsFromRegistry(t *testing.T) {
	t.Parallel()
	reg := adapter.NewRegistry(adapter.NewSimulated(models.AgentCursor, "Cursor Agent"))
	agents := defaultAgents(reg, []Workspace{{Name: "mobagent", Path: "/tmp/mobagent"}})
	if len(agents) != 1 {
		t.Fatalf("want 1 agent, got %d", len(agents))
	}
	if agents[0].ID != "cursor-1" || agents[0].Workspace != "mobagent" {
		t.Fatalf("unexpected agent: %+v", agents[0])
	}
}
