package models

import "time"

type Example struct {
	Input       string `json:"input"`
	Output      string `json:"output"`
	Explanation string `json:"explanation"`
}

type Problems struct {
	ID              uint              `gorm:"primaryKey" json:"id"`
	Title           string            `gorm:"not null" json:"title"`
	Slug            string            `gorm:"uniqueIndex;not null" json:"slug"`
	Difficulty      string            `gorm:"not null;default:medium" json:"difficulty"`
	Topics          []uint            `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"topics"`
	Hint            []string          `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"hints"`
	Details         string            `gorm:"type:text" json:"details"`
	Examples        []Example         `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"examples"`
	Constraints     []string          `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"constraints"`
	SampleTestCases []SampleTestCases `gorm:"foreignKey:ProblemID;constraint:OnDelete:CASCADE" json:"sample_test_cases"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type Topics struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"not null" json:"name"`
}

type SampleTestCases struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProblemID uint      `gorm:"not null;index" json:"problem_id"`
	Input     string    `gorm:"not null" json:"input"`
	Expected  string    `gorm:"not null" json:"expected"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
