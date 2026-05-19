package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ayussh-2/config"
	"github.com/ayussh-2/internal/models"
	"github.com/ayussh-2/internal/sandbox"
	"github.com/ayussh-2/internal/sandbox/docker"
	"github.com/ayussh-2/internal/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const previewMaxBytes = 1024

type SubmissionService struct {
	log       *zap.Logger
	db        *gorm.DB
	cfg       *config.Config
	runner    sandbox.Runner
	publisher JudgePublisher
}

type JudgePublisher interface {
	Publish(subject string, payload []byte) error
}

type JudgeMessage struct {
	SubmissionID uint `json:"submission_id"`
}

func NewSubmissionService(log *zap.Logger, db *gorm.DB, cfg *config.Config, runner sandbox.Runner, publisher JudgePublisher) *SubmissionService {
	return &SubmissionService{log: log, db: db, cfg: cfg, runner: runner, publisher: publisher}
}

type CreateSubmissionInput struct {
	UserID      uint
	ProblemSlug string
	Language    string
	SourceCode  string
	Kind        string
}

type SubmissionDetail struct {
	models.Submission
	ProblemSlug string `json:"problem_slug"`
}

func (s *SubmissionService) Create(input CreateSubmissionInput) (*models.Submission, error) {
	if input.UserID == 0 {
		return nil, utils.NewAppError(http.StatusUnauthorized, "missing user", nil)
	}
	if _, ok := docker.Languages[input.Language]; !ok {
		return nil, utils.NewAppError(http.StatusBadRequest, "unsupported language", nil)
	}
	if input.Kind != models.SubmissionKindSubmit && input.Kind != models.SubmissionKindRun {
		return nil, utils.NewAppError(http.StatusBadRequest, "invalid kind", nil)
	}

	source := strings.TrimRight(input.SourceCode, " \t\r\n") + "\n"
	if strings.TrimSpace(source) == "" {
		return nil, utils.NewAppError(http.StatusBadRequest, "source code cannot be empty", nil)
	}
	if len(source) > s.cfg.SubmissionSourceMaxBytes {
		return nil, utils.NewAppError(http.StatusBadRequest, "source code is too large", nil)
	}

	var problem models.Problems
	if err := s.db.Select("id").Where("slug = ?", input.ProblemSlug).First(&problem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAppError(http.StatusNotFound, "problem not found", err)
		}
		s.log.Error("submission: failed to look up problem", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot create submission", err)
	}

	var pending int64
	if err := s.db.Model(&models.Submission{}).
		Where("user_id = ? AND status IN ?", input.UserID, []string{models.SubmissionStatusQueued, models.SubmissionStatusRunning}).
		Count(&pending).Error; err != nil {
		s.log.Error("submission: failed to count pending submissions", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot create submission", err)
	}
	if int(pending) >= s.cfg.SubmissionMaxPending {
		return nil, utils.NewAppError(http.StatusTooManyRequests, "too many pending submissions", nil)
	}

	sub := models.Submission{
		UserID:     input.UserID,
		ProblemID:  problem.ID,
		Language:   input.Language,
		Kind:       input.Kind,
		SourceCode: source,
		Status:     models.SubmissionStatusQueued,
		Verdict:    models.VerdictPending,
	}
	if err := s.db.Create(&sub).Error; err != nil {
		s.log.Error("submission: failed to insert", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot create submission", err)
	}
	if s.publisher == nil {
		return &sub, nil
	}

	payload, err := json.Marshal(JudgeMessage{SubmissionID: sub.ID})
	if err != nil {
		s.markQueueFailure(&sub, err)
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot queue submission", err)
	}
	if err := s.publisher.Publish(s.cfg.JudgeSubject, payload); err != nil {
		s.markQueueFailure(&sub, err)
		return nil, utils.NewAppError(http.StatusServiceUnavailable, "cannot queue submission", err)
	}
	return &sub, nil
}

// Judge runs the full sandbox pipeline for one submission.
func (s *SubmissionService) Judge(submissionID uint) error {
	log := s.log.With(zap.Uint("submission_id", submissionID))

	var sub models.Submission
	if err := s.db.First(&sub, submissionID).Error; err != nil {
		log.Error("judge: failed to load submission", zap.Error(err))
		return err
	}

	if err := s.markRunning(&sub); err != nil {
		return err
	}
	if err := s.db.Where("submission_id = ?", submissionID).Delete(&models.SubmissionTestResult{}).Error; err != nil {
		log.Error("judge: failed to reset previous results", zap.Error(err))
		return err
	}

	var problem models.Problems
	if err := s.db.Select("id").First(&problem, sub.ProblemID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s.finalize(&sub, models.VerdictIE, "problem not found")
		}
		log.Error("judge: failed to load problem", zap.Error(err))
		return err
	}

	testCases, err := s.loadTestCases(sub.ProblemID, sub.Kind)
	if err != nil {
		return err
	}
	if len(testCases) == 0 {
		return s.finalize(&sub, models.VerdictIE, "no test cases")
	}

	// Total deadline keeps a runaway submission from holding the worker forever.
	// Roughly: every test case can use 2x its run budget plus a compile budget.
	totalBudget := time.Duration(s.cfg.SandboxCompileTimeoutMs+s.cfg.SandboxRunTimeoutMs*2*len(testCases)) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), totalBudget+30*time.Second)
	defer cancel()

	artifactID, compileOut, err := s.compile(ctx, &sub)
	if artifactID != "" {
		defer func() { _ = s.runner.Cleanup(artifactID) }()
	}
	if err != nil {
		if errors.Is(err, sandbox.ErrCompileFailed) ||
			errors.Is(err, sandbox.ErrCompileTimeout) ||
			errors.Is(err, sandbox.ErrUnknownLanguage) {
			return nil
		}
		if errors.Is(err, context.DeadlineExceeded) {
			log.Error("judge: hard guard exceeded during compile", zap.Error(err))
			return s.finalize(&sub, models.VerdictIE, "submission exceeded judge time limit")
		}
		return err
	}
	_ = compileOut

	limits := s.judgeLimits()
	maxRuntime := 0
	maxMemory := 0
	finalVerdict := models.VerdictAC

	for i := range testCases {
		tc := testCases[i]
		runCtx, runCancel := context.WithTimeout(ctx, time.Duration(s.cfg.SandboxRunTimeoutMs*2)*time.Millisecond)
		result, runErr := s.runner.Run(runCtx, sub.Language, artifactID, tc.Input, limits)
		runCancel()

		if runErr != nil {
			log.Error("judge: runner.Run failed", zap.Error(runErr), zap.Uint("test_case_id", tc.ID))
			if errors.Is(runErr, context.DeadlineExceeded) {
				return s.finalize(&sub, models.VerdictIE, "submission exceeded judge time limit")
			}
			return runErr
		}

		verdict := verdictFor(result, tc.Expected)
		if err := s.persistTestResult(sub.ID, tc.ID, verdict, result); err != nil {
			return err
		}

		if result.DurationMs > maxRuntime {
			maxRuntime = result.DurationMs
		}
		if result.MemoryKB > maxMemory {
			maxMemory = result.MemoryKB
		}

		if verdict != models.VerdictAC {
			id := tc.ID
			sub.FailedTestCaseID = &id
			sub.FailedInputPreview = preview(tc.Input)
			sub.FailedExpectedPreview = preview(tc.Expected)
			sub.FailedActualPreview = preview(result.Stdout)
			if verdict == models.VerdictRE {
				sub.StderrPreview = preview(result.Stderr)
			}
			finalVerdict = verdict
			break
		}
	}

	sub.RuntimeMs = maxRuntime
	sub.MemoryKB = maxMemory
	return s.finalize(&sub, finalVerdict, "")
}

