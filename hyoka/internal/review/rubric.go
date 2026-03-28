package review

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed rubric.md
var embeddedRubric string

// BuildReviewPrompt constructs a structured review prompt for the LLM-as-judge.
// It includes the original prompt, generated code files, optional reference answer,
// attribute-matched criteria (Tier 2), and prompt-specific evaluation criteria (Tier 3).
func BuildReviewPrompt(originalPrompt string, generatedFiles map[string]string, referenceFiles map[string]string, evaluationCriteria string) string {
	return BuildReviewPromptTiered(originalPrompt, generatedFiles, referenceFiles, "", evaluationCriteria)
}

// BuildReviewPromptTiered constructs a review prompt with all three tiers of criteria:
//   - Tier 1 (General): embedded rubric.md, always applied
//   - Tier 2 (Attribute-Matched): criteria matched by prompt metadata (language, service, etc.)
//   - Tier 3 (Prompt-Specific): per-prompt criteria from ## Evaluation Criteria section
func BuildReviewPromptTiered(originalPrompt string, generatedFiles map[string]string, referenceFiles map[string]string, attributeMatchedCriteria string, promptSpecificCriteria string) string {
	var b strings.Builder

	b.WriteString("You are evaluating another AI agent's work. The agent was given the prompt below ")
	b.WriteString("and asked to produce code. Review the generated code against the original prompt, ")
	b.WriteString("the general scoring rubric, and any additional evaluation criteria.\n\n")

	b.WriteString("## Original Prompt\n\n")
	b.WriteString(originalPrompt)
	b.WriteString("\n\n")

	if attributeMatchedCriteria != "" {
		b.WriteString("## Attribute-Matched Evaluation Criteria\n\n")
		b.WriteString("These criteria apply based on the prompt's language, service, or other attributes. ")
		b.WriteString("Evaluate EACH criterion individually as pass/fail:\n\n")
		b.WriteString(attributeMatchedCriteria)
		b.WriteString("\n\n")
	}

	if promptSpecificCriteria != "" {
		b.WriteString("## Prompt-Specific Evaluation Criteria\n\n")
		b.WriteString("The prompt author defined these criteria the generated code should satisfy. ")
		b.WriteString("Evaluate EACH criterion individually as pass/fail:\n\n")
		b.WriteString(promptSpecificCriteria)
		b.WriteString("\n\n")
	}

	b.WriteString("## Generated Code\n\n")
	for name, content := range generatedFiles {
		fmt.Fprintf(&b, "### %s\n```\n%s\n```\n\n", name, content)
	}

	if len(referenceFiles) > 0 {
		b.WriteString("## Reference Answer\n\n")
		for name, content := range referenceFiles {
			fmt.Fprintf(&b, "### %s\n```\n%s\n```\n\n", name, content)
		}
	} else {
		b.WriteString("## Reference Answer\n\nNo reference answer provided.\n\n")
	}

	b.WriteString(embeddedRubric)
	return b.String()
}
