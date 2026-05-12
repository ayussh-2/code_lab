// submissions.go: HTTP layer for /api/submissions. Thin: parse the request,
// hand off to the service, format the response. Auth + role checks live in
// the route group middleware.
package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ayussh-2/internal/services"
	"github.com/ayussh-2/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SubmissionController struct {
	log *zap.Logger
	svc *services.SubmissionService
}

type CreateSubmissionRequest struct {
	ProblemSlug string `json:"problem_slug" binding:"required"`
	Language    string `json:"language" binding:"required"`
	SourceCode  string `json:"source_code" binding:"required"`
	Kind        string `json:"kind" binding:"required,oneof=submit run"`
}

func NewSubmissionController(log *zap.Logger, svc *services.SubmissionService) *SubmissionController {
	return &SubmissionController{log: log, svc: svc}
}

func (sc *SubmissionController) CreateSubmission(c *gin.Context) {
	var req CreateSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sc.log.Error("failed to parse create submission request", zap.Error(err))
		utils.ValidationFail(c, "Validation Failed", utils.ValidationErrors(err))
		return
	}

	userID, ok := readUserID(c)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, "missing user")
		return
	}

	sub, err := sc.svc.Create(services.CreateSubmissionInput{
		UserID:      userID,
		ProblemSlug: req.ProblemSlug,
		Language:    req.Language,
		SourceCode:  req.SourceCode,
		Kind:        req.Kind,
	})
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			utils.Fail(c, appErr.Status, appErr.Message)
			return
		}
		utils.Fail(c, http.StatusInternalServerError, "cannot create submission")
		return
	}

	utils.Success(c, http.StatusCreated, "submission created", gin.H{"submission_id": sub.ID})
}

func (sc *SubmissionController) GetSubmission(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		utils.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	userID, ok := readUserID(c)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, "missing user")
		return
	}
	role, _ := c.Get("role")
	isAdmin := role == "admin"

	sub, err := sc.svc.GetByID(id, userID, isAdmin)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			utils.Fail(c, appErr.Status, appErr.Message)
			return
		}
		utils.Fail(c, http.StatusInternalServerError, "cannot fetch submission")
		return
	}

	utils.Success(c, http.StatusOK, "submission fetched", sub)
}

func (sc *SubmissionController) ListSubmissions(c *gin.Context) {
	userID, ok := readUserID(c)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, "missing user")
		return
	}

	opts := services.ListSubmissionsOpts{
		ProblemSlug: c.Query("problem_slug"),
		Kind:        c.Query("kind"),
	}
	if rawLimit := c.Query("limit"); rawLimit != "" {
		if v, err := strconv.Atoi(rawLimit); err == nil {
			opts.Limit = v
		}
	}

	subs, err := sc.svc.ListForUser(userID, opts)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			utils.Fail(c, appErr.Status, appErr.Message)
			return
		}
		utils.Fail(c, http.StatusInternalServerError, "cannot list submissions")
		return
	}

	utils.Success(c, http.StatusOK, "submissions fetched", subs)
}

func readUserID(c *gin.Context) (uint, bool) {
	v, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	switch id := v.(type) {
	case uint:
		return id, true
	case float64:
		return uint(id), true
	case int:
		return uint(id), true
	default:
		return 0, false
	}
}

func parseIDParam(c *gin.Context, name string) (uint, error) {
	raw := c.Param(name)
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}
