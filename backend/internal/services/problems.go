package services

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/ayussh-2/config"
	"github.com/ayussh-2/internal/models"
	"github.com/ayussh-2/internal/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ProblemService struct {
	log *zap.Logger
	db  *gorm.DB
	cfg *config.Config
}

func NewProblemService(log *zap.Logger, db *gorm.DB, cfg *config.Config) *ProblemService {
	return &ProblemService{log: log, db: db, cfg: cfg}
}

type Topics struct {
	ID   uint
	Name string
}

type TopicInput struct {
	Name string
}

type SampleTestCases struct {
	ID        uint
	ProblemID uint
	Input     string
	Expected  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Example struct {
	Input       string
	Output      string
	Explanation string
}

type Problem struct {
	ID              uint
	Title           string
	Slug            string
	Difficulty      string
	Topics          []uint
	Hint            []string
	Details         string
	Editorial       string
	Examples        []Example
	Constraints     []string
	SampleTestCases []SampleTestCases
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ProblemListItem struct {
	ID             uint     `json:"id"`
	Title          string   `json:"title"`
	Slug           string   `json:"slug"`
	Difficulty     string   `json:"difficulty"`
	Topics         []string `json:"topics"`
	AcceptanceRate float64  `json:"acceptance_rate"`
	Status         string   `json:"status,omitempty"`
}

type ProblemListFilters struct {
	Difficulty string
	TopicID    uint
	Status     string
	UserID     uint
}

type ProblemDetail struct {
	ID                uint              `json:"id"`
	Title             string            `json:"title"`
	Slug              string            `json:"slug"`
	Difficulty        string            `json:"difficulty"`
	Topics            []string          `json:"topics"`
	TopicIDs          []uint            `json:"topic_ids"`
	Hint              []string          `json:"hints"`
	Details           string            `json:"details"`
	Examples          []models.Example  `json:"examples"`
	Constraints       []string          `json:"constraints"`
	SampleTestCases   []models.TestCase `json:"sample_test_cases"`
	AcceptanceRate    float64           `json:"acceptance_rate"`
	EditorialUnlocked bool              `json:"editorial_unlocked"`
	Editorial         string            `json:"editorial,omitempty"`
	Status            string            `json:"status,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

func (ps *ProblemService) AddProblem(p Problem) (*models.Problems, error) {
	return ps.addProblemWithTx(ps.db, p)
}

func (ps *ProblemService) AddProblemsBulk(items []Problem) ([]models.Problems, error) {
	if len(items) == 0 {
		return []models.Problems{}, nil
	}

	var created []models.Problems
	txErr := ps.db.Transaction(func(tx *gorm.DB) error {
		created = make([]models.Problems, 0, len(items))
		for i := range items {
			problem, err := ps.addProblemWithTx(tx, items[i])
			if err != nil {
				return err
			}
			created = append(created, *problem)
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}
	return created, nil
}

func (ps *ProblemService) ListProblems(filters ProblemListFilters) ([]ProblemListItem, error) {
	q := ps.db.Model(&models.Problems{}).Select("id", "title", "slug", "difficulty", "topics", "submit_count", "ac_count")

	if d := strings.ToLower(strings.TrimSpace(filters.Difficulty)); d != "" {
		q = q.Where("difficulty = ?", d)
	}

	if filters.TopicID > 0 {
		q = q.Where("topics @> ?", fmt.Sprintf("[%d]", filters.TopicID))
	}

	var problems []models.Problems
	if err := q.Order("created_at desc").Find(&problems).Error; err != nil {
		ps.log.Error("failed to list problems", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot list problems", err)
	}

	statusByProblem := map[uint]string{}
	if filters.UserID > 0 {
		statusByProblem = ps.userProblemStatuses(filters.UserID)
	}

	resp := make([]ProblemListItem, 0, len(problems))
	for i := range problems {
		status := statusByProblem[problems[i].ID]
		if filters.Status != "" && status != filters.Status {
			continue
		}
		resp = append(resp, ProblemListItem{
			ID:             problems[i].ID,
			Title:          problems[i].Title,
			Slug:           problems[i].Slug,
			Difficulty:     problems[i].Difficulty,
			Topics:         ps.topicNamesByIDs(problems[i].Topics),
			AcceptanceRate: problemAcceptanceRate(problems[i].ACCount, problems[i].SubmitCount),
			Status:         status,
		})
	}
	return resp, nil
}

func (ps *ProblemService) GetProblemBySlug(slug string, viewer UserViewer) (*ProblemDetail, error) {
	var problem models.Problems
	err := ps.db.
		Preload("TestCases", "kind = ?", models.TestCaseKindSample).
		Where("slug = ?", slug).
		First(&problem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAppError(http.StatusNotFound, "problem not found", err)
		}
		ps.log.Error("failed to fetch problem by slug", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot fetch problem", err)
	}

	unlocked := ps.editorialUnlocked(viewer, problem.ID)
	resp := ProblemDetail{
		ID:                problem.ID,
		Title:             problem.Title,
		Slug:              problem.Slug,
		Difficulty:        problem.Difficulty,
		Topics:            ps.topicNamesByIDs(problem.Topics),
		TopicIDs:          problem.Topics,
		Hint:              problem.Hint,
		Details:           problem.Details,
		Examples:          problem.Examples,
		Constraints:       problem.Constraints,
		SampleTestCases:   problem.TestCases,
		AcceptanceRate:    problemAcceptanceRate(problem.ACCount, problem.SubmitCount),
		EditorialUnlocked: unlocked,
		CreatedAt:         problem.CreatedAt,
		UpdatedAt:         problem.UpdatedAt,
	}
	if unlocked {
		resp.Editorial = problem.Editorial
	}
	if viewer.UserID > 0 {
		statuses := ps.userProblemStatuses(viewer.UserID)
		resp.Status = statuses[problem.ID]
	}

	return &resp, nil
}

type UserViewer struct {
	UserID uint
	Role   string
}

func (ps *ProblemService) UpdateProblemBySlug(slug string, p Problem) (*models.Problems, error) {
	var existing models.Problems
	if err := ps.db.Where("slug = ?", slug).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAppError(http.StatusNotFound, "problem not found", err)
		}
		ps.log.Error("failed to fetch problem for update", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot update problem", err)
	}

	examples := make([]models.Example, len(p.Examples))
	for i := range p.Examples {
		examples[i] = models.Example{
			Input:       p.Examples[i].Input,
			Output:      p.Examples[i].Output,
			Explanation: p.Examples[i].Explanation,
		}
	}

	txErr := ps.db.Transaction(func(tx *gorm.DB) error {
		patch := models.Problems{
			Title:       p.Title,
			Difficulty:  p.Difficulty,
			Topics:      p.Topics,
			Hint:        p.Hint,
			Details:     p.Details,
			Examples:    examples,
			Constraints: p.Constraints,
			Editorial:   p.Editorial,
		}
		if err := tx.Model(&models.Problems{}).
			Where("id = ?", existing.ID).
			Select("Title", "Difficulty", "Topics", "Hint", "Details", "Examples", "Constraints", "Editorial").
			Updates(&patch).Error; err != nil {
			ps.log.Error("failed to update problem", zap.Error(err))
			return utils.NewAppError(http.StatusInternalServerError, "cannot update problem", err)
		}

		if err := tx.Where("problem_id = ? AND kind = ?", existing.ID, models.TestCaseKindSample).
			Delete(&models.TestCase{}).Error; err != nil {
			ps.log.Error("failed to clear sample test cases", zap.Error(err))
			return utils.NewAppError(http.StatusInternalServerError, "cannot update problem", err)
		}

		if len(p.SampleTestCases) > 0 {
			samples := make([]models.TestCase, len(p.SampleTestCases))
			for i := range p.SampleTestCases {
				samples[i] = models.TestCase{
					ProblemID: existing.ID,
					Kind:      models.TestCaseKindSample,
					Input:     p.SampleTestCases[i].Input,
					Expected:  p.SampleTestCases[i].Expected,
				}
			}
			if err := tx.Create(&samples).Error; err != nil {
				ps.log.Error("failed to recreate sample test cases", zap.Error(err))
				return utils.NewAppError(http.StatusInternalServerError, "cannot update problem", err)
			}
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	var refreshed models.Problems
	if err := ps.db.
		Preload("TestCases", "kind = ?", models.TestCaseKindSample).
		Where("id = ?", existing.ID).
		First(&refreshed).Error; err != nil {
		ps.log.Error("failed to reload problem after update", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot update problem", err)
	}
	return &refreshed, nil
}

func (ps *ProblemService) ListHiddenTestCases(slug string) ([]models.TestCase, error) {
	var existing models.Problems
	if err := ps.db.Select("id").Where("slug = ?", slug).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAppError(http.StatusNotFound, "problem not found", err)
		}
		ps.log.Error("failed to fetch problem for hidden test cases", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot fetch hidden test cases", err)
	}

	var cases []models.TestCase
	if err := ps.db.
		Where("problem_id = ? AND kind = ?", existing.ID, models.TestCaseKindHidden).
		Order("id asc").
		Find(&cases).Error; err != nil {
		ps.log.Error("failed to list hidden test cases", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot fetch hidden test cases", err)
	}
	return cases, nil
}

func (ps *ProblemService) ReplaceHiddenTestCases(slug string, items []SampleTestCases) ([]models.TestCase, error) {
	var existing models.Problems
	if err := ps.db.Select("id").Where("slug = ?", slug).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAppError(http.StatusNotFound, "problem not found", err)
		}
		ps.log.Error("failed to fetch problem for hidden test cases", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot replace hidden test cases", err)
	}

	txErr := ps.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("problem_id = ? AND kind = ?", existing.ID, models.TestCaseKindHidden).
			Delete(&models.TestCase{}).Error; err != nil {
			ps.log.Error("failed to clear hidden test cases", zap.Error(err))
			return utils.NewAppError(http.StatusInternalServerError, "cannot replace hidden test cases", err)
		}

		if len(items) == 0 {
			return nil
		}

		hidden := make([]models.TestCase, 0, len(items))
		for i := range items {
			input := strings.TrimSpace(items[i].Input)
			expected := strings.TrimSpace(items[i].Expected)
			if input == "" || expected == "" {
				continue
			}
			hidden = append(hidden, models.TestCase{
				ProblemID: existing.ID,
				Kind:      models.TestCaseKindHidden,
				Input:     input,
				Expected:  expected,
			})
		}

		if len(hidden) == 0 {
			return nil
		}

		if err := tx.Create(&hidden).Error; err != nil {
			ps.log.Error("failed to insert hidden test cases", zap.Error(err))
			return utils.NewAppError(http.StatusInternalServerError, "cannot replace hidden test cases", err)
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return ps.ListHiddenTestCases(slug)
}

func (ps *ProblemService) DeleteProblemBySlug(slug string) error {
	var existing models.Problems
	if err := ps.db.Where("slug = ?", slug).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.NewAppError(http.StatusNotFound, "problem not found", err)
		}
		ps.log.Error("failed to fetch problem for delete", zap.Error(err))
		return utils.NewAppError(http.StatusInternalServerError, "cannot delete problem", err)
	}

	if err := ps.db.Delete(&existing).Error; err != nil {
		ps.log.Error("failed to delete problem", zap.Error(err))
		return utils.NewAppError(http.StatusInternalServerError, "cannot delete problem", err)
	}
	return nil
}

func (ps *ProblemService) AddTopic(input TopicInput) (*models.Topics, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, utils.NewAppError(http.StatusBadRequest, "topic name is required", nil)
	}

	var existing models.Topics
	err := ps.db.Where("name = ?", name).First(&existing).Error
	if err == nil {
		return nil, utils.NewAppError(http.StatusConflict, "topic already exists", nil)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		ps.log.Error("failed to check existing topic", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot create topic", err)
	}

	topic := models.Topics{Name: name}
	if err := ps.db.Create(&topic).Error; err != nil {
		ps.log.Error("failed to create topic", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot create topic", err)
	}

	return &topic, nil
}

func (ps *ProblemService) ListTopics() ([]models.Topics, error) {
	var topics []models.Topics
	if err := ps.db.Order("name asc").Find(&topics).Error; err != nil {
		ps.log.Error("failed to list topics", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot list topics", err)
	}
	return topics, nil
}

func (ps *ProblemService) AddTopicsBulk(inputs []TopicInput) ([]models.Topics, error) {
	if len(inputs) == 0 {
		return []models.Topics{}, nil
	}

	normalized := make([]string, len(inputs))
	seen := make(map[string]struct{}, len(inputs))
	for i := range inputs {
		name := strings.TrimSpace(inputs[i].Name)
		if name == "" {
			return nil, utils.NewAppError(http.StatusBadRequest, "topic name is required", nil)
		}
		key := strings.ToLower(name)
		if _, exists := seen[key]; exists {
			return nil, utils.NewAppError(http.StatusBadRequest, "duplicate topic name in request", nil)
		}
		seen[key] = struct{}{}
		normalized[i] = name
	}

	var created []models.Topics
	txErr := ps.db.Transaction(func(tx *gorm.DB) error {
		for i := range normalized {
			var existing models.Topics
			err := tx.Where("name = ?", normalized[i]).First(&existing).Error
			if err == nil {
				return utils.NewAppError(http.StatusConflict, "topic already exists", nil)
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				ps.log.Error("failed to check existing topic in bulk create", zap.Error(err))
				return utils.NewAppError(http.StatusInternalServerError, "cannot create topics", err)
			}

			topic := models.Topics{Name: normalized[i]}
			if err := tx.Create(&topic).Error; err != nil {
				ps.log.Error("failed to create topic in bulk create", zap.Error(err))
				return utils.NewAppError(http.StatusInternalServerError, "cannot create topics", err)
			}
			created = append(created, topic)
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return created, nil
}

func (ps *ProblemService) addProblemWithTx(db *gorm.DB, p Problem) (*models.Problems, error) {
	examples := make([]models.Example, len(p.Examples))
	for i := range p.Examples {
		examples[i] = models.Example{
			Input:       p.Examples[i].Input,
			Output:      p.Examples[i].Output,
			Explanation: p.Examples[i].Explanation,
		}
	}

	problem := models.Problems{
		Title:       p.Title,
		Difficulty:  p.Difficulty,
		Topics:      p.Topics,
		Hint:        p.Hint,
		Details:     p.Details,
		Editorial:   p.Editorial,
		Examples:    examples,
		Constraints: p.Constraints,
	}

	txErr := db.Transaction(func(tx *gorm.DB) error {
		slug, err := ps.generateUniqueSlug(tx, p.Title)
		if err != nil {
			ps.log.Error("failed to generate problem slug", zap.Error(err))
			return utils.NewAppError(http.StatusInternalServerError, "cannot create problem", err)
		}
		problem.Slug = slug

		if err := tx.Create(&problem).Error; err != nil {
			ps.log.Error("failed to create problem", zap.Error(err))
			return utils.NewAppError(http.StatusInternalServerError, "cannot create problem", err)
		}

		if len(p.SampleTestCases) == 0 {
			return nil
		}

		sampleTestCases := make([]models.TestCase, len(p.SampleTestCases))
		for i := range p.SampleTestCases {
			sampleTestCases[i] = models.TestCase{
				ProblemID: problem.ID,
				Kind:      models.TestCaseKindSample,
				Input:     p.SampleTestCases[i].Input,
				Expected:  p.SampleTestCases[i].Expected,
			}
		}

		if err := tx.Create(&sampleTestCases).Error; err != nil {
			ps.log.Error("failed to create sample test cases", zap.Error(err))
			return utils.NewAppError(http.StatusInternalServerError, "cannot create problem", err)
		}

		problem.TestCases = sampleTestCases
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return &problem, nil
}

func (ps *ProblemService) generateUniqueSlug(db *gorm.DB, title string) (string, error) {
	base := slugify(title)
	slug := base
	suffix := 1

	for {
		var count int64
		if err := db.Model(&models.Problems{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return slug, nil
		}
		suffix++
		slug = fmt.Sprintf("%s-%d", base, suffix)
	}
}

func slugify(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "problem"
	}
	return s
}

func (ps *ProblemService) topicNamesByIDs(ids []uint) []string {
	if len(ids) == 0 {
		return []string{}
	}

	var topics []models.Topics
	if err := ps.db.Select("id", "name").Where("id IN ?", ids).Find(&topics).Error; err != nil {
		ps.log.Error("failed to map topic ids to names", zap.Error(err))
		return []string{}
	}

	nameByID := make(map[uint]string, len(topics))
	for i := range topics {
		nameByID[topics[i].ID] = topics[i].Name
	}

	names := make([]string, 0, len(ids))
	for _, id := range ids {
		if name, ok := nameByID[id]; ok {
			names = append(names, name)
		}
	}

	return names
}

func (ps *ProblemService) IncrementSubmissionStats(problemID uint, verdict string) error {
	updates := map[string]any{"submit_count": gorm.Expr("submit_count + 1")}
	if verdict == models.VerdictAC {
		updates["ac_count"] = gorm.Expr("ac_count + 1")
	}
	return ps.db.Model(&models.Problems{}).Where("id = ?", problemID).Updates(updates).Error
}

func (ps *ProblemService) editorialUnlocked(viewer UserViewer, problemID uint) bool {
	if viewer.Role == "admin" || viewer.Role == "problem_setter" {
		return true
	}
	if viewer.UserID == 0 {
		return false
	}
	var count int64
	err := ps.db.Model(&models.Submission{}).
		Where("user_id = ? AND problem_id = ? AND kind = ? AND verdict = ?",
			viewer.UserID, problemID, models.SubmissionKindSubmit, models.VerdictAC).
		Count(&count).Error
	return err == nil && count > 0
}

func (ps *ProblemService) userProblemStatuses(userID uint) map[uint]string {
	statusByProblem := map[uint]string{}

	type row struct {
		ProblemID uint
		Verdict   string
	}
	var rows []row
	if err := ps.db.Model(&models.Submission{}).
		Select("problem_id, verdict").
		Where("user_id = ? AND kind = ?", userID, models.SubmissionKindSubmit).
		Order("id asc").
		Find(&rows).Error; err != nil {
		ps.log.Error("failed to load user problem statuses", zap.Error(err))
		return statusByProblem
	}

	for _, r := range rows {
		if r.Verdict == models.VerdictAC {
			statusByProblem[r.ProblemID] = "solved"
			continue
		}
		if statusByProblem[r.ProblemID] != "solved" {
			statusByProblem[r.ProblemID] = "attempted"
		}
	}

	return statusByProblem
}

func problemAcceptanceRate(acCount, submitCount int64) float64 {
	if submitCount == 0 {
		return 0
	}
	return float64(acCount) / float64(submitCount)
}
