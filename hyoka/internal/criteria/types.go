// Package criteria provides a tiered evaluation criteria system.
// Tier 1: General criteria (always applied, from rubric.md).
// Tier 2: Attribute-matched criteria (applied when prompt metadata matches).
// Tier 3: Prompt-specific criteria (per-prompt, from ## Evaluation Criteria).
package criteria

// Criterion defines a single evaluation criterion with a name and description.
type Criterion struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
}

// MatchRule defines conditions under which a criteria set applies.
// All non-empty fields must match (AND logic).
type MatchRule struct {
	Language string `yaml:"language" json:"language,omitempty"`
	Service  string `yaml:"service" json:"service,omitempty"`
	Plane    string `yaml:"plane" json:"plane,omitempty"`
	Category string `yaml:"category" json:"category,omitempty"`
}

// CriteriaSet is a YAML file defining criteria that activate when match conditions are met.
type CriteriaSet struct {
	Match    MatchRule   `yaml:"match" json:"match"`
	Criteria []Criterion `yaml:"criteria" json:"criteria"`

	// Source records which file this criteria set was loaded from (not serialized to YAML).
	Source string `yaml:"-" json:"source,omitempty"`
}

// PromptAttributes holds the metadata extracted from a prompt for matching purposes.
type PromptAttributes struct {
	Language string
	Service  string
	Plane    string
	Category string
}

// Matches returns true if all non-empty fields in the match rule are satisfied
// by the given prompt attributes. Empty fields are wildcards.
func (m MatchRule) Matches(attrs PromptAttributes) bool {
	if m.Language != "" && m.Language != attrs.Language {
		return false
	}
	if m.Service != "" && m.Service != attrs.Service {
		return false
	}
	if m.Plane != "" && m.Plane != attrs.Plane {
		return false
	}
	if m.Category != "" && m.Category != attrs.Category {
		return false
	}
	return true
}
