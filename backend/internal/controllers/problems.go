package controllers

import (
	"errors"
	"net/http"

	"github.com/ayussh-2/internal/services"
	"github.com/ayussh-2/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProblemController struct {
	log *zap.Logger
	svc *services.ProblemService
}

type CreateProblemRequest struct {
	Title           string                     `json:"title" binding:"required"`
	Difficulty      string                     `json:"difficulty" binding:"required,oneof=easy medium hard"`
	Topics          []uint                     `json:"topics" binding:"required,min=1"`
	Hint            []string                   `json:"hints" binding:"required,min=1"`
	Details         string                     `json:"details" binding:"required"`
	Examples        []services.Example         `json:"examples" binding:"required,min=1"`
	Constraints     []string                  `json:"constraints" binding:"required,min=1"`
	SampleTestCases []services.SampleTestCases `json:"sample_test_cases" binding:"required,min=1"`
}

type BulkCreateProblemRequest struct {
	Problems []CreateProblemRequest `json:"problems" binding:"required,min=1"`
}

type CreateTopicRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
}

type BulkCreateTopicRequest struct {
	Topics []CreateTopicRequest `json:"topics" binding:"required,min=1"`
}

func NewProblemController(log *zap.Logger, svc *services.ProblemService) *ProblemController {
	return &ProblemController{
		log: log,
		svc: svc,
	}
}

func (pc *ProblemController) CreateProblem(c *gin.Context) {
	var req CreateProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pc.log.Error("failed to parse create problem request", zap.Error(err))
		utils.ValidationFail(c, "Validation Failed", utils.ValidationErrors(err))
		return
	}

	resp, err := pc.svc.AddProblem(services.Problem{
		Title:           req.Title,
		Difficulty:      req.Difficulty,
		Topics:          req.Topics,
		Hint:            req.Hint,
		Details:         req.Details,
		Examples:        req.Examples,
		Constraints:     req.Constraints,
		SampleTestCases: req.SampleTestCases,
	})
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			utils.Fail(c, appErr.Status, appErr.Message)
			return
		}
		utils.Fail(c, http.StatusInternalServerError, "cannot create problem")
		return
	}

	utils.Success(c, http.StatusCreated, "problem created", resp)
}

func (pc *ProblemController) BulkCreateProblems(c *gin.Context) {
	var req BulkCreateProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pc.log.Error("failed to parse bulk create problem request", zap.Error(err))
		utils.ValidationFail(c, "Validation Failed", utils.ValidationErrors(err))
		return
	}

	items := make([]services.Problem, len(req.Problems))
	for i := range req.Problems {
		items[i] = services.Problem{
			Title:           req.Problems[i].Title,
			Difficulty:      req.Problems[i].Difficulty,
			Topics:          req.Problems[i].Topics,
			Hint:            req.Problems[i].Hint,
			Details:         req.Problems[i].Details,
			Examples:        req.Problems[i].Examples,
			Constraints:     req.Problems[i].Constraints,
			SampleTestCases: req.Problems[i].SampleTestCases,
		}
	}

	resp, err := pc.svc.AddProblemsBulk(items)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			utils.Fail(c, appErr.Status, appErr.Message)
			return
		}
		utils.Fail(c, http.StatusInternalServerError, "cannot create problems")
		return
	}

	utils.Success(c, http.StatusCreated, "problems created", resp)
}

func (pc *ProblemController) ListProblems(c *gin.Context) {
	resp, err := pc.svc.ListProblems()
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			utils.Fail(c, appErr.Status, appErr.Message)
			return
		}
		utils.Fail(c, http.StatusInternalServerError, "cannot list problems")
		return
	}

	utils.Success(c, http.StatusOK, "problems fetched", resp)
}

func (pc *ProblemController) GetProblemBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.Fail(c, http.StatusBadRequest, "missing slug")
		return
	}

	resp, err := pc.svc.GetProblemBySlug(slug)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			utils.Fail(c, appErr.Status, appErr.Message)
			return
		}
		utils.Fail(c, http.StatusInternalServerError, "cannot fetch problem")
		return
	}

	utils.Success(c, http.StatusOK, "problem fetched", resp)
}

func (pc *ProblemController) CreateTopic(c *gin.Context) {
	var req CreateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pc.log.Error("failed to parse create topic request", zap.Error(err))
		utils.ValidationFail(c, "Validation Failed", utils.ValidationErrors(err))
		return
	}

	resp, err := pc.svc.AddTopic(services.TopicInput{Name: req.Name})
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			utils.Fail(c, appErr.Status, appErr.Message)
			return
		}
		utils.Fail(c, http.StatusInternalServerError, "cannot create topic")
		return
	}

	utils.Success(c, http.StatusCreated, "topic created", resp)
}

func (pc *ProblemController) ListTopics(c *gin.Context) {
	resp, err := pc.svc.ListTopics()
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			utils.Fail(c, appErr.Status, appErr.Message)
			return
		}
		utils.Fail(c, http.StatusInternalServerError, "cannot list topics")
		return
	}

	utils.Success(c, http.StatusOK, "topics fetched", resp)
}

func (pc *ProblemController) BulkCreateTopics(c *gin.Context) {
	var req BulkCreateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pc.log.Error("failed to parse bulk create topic request", zap.Error(err))
		utils.ValidationFail(c, "Validation Failed", utils.ValidationErrors(err))
		return
	}

	inputs := make([]services.TopicInput, len(req.Topics))
	for i := range req.Topics {
		inputs[i] = services.TopicInput{Name: req.Topics[i].Name}
	}

	resp, err := pc.svc.AddTopicsBulk(inputs)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			utils.Fail(c, appErr.Status, appErr.Message)
			return
		}
		utils.Fail(c, http.StatusInternalServerError, "cannot create topics")
		return
	}

	utils.Success(c, http.StatusCreated, "topics created", resp)
}
