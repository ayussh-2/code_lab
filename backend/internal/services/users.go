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

var usernamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,28}[a-z0-9]$`)

type UserService struct {
	log *zap.Logger
	db  *gorm.DB
	cfg *config.Config
}

func NewUserService(log *zap.Logger, db *gorm.DB, cfg *config.Config) *UserService {
	return &UserService{log: log, db: db, cfg: cfg}
}

type PublicProfile struct {
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url"`
	Bio       string    `json:"bio"`
	Rating    int       `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
}

type OwnProfile struct {
	PublicProfile
	Email string `json:"email"`
	Role  string `json:"role"`
	ID    uint   `json:"id"`
}

type UpdateProfileInput struct {
	Name      *string
	Username  *string
	AvatarURL *string
	Bio       *string
}

type SolvedByDifficulty struct {
	Easy   int `json:"easy"`
	Medium int `json:"medium"`
	Hard   int `json:"hard"`
}

type UserStats struct {
	SolvedByDifficulty SolvedByDifficulty `json:"solved_by_difficulty"`
	TotalSolved        int                `json:"total_solved"`
	AcceptanceRate     float64            `json:"acceptance_rate"`
	TotalSubmits       int64              `json:"total_submits"`
	TotalAC            int64              `json:"total_ac"`
}

type RatingPoint struct {
	Rating     int       `json:"rating"`
	RecordedAt time.Time `json:"recorded_at"`
}

type ActivityDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type ProfileSubmission struct {
	ID          uint      `json:"id"`
	ProblemID   uint      `json:"problem_id"`
	ProblemSlug string    `json:"problem_slug"`
	ProblemTitle string   `json:"problem_title"`
	Language    string    `json:"language"`
	Verdict     string    `json:"verdict"`
	Status      string    `json:"status"`
	RuntimeMs   int       `json:"runtime_ms"`
	CreatedAt   time.Time `json:"created_at"`
}

func (us *UserService) GetPublicProfile(username string) (*PublicProfile, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	if username == "" {
		return nil, utils.NewAppError(http.StatusBadRequest, "username is required", nil)
	}

	var user models.User
	if err := us.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAppError(http.StatusNotFound, "user not found", err)
		}
		us.log.Error("failed to fetch public profile", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot fetch profile", err)
	}

	return us.toPublicProfile(&user), nil
}

func (us *UserService) GetOwnProfile(userID uint) (*OwnProfile, error) {
	var user models.User
	if err := us.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAppError(http.StatusNotFound, "user not found", err)
		}
		us.log.Error("failed to fetch own profile", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot fetch profile", err)
	}

	pub := us.toPublicProfile(&user)
	return &OwnProfile{
		PublicProfile: *pub,
		Email:         user.Email,
		Role:          user.Role,
		ID:            user.ID,
	}, nil
}

func (us *UserService) UpdateProfile(userID uint, input UpdateProfileInput) (*OwnProfile, error) {
	var user models.User
	if err := us.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAppError(http.StatusNotFound, "user not found", err)
		}
		us.log.Error("failed to fetch user for update", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot update profile", err)
	}

	updates := map[string]any{}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if len(name) < 2 {
			return nil, utils.NewAppError(http.StatusBadRequest, "name must be at least 2 characters", nil)
		}
		updates["name"] = name
	}

	if input.Username != nil {
		username := strings.ToLower(strings.TrimSpace(*input.Username))
		if !usernamePattern.MatchString(username) {
			return nil, utils.NewAppError(http.StatusBadRequest, "username must be 3-30 chars: lowercase letters, numbers, underscore, hyphen", nil)
		}
		var existing models.User
		err := us.db.Where("username = ? AND id <> ?", username, userID).First(&existing).Error
		if err == nil {
			return nil, utils.NewAppError(http.StatusConflict, "username already taken", nil)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			us.log.Error("failed to check username availability", zap.Error(err))
			return nil, utils.NewAppError(http.StatusInternalServerError, "cannot update profile", err)
		}
		updates["username"] = username
	}

	if input.AvatarURL != nil {
		updates["avatar_url"] = strings.TrimSpace(*input.AvatarURL)
	}

	if input.Bio != nil {
		bio := strings.TrimSpace(*input.Bio)
		if len(bio) > 500 {
			return nil, utils.NewAppError(http.StatusBadRequest, "bio must be at most 500 characters", nil)
		}
		updates["bio"] = bio
	}

	if len(updates) == 0 {
		return us.GetOwnProfile(userID)
	}

	if err := us.db.Model(&user).Updates(updates).Error; err != nil {
		us.log.Error("failed to update profile", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot update profile", err)
	}

	return us.GetOwnProfile(userID)
}

func (us *UserService) GetStats(username string) (*UserStats, error) {
	user, err := us.userByUsername(username)
	if err != nil {
		return nil, err
	}

	type row struct {
		Difficulty string
		Count      int
	}
	var solvedRows []row
	err = us.db.Model(&models.Submission{}).
		Select("problems.difficulty, COUNT(DISTINCT submissions.problem_id) as count").
		Joins("JOIN problems ON problems.id = submissions.problem_id").
		Where("submissions.user_id = ? AND submissions.verdict = ? AND submissions.kind = ?",
			user.ID, models.VerdictAC, models.SubmissionKindSubmit).
		Group("problems.difficulty").
		Scan(&solvedRows).Error
	if err != nil {
		us.log.Error("failed to compute solved by difficulty", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot fetch stats", err)
	}

	stats := &UserStats{SolvedByDifficulty: SolvedByDifficulty{}}
	for _, r := range solvedRows {
		switch strings.ToLower(r.Difficulty) {
		case "easy":
			stats.SolvedByDifficulty.Easy = r.Count
		case "hard":
			stats.SolvedByDifficulty.Hard = r.Count
		default:
			stats.SolvedByDifficulty.Medium += r.Count
		}
	}
	stats.TotalSolved = stats.SolvedByDifficulty.Easy + stats.SolvedByDifficulty.Medium + stats.SolvedByDifficulty.Hard

	var totalSubmits int64
	if err := us.db.Model(&models.Submission{}).
		Where("user_id = ? AND kind = ?", user.ID, models.SubmissionKindSubmit).
		Count(&totalSubmits).Error; err != nil {
		us.log.Error("failed to count user submits", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot fetch stats", err)
	}

	var totalAC int64
	if err := us.db.Model(&models.Submission{}).
		Where("user_id = ? AND kind = ? AND verdict = ?", user.ID, models.SubmissionKindSubmit, models.VerdictAC).
		Count(&totalAC).Error; err != nil {
		us.log.Error("failed to count user AC", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot fetch stats", err)
	}

	stats.TotalSubmits = totalSubmits
	stats.TotalAC = totalAC
	stats.AcceptanceRate = acceptanceRate(totalAC, totalSubmits)

	return stats, nil
}

func (us *UserService) GetRatingHistory(username string, limit int) ([]RatingPoint, error) {
	user, err := us.userByUsername(username)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 500 {
		limit = 100
	}

	var rows []models.RatingHistory
	if err := us.db.Where("user_id = ?", user.ID).
		Order("recorded_at asc").
		Limit(limit).
		Find(&rows).Error; err != nil {
		us.log.Error("failed to fetch rating history", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot fetch rating history", err)
	}

	if len(rows) == 0 {
		return []RatingPoint{{Rating: user.Rating, RecordedAt: user.CreatedAt}}, nil
	}

	out := make([]RatingPoint, len(rows))
	for i := range rows {
		out[i] = RatingPoint{Rating: rows[i].Rating, RecordedAt: rows[i].RecordedAt}
	}
	return out, nil
}

func (us *UserService) GetActivityHeatmap(username string) ([]ActivityDay, error) {
	user, err := us.userByUsername(username)
	if err != nil {
		return nil, err
	}

	since := time.Now().AddDate(-1, 0, 0).Truncate(24 * time.Hour)

	type row struct {
		Day   time.Time
		Count int
	}
	var rows []row
	err = us.db.Model(&models.Submission{}).
		Select("created_at::date as day, COUNT(*) as count").
		Where("user_id = ? AND kind = ? AND created_at >= ?", user.ID, models.SubmissionKindSubmit, since).
		Group("created_at::date").
		Order("day asc").
		Scan(&rows).Error
	if err != nil {
		us.log.Error("failed to fetch activity heatmap", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot fetch activity", err)
	}

	out := make([]ActivityDay, len(rows))
	for i := range rows {
		out[i] = ActivityDay{
			Date:  rows[i].Day.Format("2006-01-02"),
			Count: rows[i].Count,
		}
	}
	return out, nil
}

func (us *UserService) ListProfileSubmissions(username string, limit int) ([]ProfileSubmission, error) {
	user, err := us.userByUsername(username)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	type row struct {
		models.Submission
		ProblemSlug  string
		ProblemTitle string
	}

	var rows []row
	err = us.db.Model(&models.Submission{}).
		Select("submissions.*, problems.slug as problem_slug, problems.title as problem_title").
		Joins("JOIN problems ON problems.id = submissions.problem_id").
		Where("submissions.user_id = ? AND submissions.kind = ?", user.ID, models.SubmissionKindSubmit).
		Order("submissions.id DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		us.log.Error("failed to list profile submissions", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot list submissions", err)
	}

	out := make([]ProfileSubmission, len(rows))
	for i := range rows {
		out[i] = ProfileSubmission{
			ID:           rows[i].ID,
			ProblemID:    rows[i].ProblemID,
			ProblemSlug:  rows[i].ProblemSlug,
			ProblemTitle: rows[i].ProblemTitle,
			Language:     rows[i].Language,
			Verdict:      rows[i].Verdict,
			Status:       rows[i].Status,
			RuntimeMs:    rows[i].RuntimeMs,
			CreatedAt:    rows[i].CreatedAt,
		}
	}
	return out, nil
}

func (us *UserService) EnsureUsername(user *models.User) error {
	if strings.TrimSpace(user.Username) != "" {
		return nil
	}
	username, err := us.generateUniqueUsername(user.Name, user.Email)
	if err != nil {
		return err
	}
	return us.db.Model(user).Update("username", username).Error
}

func (us *UserService) RecordRatingOnFirstAC(userID, problemID uint) error {
	var acCount int64
	if err := us.db.Model(&models.Submission{}).
		Where("user_id = ? AND problem_id = ? AND kind = ? AND verdict = ?",
			userID, problemID, models.SubmissionKindSubmit, models.VerdictAC).
		Count(&acCount).Error; err != nil {
		return err
	}
	if acCount != 1 {
		return nil
	}

	var problem models.Problems
	if err := us.db.Select("difficulty").First(&problem, problemID).Error; err != nil {
		return err
	}

	var user models.User
	if err := us.db.First(&user, userID).Error; err != nil {
		return err
	}

	delta := ratingDeltaForDifficulty(problem.Difficulty)
	newRating := user.Rating + delta
	if newRating < 0 {
		newRating = 0
	}

	now := time.Now()
	return us.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Update("rating", newRating).Error; err != nil {
			return err
		}
		return tx.Create(&models.RatingHistory{
			UserID:     userID,
			Rating:     newRating,
			RecordedAt: now,
		}).Error
	})
}

func (us *UserService) SeedInitialRatingHistory(userID uint, rating int, createdAt time.Time) error {
	var count int64
	if err := us.db.Model(&models.RatingHistory{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return us.db.Create(&models.RatingHistory{
		UserID:     userID,
		Rating:     rating,
		RecordedAt: createdAt,
	}).Error
}

func (us *UserService) BackfillUsernames() error {
	var users []models.User
	if err := us.db.Where("username = '' OR username IS NULL").Find(&users).Error; err != nil {
		return err
	}
	for i := range users {
		if err := us.EnsureUsername(&users[i]); err != nil {
			us.log.Warn("failed to backfill username", zap.Uint("user_id", users[i].ID), zap.Error(err))
		}
	}
	return nil
}

func (us *UserService) userByUsername(username string) (*models.User, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	var user models.User
	if err := us.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAppError(http.StatusNotFound, "user not found", err)
		}
		us.log.Error("failed to lookup user by username", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot fetch user", err)
	}
	return &user, nil
}

func (us *UserService) toPublicProfile(user *models.User) *PublicProfile {
	return &PublicProfile{
		Username:  user.Username,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		Bio:       user.Bio,
		Rating:    user.Rating,
		CreatedAt: user.CreatedAt,
	}
}

func (us *UserService) generateUniqueUsername(name, email string) (string, error) {
	base := slugifyUsername(name)
	if base == "" {
		parts := strings.Split(strings.ToLower(email), "@")
		base = slugifyUsername(parts[0])
	}
	if base == "" {
		base = "user"
	}
	if len(base) > 24 {
		base = base[:24]
	}

	candidate := base
	for i := 0; i < 100; i++ {
		if i > 0 {
			candidate = fmt.Sprintf("%s%d", base, i)
		}
		var count int64
		if err := us.db.Model(&models.User{}).Where("username = ?", candidate).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 && usernamePattern.MatchString(candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not generate unique username")
}

func slugifyUsername(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 0 && !regexp.MustCompile(`^[a-z]`).MatchString(s) {
		s = "u" + s
	}
	return s
}

func ratingDeltaForDifficulty(difficulty string) int {
	switch strings.ToLower(difficulty) {
	case "easy":
		return 10
	case "hard":
		return 25
	default:
		return 15
	}
}

func acceptanceRate(ac, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(ac) / float64(total)
}
