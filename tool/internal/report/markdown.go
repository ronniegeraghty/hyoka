package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteMarkdownReport writes an individual evaluation report as Markdown.
func WriteMarkdownReport(r *EvalReport, outputDir string, runID string, service, plane, language, category string) (string, error) {
	reportDir := filepath.Join(
		outputDir, runID, "results",
		service, plane, language, category, r.ConfigName,
	)
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return "", fmt.Errorf("creating markdown report directory: %w", err)
	}

	reportPath := filepath.Join(reportDir, "report.md")

	d := buildReportData(r)

	var b strings.Builder

	// Header
	result := "❌ FAILED"
	if r.Success {
		result = "✅ PASSED"
	}
	fmt.Fprintf(&b, "# Evaluation Report: %s\n\n", r.PromptID)
	fmt.Fprintf(&b, "**Config:** %s | **Result:** %s | **Duration:** %.1fs\n\n", r.ConfigName, result, r.Duration)

	// Overview table
	b.WriteString("## Overview\n\n")
	b.WriteString("| Field | Value |\n")
	b.WriteString("|-------|-------|\n")
	fmt.Fprintf(&b, "| Prompt ID | `%s` |\n", r.PromptID)
	fmt.Fprintf(&b, "| Config | %s |\n", r.ConfigName)
	fmt.Fprintf(&b, "| Result | %s |\n", result)
	if r.Review != nil {
		fmt.Fprintf(&b, "| Score | %d/10 |\n", r.Review.OverallScore)
	}
	fmt.Fprintf(&b, "| Duration | %.1fs |\n", r.Duration)
	fmt.Fprintf(&b, "| Timestamp | %s |\n", r.Timestamp)
	fmt.Fprintf(&b, "| Files Generated | %d |\n", len(r.GeneratedFiles))
	fmt.Fprintf(&b, "| Event Count | %d |\n", r.EventCount)
	if r.IsStub {
		b.WriteString("| Mode | ⚠️ Stub (no real Copilot) |\n")
	}
	b.WriteString("\n")

	// Config used
	if len(r.ConfigUsed) > 0 {
		b.WriteString("## Configuration\n\n")
		for k, v := range r.ConfigUsed {
			fmt.Fprintf(&b, "- **%s:** %v\n", k, v)
		}
		b.WriteString("\n")
	}

	// Error (if failed)
	if r.Error != "" {
		b.WriteString("## Error\n\n")
		fmt.Fprintf(&b, "```\n%s\n```\n\n", r.Error)
		if r.ErrorDetails != "" {
			b.WriteString("**Details:**\n\n")
			fmt.Fprintf(&b, "```\n%s\n```\n\n", r.ErrorDetails)
		}
	}

	// Prompt
	if d.Prompt != "" {
		b.WriteString("## Prompt Sent\n\n")
		fmt.Fprintf(&b, "```\n%s\n```\n\n", d.Prompt)
	}

	// Reasoning
	if d.Reasoning != "" {
		b.WriteString("## Copilot Reasoning\n\n")
		b.WriteString(d.Reasoning)
		b.WriteString("\n\n")
	}

	// Tool calls
	if len(d.ToolActions) > 0 {
		b.WriteString("## Tool Calls\n\n")
		for _, ta := range d.ToolActions {
			statusIcon := "🔧"
			if ta.Success != nil {
				if *ta.Success {
					statusIcon = "✅"
				} else {
					statusIcon = "❌"
				}
			}
			fmt.Fprintf(&b, "### %s %d. %s", statusIcon, ta.Index, ta.ToolName)
			if ta.MCPServer != "" {
				fmt.Fprintf(&b, " (via %s)", ta.MCPServer)
			}
			if ta.Duration > 0 {
				fmt.Fprintf(&b, " — %.0fms", ta.Duration)
			}
			b.WriteString("\n\n")

			if ta.Args != "" {
				b.WriteString("**Input:**\n\n")
				fmt.Fprintf(&b, "```json\n%s\n```\n\n", ta.Args)
			}
			if ta.Result != "" {
				b.WriteString("**Output:**\n\n")
				fmt.Fprintf(&b, "```\n%s\n```\n\n", truncateStr(ta.Result, 2000))
			}
			if ta.Error != "" {
				b.WriteString("**Error:**\n\n")
				fmt.Fprintf(&b, "```\n%s\n```\n\n", ta.Error)
			}
		}
	}

	// Generated files
	if len(r.GeneratedFiles) > 0 {
		b.WriteString("## Generated Files\n\n")
		for _, f := range r.GeneratedFiles {
			fmt.Fprintf(&b, "- `%s`\n", f)
		}
		b.WriteString("\n")
	}

	// Reviewed (annotated) files
	if len(r.ReviewedFiles) > 0 {
		b.WriteString("## Reviewed Code (Annotated)\n\n")
		b.WriteString("Files with inline `REVIEW:` comments. Annotated files saved in `reviewed-code/`.\n\n")
		for _, rf := range r.ReviewedFiles {
			fmt.Fprintf(&b, "### %s\n\n", rf.Path)
			lang := langFromPath(rf.Path)
			fmt.Fprintf(&b, "```%s\n%s\n```\n\n", lang, rf.Content)
		}
	}

	// Final reply
	if d.FinalReply != "" {
		b.WriteString("## Copilot Response\n\n")
		b.WriteString(d.FinalReply)
		b.WriteString("\n\n")
	}

	// Verification
	if r.Verification != nil {
		b.WriteString("## Verification\n\n")
		verResult := "❌ FAIL"
		if r.Verification.Pass {
			verResult = "✅ PASS"
		}
		fmt.Fprintf(&b, "**Result:** %s\n\n", verResult)
		if r.Verification.Summary != "" {
			fmt.Fprintf(&b, "**Summary:** %s\n\n", r.Verification.Summary)
		}
		if r.Verification.Reasoning != "" {
			b.WriteString("**Reasoning:**\n\n")
			b.WriteString(r.Verification.Reasoning)
			b.WriteString("\n\n")
		}
	}

	// Tool usage evaluation
	if r.ToolUsage != nil {
		b.WriteString("## Tool Usage Evaluation\n\n")
		matchResult := "❌ MISMATCH"
		if r.ToolUsage.Match {
			matchResult = "✅ MATCH"
		}
		fmt.Fprintf(&b, "**Result:** %s\n\n", matchResult)

		b.WriteString("| Category | Tools |\n")
		b.WriteString("|----------|-------|\n")
		fmt.Fprintf(&b, "| Expected | %s |\n", strings.Join(r.ToolUsage.ExpectedTools, ", "))
		fmt.Fprintf(&b, "| Actual | %s |\n", strings.Join(r.ToolUsage.ActualTools, ", "))
		if len(r.ToolUsage.MatchedTools) > 0 {
			fmt.Fprintf(&b, "| Matched | %s |\n", strings.Join(r.ToolUsage.MatchedTools, ", "))
		}
		if len(r.ToolUsage.MissingTools) > 0 {
			fmt.Fprintf(&b, "| ⚠️ Missing | %s |\n", strings.Join(r.ToolUsage.MissingTools, ", "))
		}
		if len(r.ToolUsage.ExtraTools) > 0 {
			fmt.Fprintf(&b, "| Extra | %s |\n", strings.Join(r.ToolUsage.ExtraTools, ", "))
		}
		b.WriteString("\n")
	}

	// Build
	if r.Build != nil {
		b.WriteString("## Build Verification\n\n")
		buildResult := "❌ FAIL"
		if r.Build.Success {
			buildResult = "✅ PASS"
		}
		fmt.Fprintf(&b, "**Result:** %s | **Language:** %s | **Command:** `%s` | **Exit Code:** %d\n\n",
			buildResult, r.Build.Language, r.Build.Command, r.Build.ExitCode)
		if r.Build.Stdout != "" || r.Build.Stderr != "" {
			output := r.Build.Stdout
			if r.Build.Stderr != "" {
				if output != "" {
					output += "\n"
				}
				output += r.Build.Stderr
			}
			b.WriteString("<details>\n<summary>Build Output</summary>\n\n")
			fmt.Fprintf(&b, "```\n%s\n```\n\n", truncateStr(output, 5000))
			b.WriteString("</details>\n\n")
		}
	}

	// Review scores
	if r.Review != nil {
		b.WriteString("## Code Review (LLM-as-Judge)\n\n")
		fmt.Fprintf(&b, "**Overall Score: %d/10**\n\n", r.Review.OverallScore)

		b.WriteString("### Dimension Scores\n\n")
		b.WriteString("| Dimension | Score |\n")
		b.WriteString("|-----------|-------|\n")
		fmt.Fprintf(&b, "| Correctness | %d/10 |\n", r.Review.Scores.Correctness)
		fmt.Fprintf(&b, "| Completeness | %d/10 |\n", r.Review.Scores.Completeness)
		fmt.Fprintf(&b, "| Best Practices | %d/10 |\n", r.Review.Scores.BestPractices)
		fmt.Fprintf(&b, "| Error Handling | %d/10 |\n", r.Review.Scores.ErrorHandling)
		fmt.Fprintf(&b, "| Package Usage | %d/10 |\n", r.Review.Scores.PackageUsage)
		fmt.Fprintf(&b, "| Code Quality | %d/10 |\n", r.Review.Scores.CodeQuality)
		if r.Review.Scores.ReferenceSimilarity > 0 {
			fmt.Fprintf(&b, "| Reference Similarity | %d/10 |\n", r.Review.Scores.ReferenceSimilarity)
		}
		b.WriteString("\n")

		if r.Review.Summary != "" {
			b.WriteString("### Summary\n\n")
			b.WriteString(r.Review.Summary)
			b.WriteString("\n\n")
		}

		if len(r.Review.Strengths) > 0 {
			b.WriteString("### Strengths\n\n")
			for _, s := range r.Review.Strengths {
				fmt.Fprintf(&b, "- %s\n", s)
			}
			b.WriteString("\n")
		}

		if len(r.Review.Issues) > 0 {
			b.WriteString("### Issues\n\n")
			for _, s := range r.Review.Issues {
				fmt.Fprintf(&b, "- %s\n", s)
			}
			b.WriteString("\n")
		}
	}

	// Footer
	b.WriteString("---\n\n")
	b.WriteString("[← Back to Summary](../../../../../../summary.md)\n")

	if err := os.WriteFile(reportPath, []byte(b.String()), 0644); err != nil {
		return "", fmt.Errorf("writing markdown report: %w", err)
	}

	return reportPath, nil
}