func (s *SubmissionService) GetByID(id, requesterID uint, isAdmin bool) (*models.Submission, error) {
	var sub models.Submission
	if err := s.db.Preload("Results").First(&sub, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAppError(http.StatusNotFound, "submission not found", err)
		}
		s.log.Error("submission: failed to fetch", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot fetch submission", err)
	}

	if !isAdmin && sub.UserID != requesterID {
		return nil, utils.NewAppError(http.StatusForbidden, "forbidden", nil)
	}

	return &sub, nil
}

type ListSubmissionsOpts struct {
	ProblemSlug string
	Kind        string
	Limit       int
}

func (s *SubmissionService) ListForUser(userID uint, opts ListSubmissionsOpts) ([]models.Submission, error) {
	q := s.db.Where("user_id = ?", userID).Order("id DESC")

	if opts.ProblemSlug != "" {
		var problem models.Problems
		if err := s.db.Select("id").Where("slug = ?", opts.ProblemSlug).First(&problem).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return []models.Submission{}, nil
			}
			return nil, utils.NewAppError(http.StatusInternalServerError, "cannot list submissions", err)
		}
		q = q.Where("problem_id = ?", problem.ID)
	}

	if opts.Kind == models.SubmissionKindSubmit || opts.Kind == models.SubmissionKindRun {
		q = q.Where("kind = ?", opts.Kind)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	q = q.Limit(limit)

	var subs []models.Submission
	if err := q.Find(&subs).Error; err != nil {
		s.log.Error("submission: failed to list", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot list submissions", err)
	}
	return subs, nil
}

// --- helpers ---

func (s *SubmissionService) markRunning(sub *models.Submission) error {
	sub.Status = models.SubmissionStatusRunning
	if err := s.db.Model(sub).Updates(map[string]any{"status": models.SubmissionStatusRunning}).Error; err != nil {
		s.log.Error("judge: failed to mark running", zap.Error(err), zap.Uint("submission_id", sub.ID))
		return err
	}
	return nil
}

func (s *SubmissionService) loadTestCases(problemID uint, kind string) ([]models.TestCase, error) {
	q := s.db.Where("problem_id = ?", problemID).Order("id ASC")
	if kind == models.SubmissionKindRun {
		q = q.Where("kind = ?", models.TestCaseKindSample)
	}
	var cases []models.TestCase
	if err := q.Find(&cases).Error; err != nil {
		return nil, fmt.Errorf("load test cases: %w", err)
	}
	return cases, nil
}

func (s *SubmissionService) compile(ctx context.Context, sub *models.Submission) (string, string, error) {
	compileCtx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.SandboxCompileTimeoutMs+5000)*time.Millisecond)
	defer cancel()

	artifactID, output, err := s.runner.Compile(compileCtx, sub.Language, sub.SourceCode)
	switch {
	case errors.Is(err, sandbox.ErrCompileFailed):
		sub.CompilerOutput = output
		if finalizeErr := s.finalize(sub, models.VerdictCE, ""); finalizeErr != nil {
			return "", output, finalizeErr
		}
		return artifactID, output, err
	case errors.Is(err, sandbox.ErrCompileTimeout):
		sub.CompilerOutput = output
		if finalizeErr := s.finalize(sub, models.VerdictIE, "compile timed out"); finalizeErr != nil {
			return "", output, finalizeErr
		}
		return "", output, err
	case errors.Is(err, sandbox.ErrUnknownLanguage):
		if finalizeErr := s.finalize(sub, models.VerdictIE, "unknown language"); finalizeErr != nil {
			return "", output, finalizeErr
		}
		return "", output, err
	case err != nil:
		s.log.Error("judge: compile failed", zap.Error(err), zap.Uint("submission_id", sub.ID))
		return "", output, err
	}
	return artifactID, output, nil
}

