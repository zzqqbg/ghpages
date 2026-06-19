package main

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ghpages/mobagent/backend/internal/agent"
	"github.com/ghpages/mobagent/backend/internal/adapter"
	"github.com/ghpages/mobagent/backend/internal/api"
	"github.com/ghpages/mobagent/backend/internal/auth"
	"github.com/ghpages/mobagent/backend/internal/config"
	"github.com/ghpages/mobagent/backend/internal/events"
	"github.com/ghpages/mobagent/backend/internal/log"
	"github.com/ghpages/mobagent/backend/internal/models"
	"github.com/ghpages/mobagent/backend/internal/store"
	"github.com/ghpages/mobagent/backend/internal/worker"
	"github.com/ghpages/mobagent/backend/internal/ws"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = filepath.Join("..", "data", "logs")
	}
	if err := log.Init(log.Config{
		Dir:    logDir,
		Level:  os.Getenv("LOG_LEVEL"),
		Retain: 30 * 24 * time.Hour,
	}); err != nil {
		panic(err)
	}
	defer log.Sync()

	reg := adapter.NewRegistry(adapter.DefaultSimulatedAdapters()...)
	cfg := config.Load(reg)
	st := store.New(cfg)
	authStore := auth.Load(cfg.DataDir)

	var hub *ws.Hub
	eng := events.NewEngine(func(taskID string, ev models.AgentEvent) {
		if hub != nil {
			hub.BroadcastTask(taskID, ev)
		}
	})
	hub = ws.NewHub(eng)
	st.SeedEventsInto(eng)

	workerHub := worker.NewHub(st, eng, authStore)
	runner := agent.NewRunner(st, eng, reg, workerHub)
	workerHub.SetTaskDoneHandler(func(taskID, status, errMsg string) {
		task, ok := st.GetTask(taskID)
		if !ok {
			return
		}
		runner.OnWorkerTaskDone(task.AgentID)
	})
	h := api.New(st, eng, runner, hub)

	stop := make(chan struct{})
	hub.StartHeartbeat(stop)
	defer close(stop)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(log.Recover())
	r.Use(log.RequestLogger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-MobAgent-Token"},
		AllowCredentials: true,
	}))

	r.GET("/health", h.Health)
	r.GET("/ws", h.WS)
	r.GET("/ws/worker", func(c *gin.Context) { workerHub.HandleWS(c.Writer, c.Request) })

	apiGroup := r.Group("/api")
	apiGroup.Use(api.AccountMiddleware(authStore))
	{
		apiGroup.GET("/agents", h.ListAgents)
		apiGroup.GET("/agents/:id", h.GetAgent)
		apiGroup.GET("/tasks", h.ListTasks)
		apiGroup.POST("/tasks", h.CreateTask)
		apiGroup.GET("/tasks/:id", h.GetTask)
		apiGroup.POST("/tasks/:id/:action", h.TaskAction)
		apiGroup.GET("/tasks/:id/events", h.TaskEvents)
		apiGroup.GET("/tasks/:id/console", h.TaskConsole)
		apiGroup.GET("/tasks/:id/diffs", h.TaskDiffs)
		apiGroup.GET("/tasks/:id/diff", h.TaskDiffFile)
		apiGroup.GET("/metrics", h.Metrics)
		apiGroup.GET("/workspaces", h.ListWorkspaces)
		apiGroup.GET("/tasks/:id/files", h.TaskFiles)
	}

	staticRoot := os.Getenv("STATIC_DIR")
	if staticRoot == "" {
		staticRoot = filepath.Join("..", "frontend", "dist")
	}
	if _, err := os.Stat(staticRoot); err == nil {
		r.Static("/assets", filepath.Join(staticRoot, "assets"))
		r.GET("/", func(c *gin.Context) { c.File(filepath.Join(staticRoot, "index.html")) })
		r.NoRoute(func(c *gin.Context) {
			if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			h.StaticFrontend(c)
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8790"
	}
	addr := "0.0.0.0:" + port
	log.Logger().Info("server starting",
		zap.String("addr", addr),
		zap.String("log_dir", logDir),
		zap.String("config", cfg.ConfigPath),
		zap.String("workspaces_root", cfg.WorkspacesRoot),
		zap.Int("agents", len(cfg.Agents)),
		zap.Int("workspaces", len(cfg.Workspaces)),
	)
	if err := r.Run(addr); err != nil {
		log.Logger().Fatal("server exit", zap.Error(err))
	}
}