// WriteSummaryMarkdown writes a cross-config comparison summary as Markdown.
func WriteSummaryMarkdown(s *RunSummary, outputDir string) (string, error) {
	summaryDir := filepath.Join(outputDir, s.RunID)
	if err := os.MkdirAll(summaryDir, 0755); err != nil {
		return "", fmt.Errorf("creating summary directory: %w", err)
	}

	summaryPath := filepath.Join(summaryDir, "summary.md")

	matrix := buildMatrix(s)

	var b strings.Builder

	// Header
	fmt.Fprintf(&b, "# Evaluation Summary: %s\n\n", s.RunID)

	// Stats table
	b.WriteString("## Run Statistics\n\n")
	b.WriteString("| Metric | Value |\n")
	b.WriteString("|--------|-------|\n")
	fmt.Fprintf(&b, "| Run ID | `%s` |\n", s.RunID)
	fmt.Fprintf(&b, "| Timestamp | %s |\n", s.Timestamp)
	fmt.Fprintf(&b, "| Total Prompts | %d |\n", s.TotalPrompts)
	fmt.Fprintf(&b, "| Total Configs | %d |\n", s.TotalConfigs)
	fmt.Fprintf(&b, "| Total Evaluations | %d |\n", s.TotalEvals)
	fmt.Fprintf(&b, "| Passed | %d |\n", s.Passed)
	fmt.Fprintf(&b, "| Failed | %d |\n", s.Failed)
	fmt.Fprintf(&b, "| Errors | %d |\n", s.Errors)
	fmt.Fprintf(&b, "| Duration | %.1fs |\n", s.Duration)
	b.WriteString("\n")

	// Cross-config comparison matrix
	if len(matrix.Configs) > 0 && len(matrix.Prompts) > 0 {
		b.WriteString("## Comparison Matrix\n\n")

		// Header row
		b.WriteString("| Prompt |")
		for _, cfg := range matrix.Configs {
			fmt.Fprintf(&b, " %s |", cfg)
		}
		b.WriteString("\n")

		// Separator
		b.WriteString("|--------|")
		for range matrix.Configs {
			b.WriteString("--------|")
		}
		b.WriteString("\n")

		// Data rows
		for _, pid := range matrix.Prompts {
			fmt.Fprintf(&b, "| %s |", pid)
			for _, cfg := range matrix.Configs {
				cell := matrix.Cells[pid][cfg]
				if cell == nil {
					b.WriteString(" — |")
					continue
				}
				icon := "❌"
				if cell.Success {
					icon = "✅"
				}
				if cell.HasReview {
					fmt.Fprintf(&b, " %s %d/10 |", icon, cell.Score)
				} else if cell.Error != "" {
					fmt.Fprintf(&b, " ⚠️ Error |")
				} else {
					fmt.Fprintf(&b, " %s |", icon)
				}
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Detailed results
	b.WriteString("## Detailed Results\n\n")
	b.WriteString("| Prompt | Config | Result | Score | Duration | Files |\n")
	b.WriteString("|--------|--------|--------|-------|----------|-------|\n")
	for _, r := range s.Results {
		icon := "❌"
		if r.Success {
			icon = "✅"
		}
		score := "—"
		if r.Review != nil {
			score = fmt.Sprintf("%d/10", r.Review.OverallScore)
		}
		// Build relative link to individual report
		service, _ := r.PromptMeta["service"].(string)
		plane, _ := r.PromptMeta["plane"].(string)
		language, _ := r.PromptMeta["language"].(string)
		category, _ := r.PromptMeta["category"].(string)
		promptCell := r.PromptID
		if service != "" && plane != "" && language != "" && category != "" {
			link := filepath.Join("results", service, plane, language, category, r.ConfigName, "report.md")
			promptCell = fmt.Sprintf("[%s](%s)", r.PromptID, link)
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %.1fs | %d |\n",
			promptCell, r.ConfigName, icon, score, r.Duration, len(r.GeneratedFiles))
	}
	b.WriteString("\n")

	if err := os.WriteFile(summaryPath, []byte(b.String()), 0644); err != nil {
		return "", fmt.Errorf("writing markdown summary: %w", err)
	}

	return summaryPath, nil
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "\n... (truncated)"
}

// langFromPath returns a markdown code fence language hint from a file path.
func langFromPath(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".py":
		return "python"
	case ".go":
		return "go"
	case ".cs":
		return "csharp"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".java":
		return "java"
	case ".rs":
		return "rust"
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	case ".sh":
		return "bash"
	default:
		return ""
	}
}
