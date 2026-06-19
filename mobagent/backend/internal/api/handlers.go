package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghpages/mobagent/backend/internal/agent"
	"github.com/ghpages/mobagent/backend/internal/events"
	"github.com/ghpages/mobagent/backend/internal/log"
	"github.com/ghpages/mobagent/backend/internal/models"
	"github.com/ghpages/mobagent/backend/internal/store"
	"github.com/ghpages/mobagent/backend/internal/ws"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	store  *store.Store
	engine *events.Engine
	runner *agent.Runner
	hub    *ws.Hub
}

func New(st *store.Store, eng *events.Engine, runner *agent.Runner, hub *ws.Hub) *Handler {
	return &Handler{store: st, engine: eng, runner: runner, hub: hub}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) ListAgents(c *gin.Context) {
	c.JSON(http.StatusOK, h.store.ListAgentsForAccount(accountID(c)))
}

func (h *Handler) GetAgent(c *gin.Context) {
	a, ok := h.store.GetAgentForAccount(accountID(c), c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	type agentView struct {
		models.Agent
		QueueLength int `json:"queueLength"`
	}
	c.JSON(http.StatusOK, agentView{Agent: *a, QueueLength: h.runner.QueueLength(a.ID)})
}

func (h *Handler) ListTasks(c *gin.Context) {
	c.JSON(http.StatusOK, h.store.ListTasksForAccount(accountID(c)))
}

func (h *Handler) CreateTask(c *gin.Context) {
	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Logger().Warn("create task bind failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task, err := h.runner.CreateTask(accountID(c), req)
	if err != nil {
		log.Logger().Warn("create task rejected", zap.Error(err), zap.String("agent_id", req.AgentID))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	status := http.StatusCreated
	if task.Status == "queued" {
		status = http.StatusAccepted
	}
	c.JSON(status, task)
}

func (h *Handler) GetTask(c *gin.Context) {
	t, ok := h.store.GetTask(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *Handler) TaskAction(c *gin.Context) {
	action := models.TaskAction(c.Param("action"))
	if err := h.runner.Action(c.Param("id"), action); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, _ := h.store.GetTask(c.Param("id"))
	c.JSON(http.StatusOK, t)
}

func (h *Handler) TaskEvents(c *gin.Context) {
	taskID := c.Param("id")
	evts := h.engine.AllEvents(taskID)
	if len(evts) == 0 {
		// fallback: demo seed may exist before engine sync
		if t, ok := h.store.GetTask(taskID); ok && t.ID == h.store.DemoTaskID() {
			evts = h.engine.AllEvents(taskID)
		}
	}
	c.JSON(http.StatusOK, evts)
}

func (h *Handler) ListWorkspaces(c *gin.Context) {
	c.JSON(http.StatusOK, h.store.ListWorkspacesForAccount(accountID(c)))
}

func (h *Handler) TaskFiles(c *gin.Context) {
	taskID := c.Param("id")
	diffs := h.store.GetDiffs(taskID)
	type fileItem struct {
		Path      string `json:"path"`
		Status    string `json:"status"`
		Additions int    `json:"additions"`
		Deletions int    `json:"deletions"`
	}
	files := make([]fileItem, 0, len(diffs)+4)
	for _, d := range diffs {
		files = append(files, fileItem{Path: d.Path, Status: d.Status, Additions: d.Additions, Deletions: d.Deletions})
	}
	// common project files for Files tab
	extra := []fileItem{
		{Path: "package.json", Status: "read"},
		{Path: "tsconfig.json", Status: "read"},
		{Path: "src/index.ts", Status: "read"},
		{Path: "src/auth/middleware.ts", Status: "read"},
	}
	seen := map[string]bool{}
	for _, f := range files {
		seen[f.Path] = true
	}
	for _, e := range extra {
		if !seen[e.Path] {
			files = append(files, e)
		}
	}
	c.JSON(http.StatusOK, files)
}

func (h *Handler) TaskConsole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"lines": h.store.ConsoleLines(c.Param("id"))})
}

func (h *Handler) TaskDiffs(c *gin.Context) {
	c.JSON(http.StatusOK, h.store.GetDiffs(c.Param("id")))
}

func (h *Handler) TaskDiffFile(c *gin.Context) {
	d, ok := h.store.GetDiff(c.Param("id"), c.Query("path"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *Handler) Metrics(c *gin.Context) {
	agents := h.store.ListAgents()
	var cpu, mem float64
	var tokens int
	var cost float64
	for _, a := range agents {
		cpu += a.CPU
		mem += a.MemoryMB
		tokens += a.Tokens
		cost += a.CostUSD
	}
	c.JSON(http.StatusOK, models.ResourceSnapshot{
		CPU: cpu, MemoryMB: mem, Tokens: tokens, CostUSD: cost,
	})
}

func (h *Handler) WS(c *gin.Context) {
	h.hub.HandleWS(c.Writer, c.Request)
}

func (h *Handler) StaticFrontend(c *gin.Context) {
	root := os.Getenv("STATIC_DIR")
	if root == "" {
		root = filepath.Join("..", "frontend", "dist")
	}
	path := c.Request.URL.Path
	if path == "" || path == "/" {
		c.File(filepath.Join(root, "index.html"))
		return
	}
	full := filepath.Join(root, filepath.Clean(strings.TrimPrefix(path, "/")))
	if info, err := os.Stat(full); err != nil || info.IsDir() {
		c.File(filepath.Join(root, "index.html"))
		return
	}
	c.File(full)
}
