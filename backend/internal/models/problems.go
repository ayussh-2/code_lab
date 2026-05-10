package models

import "time"

const (
	TestCaseKindSample = "sample"
	TestCaseKindHidden = "hidden"
)

type Example struct {
	Input       string `json:"input"`
	Output      string `json:"output"`
	Explanation string `json:"explanation"`
}

type Problems struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Title       string     `gorm:"not null" json:"title"`
	Slug        string     `gorm:"uniqueIndex;not null" json:"slug"`
	Difficulty  string     `gorm:"not null;default:medium" json:"difficulty"`
	Topics      []uint     `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"topics"`
	Hint        []string   `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"hints"`
	Details     string     `gorm:"type:text" json:"details"`
	Examples    []Example  `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"examples"`
	Constraints []string   `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"constraints"`
	TestCases   []TestCase `gorm:"foreignKey:ProblemID;constraint:OnDelete:CASCADE" json:"test_cases"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Topics struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"not null" json:"name"`
}

type TestCase struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProblemID uint      `gorm:"not null;index" json:"problem_id"`
	Kind      string    `gorm:"not null;size:16;default:sample;index" json:"kind"`
	Input     string    `gorm:"not null" json:"input"`
	Expected  string    `gorm:"not null" json:"expected"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (TestCase) TableName() string { return "test_cases" }
