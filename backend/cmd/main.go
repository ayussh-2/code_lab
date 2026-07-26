package main

import (
	"github.com/ayussh-2/config"
	"github.com/ayussh-2/internal/database"
	"github.com/ayussh-2/internal/logger"
	"github.com/ayussh-2/internal/middlewares"
	"github.com/ayussh-2/internal/queue"
	"github.com/ayussh-2/internal/routes"
	"github.com/ayussh-2/internal/sandbox/docker"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()
	if cfg.Env == "production" {
		gin.SetMode("release")
	}

	log, err := logger.Init(cfg.Env)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer log.Sync()

	log.Info("Starting server", zap.String("env", cfg.Env))

	db, err := database.NewPostgres(cfg, log)
	if err != nil {
		log.Fatal("Cannot connect to DB", zap.Error(err))
	}

	dockerCli, err := docker.NewDefaultClient()
	if err != nil {
		log.Fatal("Cannot connect to docker", zap.Error(err))
	}
	defer dockerCli.Close()

	runner := docker.NewRunner(dockerCli, cfg.SandboxWorkDir)

	queueClient, err := queue.New(cfg)
	if err != nil {
		log.Fatal("Cannot connect to NATS", zap.Error(err))
	}
	defer queueClient.Close()

	if err := queueClient.EnsureStream(cfg.JudgeStream, []string{cfg.JudgeSubject}); err != nil {
		log.Fatal("Cannot ensure judge stream", zap.Error(err))
	}

	server := gin.New()
	server.Use(gin.Logger(), gin.Recovery())
	server.Use(middlewares.CORSMiddleware(cfg.FrontendURL))
	server.Use(middlewares.CSRFOriginCheck(cfg.FrontendURL))

	routes.RegisterRoutes(server, log, db, cfg, runner, queueClient, dockerCli)

	if err := server.Run(); err != nil {
		log.Error("Failed to start server", zap.Error(err))
	}
}
