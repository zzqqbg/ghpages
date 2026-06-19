package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/ghpages/mobagent/backend/internal/adapter"
	"github.com/ghpages/mobagent/backend/internal/config"
	"github.com/ghpages/mobagent/backend/internal/log"
	"github.com/ghpages/mobagent/backend/internal/workerclient"
	"go.uber.org/zap"
)

func main() {
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = filepath.Join("..", "data", "logs")
	}
	_ = log.Init(log.Config{Dir: logDir, Level: os.Getenv("LOG_LEVEL")})
	defer log.Sync()

	token := os.Getenv("MOBAGENT_TOKEN")
	if token == "" {
		log.Logger().Fatal("MOBAGENT_TOKEN required")
	}
	cloudURL := os.Getenv("MOBAGENT_CLOUD_WS")
	if cloudURL == "" {
		cloudURL = "ws://127.0.0.1:8790/ws/worker"
	}

	reg := adapter.NewRegistry(adapter.DefaultSimulatedAdapters()...)
	cfg := config.Load(reg)
	client := workerclient.New(cloudURL, token, reg, cfg)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Logger().Info("worker starting",
		zap.String("cloud", cloudURL),
		zap.Int("agents", len(cfg.Agents)),
		zap.Int("workspaces", len(cfg.Workspaces)),
	)
	if err := client.Run(ctx); err != nil && err != context.Canceled {
		log.Logger().Fatal("worker exit", zap.Error(err))
	}
}
