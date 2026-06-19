package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ghpages/mobagent/backend/internal/adapter"
	"github.com/ghpages/mobagent/backend/internal/models"
)

type Workspace struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type AgentDef struct {
	ID        string           `json:"id"`
	Type      models.AgentType `json:"type"`
	Name      string           `json:"name"`
	Workspace string           `json:"workspace"`
	Branch    string           `json:"branch"`
}

type fileConfig struct {
	WorkspacesRoot string     `json:"workspacesRoot"`
	Workspaces     []string   `json:"workspaces"`
	Agents         []AgentDef `json:"agents"`
}

type Config struct {
	DataDir        string
	ConfigPath     string
	WorkspacesRoot string
	Workspaces     []Workspace
	Agents         []AgentDef
}

func Load(reg *adapter.Registry) *Config {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = filepath.Join("..", "data")
	}

	cfg := &Config{
		DataDir:        dataDir,
		WorkspacesRoot: os.Getenv("WORKSPACES_ROOT"),
	}
	if cfg.WorkspacesRoot == "" {
		cfg.WorkspacesRoot = filepath.Join(dataDir, "workspaces")
	}

	cfg.ConfigPath = os.Getenv("MOBAGENT_CONFIG")
	if cfg.ConfigPath == "" {
		cfg.ConfigPath = filepath.Join(dataDir, "config.json")
	}

	var fc fileConfig
	if raw, err := os.ReadFile(cfg.ConfigPath); err == nil {
		_ = json.Unmarshal(raw, &fc)
	}
	if fc.WorkspacesRoot != "" {
		cfg.WorkspacesRoot = fc.WorkspacesRoot
	}

	cfg.Workspaces = discoverWorkspaces(cfg.WorkspacesRoot)
	if len(fc.Workspaces) > 0 {
		cfg.Workspaces = namedWorkspaces(fc.Workspaces, cfg.WorkspacesRoot)
	}
	if len(cfg.Workspaces) == 0 {
		if env := strings.TrimSpace(os.Getenv("WORKSPACES")); env != "" {
			cfg.Workspaces = namedWorkspaces(strings.Split(env, ","), cfg.WorkspacesRoot)
		}
	}

	cfg.Agents = fc.Agents
	if len(cfg.Agents) == 0 {
		cfg.Agents = defaultAgents(reg, cfg.Workspaces)
	}
	return cfg
}

func discoverWorkspaces(root string) []Workspace {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	out := make([]Workspace, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		out = append(out, Workspace{Name: e.Name(), Path: filepath.Join(root, e.Name())})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func namedWorkspaces(names []string, root string) []Workspace {
	out := make([]Workspace, 0, len(names))
	for _, raw := range names {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		path := filepath.Join(root, name)
		if st, err := os.Stat(path); err == nil && st.IsDir() {
			out = append(out, Workspace{Name: name, Path: path})
			continue
		}
		out = append(out, Workspace{Name: name, Path: path})
	}
	return out
}

func defaultAgents(reg *adapter.Registry, workspaces []Workspace) []AgentDef {
	defaultWS := ""
	if len(workspaces) > 0 {
		defaultWS = workspaces[0].Name
	}
	out := make([]AgentDef, 0)
	for _, ad := range reg.All() {
		t := ad.Type()
		out = append(out, AgentDef{
			ID:        string(t) + "-1",
			Type:      t,
			Name:      ad.Name(),
			Workspace: defaultWS,
			Branch:    "main",
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (c *Config) WorkspaceNames() []string {
	out := make([]string, len(c.Workspaces))
	for i, w := range c.Workspaces {
		out[i] = w.Name
	}
	return out
}

func (c *Config) ResolveWorkspace(name string) (Workspace, bool) {
	for _, w := range c.Workspaces {
		if w.Name == name {
			return w, true
		}
	}
	return Workspace{}, false
}