func (s *SubmissionService) judgeLimits() sandbox.Limits {
	return sandbox.Limits{
		RunTimeoutMs:     s.cfg.SandboxRunTimeoutMs,
		CompileTimeoutMs: s.cfg.SandboxCompileTimeoutMs,
		MemoryMB:         s.cfg.SandboxMemoryMB,
		CPUs:             s.cfg.SandboxCPUs,
		PidsLimit:        s.cfg.SandboxPidsLimit,
		StdoutMaxBytes:   s.cfg.SandboxStdoutMaxBytes,
		StderrMaxBytes:   s.cfg.SandboxStderrMaxBytes,
	}
}

func (s *SubmissionService) persistTestResult(submissionID, testCaseID uint, verdict string, result sandbox.RunResult) error {
	row := models.SubmissionTestResult{
		SubmissionID:        submissionID,
		TestCaseID:          testCaseID,
		Verdict:             verdict,
		RuntimeMs:           result.DurationMs,
		MemoryKB:            result.MemoryKB,
		ActualOutputPreview: preview(result.Stdout),
	}
	if err := s.db.Create(&row).Error; err != nil {
		s.log.Error("judge: failed to persist test result", zap.Error(err),
			zap.Uint("submission_id", submissionID), zap.Uint("test_case_id", testCaseID))
		return err
	}
	return nil
}

