package judge

import (
	"fmt"

	"github.com/ayussh-2/config"
	"github.com/ayussh-2/internal/database"
	"github.com/ayussh-2/internal/logger"
	"github.com/ayussh-2/internal/metrics"
	"github.com/ayussh-2/internal/queue"
	"github.com/ayussh-2/internal/sandbox/docker"
	"github.com/ayussh-2/internal/services"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Bootstrap initializes all dependencies for the judge worker.
func Bootstrap(cfg *config.Config) (*Dependencies, error) {
	log, err := logger.Init(cfg.Env)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	db, err := database.NewPostgres(cfg, log)
	if err != nil {
		log.Fatal("Cannot connect to DB", zap.Error(err))
	}

	queueClient, err := queue.New(cfg)
	if err != nil {
		log.Fatal("Cannot connect to NATS", zap.Error(err))
	}

	if err := queueClient.EnsureStream(cfg.JudgeStream, []string{cfg.JudgeSubject}); err != nil {
		log.Fatal("Cannot ensure judge stream", zap.Error(err))
	}

	dockerCli, err := docker.NewDefaultClient()
	if err != nil {
		log.Fatal("Cannot connect to docker", zap.Error(err))
	}

	return &Dependencies{
		Config:  cfg,
		Log:     log,
		DB:      db,
		Queue:   queueClient,
		Docker:  dockerCli,
		Metrics: metrics.NewJudgeMetrics(),
	}, nil
}

// Dependencies holds all initialized components for the judge worker.
type Dependencies struct {
	Config  *config.Config
	Log     *zap.Logger
	DB      *gorm.DB
	Queue   *queue.Client
	Docker  *client.Client
	Metrics *metrics.JudgeMetrics
}

// Service returns a configured submission service.
func (d *Dependencies) Service() *services.SubmissionService {
	runner := docker.NewRunner(d.Docker, d.Config.SandboxWorkDir)
	return services.NewSubmissionService(d.Log, d.DB, d.Config, runner, nil)
}

// Close cleans up all resources.
func (d *Dependencies) Close() error {
	if d.DB != nil {
		sqlDB, err := d.DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
	if d.Queue != nil {
		d.Queue.Close()
	}
	if d.Docker != nil {
		d.Docker.Close()
	}
	if d.Log != nil {
		d.Log.Sync()
	}
	return nil
}

// ValidateConfig checks if the configuration is valid for the judge worker.
func ValidateConfig(cfg *config.Config) error {
	if cfg.JudgeStream == "" {
		return fmt.Errorf("JudgeStream not configured")
	}
	if cfg.JudgeSubject == "" {
		return fmt.Errorf("JudgeSubject not configured")
	}
	if cfg.JudgeConcurrency <= 0 {
		return fmt.Errorf("JudgeConcurrency must be positive")
	}
	return nil
}
