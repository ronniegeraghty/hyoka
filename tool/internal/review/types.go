package review

// ReviewScores holds individual dimension scores from the LLM-as-judge review.
type ReviewScores struct {
	Correctness         int `json:"correctness"`
	Completeness        int `json:"completeness"`
	BestPractices       int `json:"best_practices"`
	ErrorHandling       int `json:"error_handling"`
	PackageUsage        int `json:"package_usage"`
	CodeQuality         int `json:"code_quality"`
	ReferenceSimilarity int `json:"reference_similarity,omitempty"`
}

// ReviewEvent captures a single event from the review Copilot session.
type ReviewEvent struct {
	Type     string  `json:"type"`
	ToolName string  `json:"tool_name,omitempty"`
	ToolArgs string  `json:"tool_args,omitempty"`
	Content  string  `json:"content,omitempty"`
	Result   string  `json:"result,omitempty"`
	Error    string  `json:"error,omitempty"`
	Duration float64 `json:"duration_ms,omitempty"`
}

// ReviewResult holds the full output from an LLM-as-judge code review.
type ReviewResult struct {
	Scores       ReviewScores  `json:"scores"`
	OverallScore int           `json:"overall_score"`
	Summary      string        `json:"summary"`
	Issues       []string      `json:"issues"`
	Strengths    []string      `json:"strengths"`
	Events       []ReviewEvent `json:"events,omitempty"`
}
