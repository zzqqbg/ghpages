package log

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var L *zap.Logger

type Config struct {
	Dir     string
	Level   string
	Retain  time.Duration
	Service string
}

func Init(cfg Config) error {
	if cfg.Dir == "" {
		cfg.Dir = filepath.Join("..", "data", "logs")
	}
	if cfg.Service == "" {
		cfg.Service = "mobagent-api"
	}

	level := zapcore.InfoLevel
	_ = level.UnmarshalText([]byte(cfg.Level))

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "ts"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encCfg.MessageKey = "msg"
	encCfg.LevelKey = "level"
	encCfg.CallerKey = "caller"

	jsonEnc := zapcore.NewJSONEncoder(encCfg)
	fileWS := zapcore.AddSync(NewWeeklyWriter(cfg.Dir, cfg.Service, cfg.Retain))
	fileCore := zapcore.NewCore(jsonEnc, fileWS, level)

	consoleEnc := zapcore.NewConsoleEncoder(encCfg)
	consoleCore := zapcore.NewCore(consoleEnc, zapcore.AddSync(os.Stdout), level)

	core := zapcore.NewTee(fileCore, consoleCore)
	L = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	zap.ReplaceGlobals(L)
	return nil
}

func Sync() {
	if L != nil {
		_ = L.Sync()
	}
}

func Sugar() *zap.SugaredLogger {
	if L == nil {
		return zap.NewNop().Sugar()
	}
	return L.Sugar()
}

func Logger() *zap.Logger {
	if L == nil {
		return zap.NewNop()
	}
	return L
}
