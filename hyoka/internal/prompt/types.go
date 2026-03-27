package prompt

// Prompt represents a parsed .prompt.md file with frontmatter metadata and prompt text.
type Prompt struct {
	ID              string            `yaml:"id" json:"id"`
	Service         string            `yaml:"service" json:"service"`
	Plane           string            `yaml:"plane" json:"plane"`
	Language        string            `yaml:"language" json:"language"`
	Category        string            `yaml:"category" json:"category"`
	Difficulty      string            `yaml:"difficulty" json:"difficulty"`
	Description     string            `yaml:"description" json:"description"`
	SDKPackage      string            `yaml:"sdk_package" json:"sdk_package"`
	DocURL          string            `yaml:"doc_url" json:"doc_url"`
	Tags            []string          `yaml:"tags" json:"tags"`
	Created         string            `yaml:"created" json:"created"`
	Author          string            `yaml:"author" json:"author"`
	ProjectContext  map[string]string `yaml:"project_context" json:"project_context,omitempty"`
	StarterProject  string            `yaml:"starter_project" json:"starter_project,omitempty"`
	ReferenceAnswer string            `yaml:"reference_answer" json:"reference_answer,omitempty"`
	Timeout         int               `yaml:"timeout" json:"timeout,omitempty"`
	ExpectedPkgs    []string          `yaml:"expected_packages" json:"expected_packages,omitempty"`
	ExpectedTools   []string          `yaml:"expected_tools" json:"expected_tools,omitempty"`

	// The prompt text extracted from the ## Prompt section
	PromptText string `yaml:"-" json:"prompt_text"`

	// The evaluation criteria text extracted from the ## Evaluation Criteria section
	EvaluationCriteria string `yaml:"-" json:"evaluation_criteria,omitempty"`

	// Source file path
	FilePath string `yaml:"-" json:"file_path"`
}

// Filter defines criteria for selecting prompts.
type Filter struct {
	Service  string
	Plane    string
	Language string
	Category string
	Tags     []string
	PromptID string
}

// Matches returns true if the prompt matches all non-empty filter criteria.
func (p *Prompt) Matches(f Filter) bool {
	if f.PromptID != "" && p.ID != f.PromptID {
		return false
	}
	if f.Service != "" && p.Service != f.Service {
		return false
	}
	if f.Plane != "" && p.Plane != f.Plane {
		return false
	}
	if f.Language != "" && p.Language != f.Language {
		return false
	}
	if f.Category != "" && p.Category != f.Category {
		return false
	}
	if len(f.Tags) > 0 {
		tagSet := make(map[string]bool, len(p.Tags))
		for _, t := range p.Tags {
			tagSet[t] = true
		}
		for _, required := range f.Tags {
			if !tagSet[required] {
				return false
			}
		}
	}
	return true
}
