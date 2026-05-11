package models

import "time"

const (
	SubmissionStatusQueued  = "queued"
	SubmissionStatusRunning = "running"
	SubmissionStatusDone    = "done"

	SubmissionKindSubmit = "submit"
	SubmissionKindRun    = "run"

	VerdictPending = "PENDING"
	VerdictAC      = "AC"
	VerdictWA      = "WA"
	VerdictTLE     = "TLE"
	VerdictMLE     = "MLE"
	VerdictCE      = "CE"
	VerdictRE      = "RE"
	VerdictIE      = "IE"
)

type Submission struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	UserID     uint   `gorm:"not null;index" json:"user_id"`
	ProblemID  uint   `gorm:"not null;index" json:"problem_id"`
	Language   string `gorm:"not null;size:32" json:"language"`
	Kind       string `gorm:"not null;size:16;default:submit;index" json:"kind"`
	SourceCode string `gorm:"type:text;not null" json:"source_code"`
	Status     string `gorm:"not null;size:16;default:queued;index" json:"status"`
	Verdict    string `gorm:"not null;size:16;default:PENDING;index" json:"verdict"`

	RuntimeMs int `gorm:"default:0" json:"runtime_ms"`
	MemoryKB  int `gorm:"default:0" json:"memory_kb"`
	Score     int `gorm:"default:0" json:"score"`

	FailedTestCaseID      *uint  `json:"failed_test_case_id,omitempty"`
	FailedInputPreview    string `gorm:"type:text" json:"failed_input_preview,omitempty"`
	FailedExpectedPreview string `gorm:"type:text" json:"failed_expected_preview,omitempty"`
	FailedActualPreview   string `gorm:"type:text" json:"failed_actual_preview,omitempty"`

	CompilerOutput string `gorm:"type:text" json:"compiler_output,omitempty"`
	StderrPreview  string `gorm:"type:text" json:"stderr_preview,omitempty"`
	Error          string `gorm:"type:text" json:"error,omitempty"`

	Results []SubmissionTestResult `gorm:"foreignKey:SubmissionID;constraint:OnDelete:CASCADE" json:"results,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	JudgedAt  *time.Time `json:"judged_at,omitempty"`
}

type SubmissionTestResult struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	SubmissionID        uint      `gorm:"not null;index" json:"submission_id"`
	TestCaseID          uint      `gorm:"not null;index" json:"test_case_id"`
	Verdict             string    `gorm:"not null;size:16" json:"verdict"`
	RuntimeMs           int       `json:"runtime_ms"`
	MemoryKB            int       `json:"memory_kb"`
	ActualOutputPreview string    `gorm:"type:text" json:"actual_output_preview,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}
