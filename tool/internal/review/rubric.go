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
// and optional prompt-specific evaluation criteria.
func BuildReviewPrompt(originalPrompt string, generatedFiles map[string]string, referenceFiles map[string]string, evaluationCriteria string) string {
	var b strings.Builder

	b.WriteString("You are evaluating another AI agent's work. The agent was given the prompt below ")
	b.WriteString("and asked to produce code. Review the generated code against the original prompt, ")
	b.WriteString("the general scoring rubric, and any prompt-specific evaluation criteria.\n\n")

	b.WriteString("## Original Prompt\n\n")
	b.WriteString(originalPrompt)
	b.WriteString("\n\n")

	if evaluationCriteria != "" {
		b.WriteString("## Prompt-Specific Evaluation Criteria\n\n")
		b.WriteString("The prompt author defined these criteria the generated code should satisfy:\n\n")
		b.WriteString(evaluationCriteria)
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

	b.WriteString(scoringRubric(len(referenceFiles) > 0))
	return b.String()
}

func scoringRubric(hasReference bool) string {
	refInstruction := "How similar is it to the reference? (1-10)"
	if !hasReference {
		refInstruction = "Skip (no reference provided), output 0"
	}
	return strings.ReplaceAll(embeddedRubric, "{{REFERENCE_INSTRUCTION}}", refInstruction)
}