func (s *SubmissionService) finalize(sub *models.Submission, verdict, internalErr string) error {
	now := time.Now()
	sub.Status = models.SubmissionStatusDone
	sub.Verdict = verdict
	sub.JudgedAt = &now
	if internalErr != "" {
		sub.Error = internalErr
	}
	if err := s.db.Save(sub).Error; err != nil {
		s.log.Error("judge: failed to finalize", zap.Error(err), zap.Uint("submission_id", sub.ID))
		return err
	}
	s.afterSubmitFinalize(sub, verdict)
	return nil
}

func (s *SubmissionService) afterSubmitFinalize(sub *models.Submission, verdict string) {
	if sub.Kind != models.SubmissionKindSubmit {
		return
	}

	updates := map[string]any{"submit_count": gorm.Expr("submit_count + 1")}
	if verdict == models.VerdictAC {
		updates["ac_count"] = gorm.Expr("ac_count + 1")
	}
	if err := s.db.Model(&models.Problems{}).Where("id = ?", sub.ProblemID).Updates(updates).Error; err != nil {
		s.log.Error("judge: failed to update problem stats", zap.Error(err), zap.Uint("problem_id", sub.ProblemID))
	}

	if verdict == models.VerdictAC {
		userSvc := NewUserService(s.log, s.db, s.cfg)
		if err := userSvc.RecordRatingOnFirstAC(sub.UserID, sub.ProblemID); err != nil {
			s.log.Error("judge: failed to record rating", zap.Error(err), zap.Uint("user_id", sub.UserID))
		}
	}
}

func (s *SubmissionService) MarkInternalError(submissionID uint, message string) error {
	now := time.Now()
	return s.db.Model(&models.Submission{}).Where("id = ?", submissionID).Updates(map[string]any{
		"status":    models.SubmissionStatusDone,
		"verdict":   models.VerdictIE,
		"error":     message,
		"judged_at": now,
	}).Error
}

func (s *SubmissionService) Verdict(submissionID uint) (string, error) {
	var sub models.Submission
	if err := s.db.Select("verdict").First(&sub, submissionID).Error; err != nil {
		return "", err
	}
	return sub.Verdict, nil
}

func (s *SubmissionService) markQueueFailure(sub *models.Submission, err error) {
	if updateErr := s.MarkInternalError(sub.ID, err.Error()); updateErr != nil {
		s.log.Error("submission: failed to mark queue failure", zap.Error(updateErr), zap.Uint("submission_id", sub.ID))
	}
}

func verdictFor(r sandbox.RunResult, expected string) string {
	switch {
	case r.TimedOut:
		return models.VerdictTLE
	case r.OOM:
		return models.VerdictMLE
	case r.ExitCode != 0:
		return models.VerdictRE
	case !sandbox.Equal(expected, r.Stdout):
		return models.VerdictWA
	default:
		return models.VerdictAC
	}
}

func preview(s string) string {
	if len(s) <= previewMaxBytes {
		return s
	}
	return s[:previewMaxBytes] + "\n... [truncated]"
}
