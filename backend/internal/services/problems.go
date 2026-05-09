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
	Examples        []Example
	Constraints     []string
	SampleTestCases []SampleTestCases
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ProblemListItem struct {
	ID         uint     `json:"id"`
	Title      string   `json:"title"`
	Slug       string   `json:"slug"`
	Difficulty string   `json:"difficulty"`
	Topics     []string `json:"topics"`
}

type ProblemDetail struct {
	ID              uint                     `json:"id"`
	Title           string                   `json:"title"`
	Slug            string                   `json:"slug"`
	Difficulty      string                   `json:"difficulty"`
	Topics          []string                 `json:"topics"`
	TopicIDs        []uint                   `json:"topic_ids"`
	Hint            []string                 `json:"hints"`
	Details         string                   `json:"details"`
	Examples        []models.Example         `json:"examples"`
	Constraints     []string                 `json:"constraints"`
	SampleTestCases []models.SampleTestCases `json:"sample_test_cases"`
	CreatedAt       time.Time                `json:"created_at"`
	UpdatedAt       time.Time                `json:"updated_at"`
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

func (ps *ProblemService) ListProblems() ([]ProblemListItem, error) {
	var problems []models.Problems
	if err := ps.db.Select("id", "title", "slug", "difficulty", "topics").Order("created_at desc").Find(&problems).Error; err != nil {
		ps.log.Error("failed to list problems", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot list problems", err)
	}

	resp := make([]ProblemListItem, 0, len(problems))
	for i := range problems {
		resp = append(resp, ProblemListItem{
			ID:         problems[i].ID,
			Title:      problems[i].Title,
			Slug:       problems[i].Slug,
			Difficulty: problems[i].Difficulty,
			Topics:     ps.topicNamesByIDs(problems[i].Topics),
		})
	}
	return resp, nil
}

func (ps *ProblemService) GetProblemBySlug(slug string) (*ProblemDetail, error) {
	var problem models.Problems
	err := ps.db.Preload("SampleTestCases").Where("slug = ?", slug).First(&problem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAppError(http.StatusNotFound, "problem not found", err)
		}
		ps.log.Error("failed to fetch problem by slug", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot fetch problem", err)
	}

	resp := ProblemDetail{
		ID:              problem.ID,
		Title:           problem.Title,
		Slug:            problem.Slug,
		Difficulty:      problem.Difficulty,
		Topics:          ps.topicNamesByIDs(problem.Topics),
		TopicIDs:        problem.Topics,
		Hint:            problem.Hint,
		Details:         problem.Details,
		Examples:        problem.Examples,
		Constraints:     problem.Constraints,
		SampleTestCases: problem.SampleTestCases,
		CreatedAt:       problem.CreatedAt,
		UpdatedAt:       problem.UpdatedAt,
	}

	return &resp, nil
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
		}
		if err := tx.Model(&models.Problems{}).
			Where("id = ?", existing.ID).
			Select("Title", "Difficulty", "Topics", "Hint", "Details", "Examples", "Constraints").
			Updates(&patch).Error; err != nil {
			ps.log.Error("failed to update problem", zap.Error(err))
			return utils.NewAppError(http.StatusInternalServerError, "cannot update problem", err)
		}

		if err := tx.Where("problem_id = ?", existing.ID).Delete(&models.SampleTestCases{}).Error; err != nil {
			ps.log.Error("failed to clear sample test cases", zap.Error(err))
			return utils.NewAppError(http.StatusInternalServerError, "cannot update problem", err)
		}

		if len(p.SampleTestCases) > 0 {
			samples := make([]models.SampleTestCases, len(p.SampleTestCases))
			for i := range p.SampleTestCases {
				samples[i] = models.SampleTestCases{
					ProblemID: existing.ID,
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
	if err := ps.db.Preload("SampleTestCases").Where("id = ?", existing.ID).First(&refreshed).Error; err != nil {
		ps.log.Error("failed to reload problem after update", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot update problem", err)
	}
	return &refreshed, nil
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

		sampleTestCases := make([]models.SampleTestCases, len(p.SampleTestCases))
		for i := range p.SampleTestCases {
			sampleTestCases[i] = models.SampleTestCases{
				ProblemID: problem.ID,
				Input:     p.SampleTestCases[i].Input,
				Expected:  p.SampleTestCases[i].Expected,
			}
		}

		if err := tx.Create(&sampleTestCases).Error; err != nil {
			ps.log.Error("failed to create sample test cases", zap.Error(err))
			return utils.NewAppError(http.StatusInternalServerError, "cannot create problem", err)
		}

		problem.SampleTestCases = sampleTestCases
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
