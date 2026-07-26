package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ayussh-2/config"
	"github.com/ayussh-2/internal/database"
	"github.com/ayussh-2/internal/logger"
	"github.com/ayussh-2/internal/models"
	"github.com/ayussh-2/internal/sandbox"
	"github.com/ayussh-2/internal/sandbox/docker"
	"github.com/ayussh-2/internal/utils"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type APIResponse struct {
	Error   bool            `json:"error"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type CreateSubmissionResponse struct {
	SubmissionID uint `json:"submission_id"`
}

type GetSubmissionResponse struct {
	ID      uint   `json:"id"`
	Status  string `json:"status"`
	Verdict string `json:"verdict"`
}

func main() {
	cfg := config.LoadConfig()

	log, err := logger.Init(cfg.Env)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer log.Sync()

	log.Info("Loading database connection...")
	db, err := database.NewPostgres(cfg, log)
	if err != nil {
		log.Fatal("cannot connect to db", zap.Error(err))
	}

	// 1. Seed or retrieve verified benchmark user
	email := "benchmark@example.com"
	var user models.User
	err = db.Where("email = ?", email).First(&user).Error
	if err != nil {
		log.Info("Benchmark user not found, creating one...")
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("failed to hash password", zap.Error(err))
		}
		user = models.User{
			Name:          "Benchmark User",
			Email:         email,
			Password:      string(hashedPassword),
			Username:      "benchmark_user",
			Role:          "user",
			EmailVerified: true,
		}
		if err := db.Create(&user).Error; err != nil {
			log.Fatal("failed to create benchmark user", zap.Error(err))
		}
	} else {
		// Ensure user is verified
		if !user.EmailVerified {
			user.EmailVerified = true
			if err := db.Save(&user).Error; err != nil {
				log.Fatal("failed to update benchmark user verification status", zap.Error(err))
			}
		}
	}

	// 2. Generate access token
	token, err := utils.GenerateAccessToken(cfg, user.ID, user.Email, user.Role)
	if err != nil {
		log.Fatal("failed to generate access token", zap.Error(err))
	}
	log.Info("Successfully generated JWT token", zap.String("email", email))

	// 3. Ensure the problem "echo-input" exists in the database
	var problem models.Problems
	err = db.Where("slug = ?", "echo-input").First(&problem).Error
	if err != nil {
		log.Fatal("problem 'echo-input' does not exist in the database; please run 'make seed' first", zap.Error(err))
	}

	apiURL := fmt.Sprintf("http://localhost:%s/api/submissions", cfg.Port)
	client := &http.Client{Timeout: 10 * time.Second}

	// ----------------------------------------------------
	// Measurement 1 — Async end-to-end latency
	// ----------------------------------------------------
	log.Info("Starting Measurement 1 (Async end-to-end latency) - 20 runs...")
	var m1Durations []time.Duration
	m1Min := time.Duration(1<<63 - 1)
	var m1Max time.Duration
	var m1Total time.Duration

	for i := 0; i < 20; i++ {
		if i > 0 {
			time.Sleep(3100 * time.Millisecond)
		}
		start := time.Now()

		// 1. Post submission
		reqBody, _ := json.Marshal(map[string]string{
			"problem_slug": "echo-input",
			"language":     "python",
			"source_code":  "print(\"Hello World\")",
			"kind":         "submit",
		})
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBody))
		if err != nil {
			log.Fatal("failed to create POST request", zap.Error(err))
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Origin", cfg.FrontendURL)

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal("POST request failed", zap.Error(err))
		}
		if resp.StatusCode != http.StatusCreated {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			log.Fatal("POST request returned non-201 status", zap.Int("status", resp.StatusCode), zap.String("body", string(bodyBytes)))
		}

		var apiResp APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			resp.Body.Close()
			log.Fatal("failed to decode POST response", zap.Error(err))
		}
		resp.Body.Close()

		var createData CreateSubmissionResponse
		if err := json.Unmarshal(apiResp.Data, &createData); err != nil {
			log.Fatal("failed to unmarshal POST data", zap.Error(err))
		}

		submissionID := createData.SubmissionID

		// 2. Poll status endpoint every 100ms
		statusURL := fmt.Sprintf("%s/%d", apiURL, submissionID)
		for {
			time.Sleep(100 * time.Millisecond)

			reqGet, err := http.NewRequest("GET", statusURL, nil)
			if err != nil {
				log.Fatal("failed to create GET request", zap.Error(err))
			}
			reqGet.Header.Set("Authorization", "Bearer "+token)

			respGet, err := client.Do(reqGet)
			if err != nil {
				log.Fatal("GET request failed", zap.Error(err))
			}
			if respGet.StatusCode != http.StatusOK {
				bodyBytes, _ := io.ReadAll(respGet.Body)
				respGet.Body.Close()
				log.Fatal("GET request returned non-200 status", zap.Int("status", respGet.StatusCode), zap.String("body", string(bodyBytes)))
			}

			var apiRespGet APIResponse
			if err := json.NewDecoder(respGet.Body).Decode(&apiRespGet); err != nil {
				respGet.Body.Close()
				log.Fatal("failed to decode GET response", zap.Error(err))
			}
			respGet.Body.Close()

			var getSub GetSubmissionResponse
			if err := json.Unmarshal(apiRespGet.Data, &getSub); err != nil {
				log.Fatal("failed to unmarshal GET data", zap.Error(err))
			}

			if getSub.Verdict != "PENDING" {
				break
			}
		}

		delta := time.Since(start)
		m1Durations = append(m1Durations, delta)
		if delta < m1Min {
			m1Min = delta
		}
		if delta > m1Max {
			m1Max = delta
		}
		m1Total += delta
		log.Info(fmt.Sprintf("Run %02d completed: %v (Verdict: %s)", i+1, delta, "Done"))
	}
	m1Avg := m1Total / 20

	// ----------------------------------------------------
	// Measurement 2 — API response time
	// ----------------------------------------------------
	log.Info("Starting Measurement 2 (API response time) - 20 runs...")
	var m2Total time.Duration

	for i := 0; i < 20; i++ {
		if i > 0 {
			time.Sleep(3100 * time.Millisecond)
		}
		start := time.Now()

		reqBody, _ := json.Marshal(map[string]string{
			"problem_slug": "echo-input",
			"language":     "python",
			"source_code":  "print(\"Hello World\")",
			"kind":         "submit",
		})
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBody))
		if err != nil {
			log.Fatal("failed to create POST request", zap.Error(err))
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Origin", cfg.FrontendURL)

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal("POST request failed", zap.Error(err))
		}
		resp.Body.Close()

		delta := time.Since(start)
		m2Total += delta
		log.Info(fmt.Sprintf("Run %02d completed: %v", i+1, delta))
	}
	m2Avg := m2Total / 20

	// ----------------------------------------------------
	// Measurement 3 — Sync baseline simulation
	// ----------------------------------------------------
	log.Info("Starting Measurement 3 (Sync baseline simulation) - 20 runs...")
	dockerCli, err := docker.NewDefaultClient()
	if err != nil {
		log.Fatal("failed to connect to docker", zap.Error(err))
	}
	defer dockerCli.Close()

	runner := docker.NewRunner(dockerCli, cfg.SandboxWorkDir)
	limits := sandbox.Limits{
		RunTimeoutMs:     cfg.SandboxRunTimeoutMs,
		CompileTimeoutMs: cfg.SandboxCompileTimeoutMs,
		MemoryMB:         cfg.SandboxMemoryMB,
		CPUs:             cfg.SandboxCPUs,
		PidsLimit:        cfg.SandboxPidsLimit,
		StdoutMaxBytes:   cfg.SandboxStdoutMaxBytes,
		StderrMaxBytes:   cfg.SandboxStderrMaxBytes,
	}

	var m3Durations []time.Duration
	m3Min := time.Duration(1<<63 - 1)
	var m3Max time.Duration
	var m3Total time.Duration

	ctx := context.Background()

	for i := 0; i < 20; i++ {
		start := time.Now()

		// Simulate compiling / preparing workspace for python
		artifactID, _, err := runner.Compile(ctx, "python", "print(\"Hello World\")")
		if err != nil {
			log.Fatal("direct Compile failed", zap.Error(err))
		}

		// Run execution inside docker sandbox
		_, err = runner.Run(ctx, "python", artifactID, "", limits)
		if err != nil {
			_ = runner.Cleanup(artifactID)
			log.Fatal("direct Run failed", zap.Error(err))
		}

		// Cleanup workspace
		_ = runner.Cleanup(artifactID)

		delta := time.Since(start)
		m3Durations = append(m3Durations, delta)
		if delta < m3Min {
			m3Min = delta
		}
		if delta > m3Max {
			m3Max = delta
		}
		m3Total += delta
		log.Info(fmt.Sprintf("Run %02d completed: %v", i+1, delta))
	}
	m3Avg := m3Total / 20

	// Calculate Improvement: reduction in API blocking time
	// API blocking time is simulated blocking sync time (m3Avg) vs actual async API response time (m2Avg)
	// Improvement = (Sync Time - Async Time) / Sync Time * 100
	improvement := float64(m3Avg.Milliseconds()-m2Avg.Milliseconds()) / float64(m3Avg.Milliseconds()) * 100.0

	// Print final formatted summary
	fmt.Println("\n====================================================")
	fmt.Println("LATENCY MEASUREMENT RESULTS")
	fmt.Println("====================================================")
	fmt.Printf("Async API response time (submit endpoint): %dms avg\n", m2Avg.Milliseconds())
	fmt.Printf("Sync sandbox execution time (direct Docker call): %dms avg\n", m3Avg.Milliseconds())
	fmt.Printf("Improvement: %.2f%% reduction in API blocking time\n", improvement)
	fmt.Printf("End-to-end verdict latency (async): %dms avg\n", m1Avg.Milliseconds())
	fmt.Println("====================================================")
}
