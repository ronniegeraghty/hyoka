package report

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

// Parsed once at package init for reuse across calls.
var (
	parsedReportTemplate  = template.Must(template.New("report").Funcs(htmlFuncMap()).Parse(reportTemplate))
	parsedSummaryTemplate = template.Must(template.New("summary").Funcs(htmlFuncMap()).Parse(summaryTemplate))
)

// WriteHTMLReport writes an individual evaluation report as HTML.
func WriteHTMLReport(r *EvalReport, outputDir string, runID string, service, plane, language, category string) (string, error) {
	reportDir := filepath.Join(
		outputDir, runID, "results",
		service, plane, language, category, r.ConfigName,
	)
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return "", fmt.Errorf("creating HTML report directory: %w", err)
	}

	reportPath := filepath.Join(reportDir, "report.html")

	f, err := os.Create(reportPath)
	if err != nil {
		return "", fmt.Errorf("creating HTML report file: %w", err)
	}
	defer f.Close()

	data := buildReportData(r)

	// Read file contents from the generated-code directory for expandable display (Issue 3)
	codeDir := filepath.Join(reportDir, "generated-code")
	data.FileContents = readFileContents(codeDir, r.GeneratedFiles, r.StarterFiles)

	if err := parsedReportTemplate.Execute(f, data); err != nil {
		return "", fmt.Errorf("executing report template: %w", err)
	}

	return reportPath, nil
}

// WriteSummaryHTML writes a cross-config comparison summary as HTML.
func WriteSummaryHTML(s *RunSummary, outputDir string) (string, error) {
	summaryDir := filepath.Join(outputDir, s.RunID)
	if err := os.MkdirAll(summaryDir, 0755); err != nil {
		return "", fmt.Errorf("creating summary directory: %w", err)
	}

	summaryPath := filepath.Join(summaryDir, "summary.html")

	f, err := os.Create(summaryPath)
	if err != nil {
		return "", fmt.Errorf("creating summary HTML file: %w", err)
	}
	defer f.Close()

	matrix := buildMatrix(s)
	stats := ComputeSummaryStats(s)

	data := struct {
		Summary *RunSummary
		Matrix  *MatrixData
		Stats   *SummaryStats
	}{
		Summary: s,
		Matrix:  matrix,
		Stats:   stats,
	}

	if err := parsedSummaryTemplate.Execute(f, data); err != nil {
		return "", fmt.Errorf("executing summary template: %w", err)
	}

	return summaryPath, nil
}

// MatrixData holds the cross-config comparison matrix.
type MatrixData struct {
	Configs []string
	Prompts []string
	Cells   map[string]map[string]*MatrixCell // [promptID][configName]
}

// MatrixCell holds the data for one cell in the matrix.
type MatrixCell struct {
	Success    bool
	Score      int
	BuildPass  bool
	HasReview  bool
	Duration   float64
	Error      string
	FileCount  int
	ToolCalls  []string
	ReportLink string
}

func buildMatrix(s *RunSummary) *MatrixData {
	m := &MatrixData{
		Cells: make(map[string]map[string]*MatrixCell),
	}

	configSet := make(map[string]bool)
	promptSet := make(map[string]bool)

	for _, r := range s.Results {
		if !promptSet[r.PromptID] {
			promptSet[r.PromptID] = true
			m.Prompts = append(m.Prompts, r.PromptID)
		}
		if !configSet[r.ConfigName] {
			configSet[r.ConfigName] = true
			m.Configs = append(m.Configs, r.ConfigName)
		}

		if m.Cells[r.PromptID] == nil {
			m.Cells[r.PromptID] = make(map[string]*MatrixCell)
		}

		cell := &MatrixCell{
			Success:   r.Success,
			Duration:  r.Duration,
			Error:     r.Error,
			FileCount: len(r.GeneratedFiles),
			ToolCalls: r.ToolCalls,
		}
		if r.Build != nil {
			cell.BuildPass = r.Build.Success
		}
		if r.Review != nil {
			cell.Score = r.Review.OverallScore
			cell.HasReview = true
		}
		// Build relative link from summary.html to individual report
		service, _ := r.PromptMeta["service"].(string)
		plane, _ := r.PromptMeta["plane"].(string)
		language, _ := r.PromptMeta["language"].(string)
		category, _ := r.PromptMeta["category"].(string)
		if service != "" && plane != "" && language != "" && category != "" {
			cell.ReportLink = filepath.Join("results", service, plane, language, category, r.ConfigName, "report.html")
		}
		m.Cells[r.PromptID][r.ConfigName] = cell
	}

	return m
}

// ReportTemplateData is the enriched data passed to the individual report template.
type ReportTemplateData struct {
	*EvalReport
	Prompt        string
	Reasoning     string
	FinalReply    string
	ToolActions   []ToolAction
	TimelineSteps []TimelineStep
	FileCount     int
	FileContents  map[string]string // filename → content for expandable display
}

// ToolAction represents one tool invocation extracted from session events.
type ToolAction struct {
	Index     int
	ToolName  string
	Args      string
	Result    string
	Error     string
	Success   *bool
	Duration  float64
	MCPServer string
}

// TimelineStep represents one chronological step in the agent workflow.
type TimelineStep struct {
	Index     int
	Phase     string // "generation", "verification", "review"
	StepType  string // "prompt", "reasoning", "tool_call", "message", "complete"
	Icon      string
	Title     string
	Content   string  // main content (tool result, reasoning text)
	Detail    string  // collapsible detail (tool args, full text)
	Duration  float64 // milliseconds
	Success   *bool
	Error     string
	ToolName  string
	MCPServer string
}

// buildReportData extracts structured sections from session events and
// builds a chronological timeline of agent steps.
func buildReportData(r *EvalReport) *ReportTemplateData {
	d := &ReportTemplateData{
		EvalReport: r,
		FileCount:  len(r.GeneratedFiles),
	}

	var reasoningParts []string
	var messageParts []string
	stepIndex := 0

	type pendingTool struct {
		stepIdx int
		name    string
	}
	var pendingTools []pendingTool

	for _, ev := range r.SessionEvents {
		switch ev.Type {
		case "user.message":
			if d.Prompt == "" && ev.Content != "" {
				d.Prompt = ev.Content
			}
			if ev.Content != "" {
				stepIndex++
				d.TimelineSteps = append(d.TimelineSteps, TimelineStep{
					Index:    stepIndex,
					Phase:    "generation",
					StepType: "prompt",
					Icon:     "📝",
					Title:    "Prompt sent",
					Content:  ev.Content,
				})
			}
		case "assistant.reasoning":
			if ev.Content != "" {
				reasoningParts = append(reasoningParts, ev.Content)
				stepIndex++
				title := ev.Content
				if len(title) > 80 {
					title = title[:80] + "…"
				}
				d.TimelineSteps = append(d.TimelineSteps, TimelineStep{
					Index:    stepIndex,
					Phase:    "generation",
					StepType: "reasoning",
					Icon:     "🤔",
					Title:    title,
					Content:  ev.Content,
				})
			}
		case "tool.execution_start":
			if ev.ToolName != "" {
				d.ToolActions = append(d.ToolActions, ToolAction{
					Index:     len(d.ToolActions) + 1,
					ToolName:  ev.ToolName,
					Args:      ev.ToolArgs,
					MCPServer: ev.MCPServerName,
				})
				stepIndex++
				toolTitle := ev.ToolName
				if ev.FilePath != "" {
					toolTitle += " → " + ev.FilePath
				}
				step := TimelineStep{
					Index:     stepIndex,
					Phase:     "generation",
					StepType:  "tool_call",
					Icon:      "🔧",
					Title:     "Tool call: " + toolTitle,
					Detail:    ev.ToolArgs,
					ToolName:  ev.ToolName,
					MCPServer: ev.MCPServerName,
				}
				d.TimelineSteps = append(d.TimelineSteps, step)
				pendingTools = append(pendingTools, pendingTool{len(d.TimelineSteps) - 1, ev.ToolName})
			}
		case "tool.execution_complete":
			// Update ToolActions (backward compat)
			for i := len(d.ToolActions) - 1; i >= 0; i-- {
				if d.ToolActions[i].ToolName == ev.ToolName && d.ToolActions[i].Result == "" && d.ToolActions[i].Error == "" {
					d.ToolActions[i].Result = ev.ToolResult
					d.ToolActions[i].Error = ev.Error
					d.ToolActions[i].Success = ev.ToolSuccess
					d.ToolActions[i].Duration = ev.Duration
					break
				}
			}
			// Update matching timeline step
			for i := len(pendingTools) - 1; i >= 0; i-- {
				if pendingTools[i].name == ev.ToolName {
					idx := pendingTools[i].stepIdx
					d.TimelineSteps[idx].Content = ev.ToolResult
					d.TimelineSteps[idx].Duration = ev.Duration
					d.TimelineSteps[idx].Success = ev.ToolSuccess
					if ev.Error != "" {
						d.TimelineSteps[idx].Error = ev.Error
					}
					pendingTools = append(pendingTools[:i], pendingTools[i+1:]...)
					break
				}
			}
		case "assistant.message":
			if ev.Content != "" {
				messageParts = append(messageParts, ev.Content)
				stepIndex++
				d.TimelineSteps = append(d.TimelineSteps, TimelineStep{
					Index:    stepIndex,
					Phase:    "generation",
					StepType: "message",
					Icon:     "💬",
					Title:    "Agent reply",
					Content:  ev.Content,
				})
			}
		}
	}

	// Add generation-complete step
	if len(d.TimelineSteps) > 0 {
		stepIndex++
		summary := fmt.Sprintf("%d files created", d.FileCount)
		d.TimelineSteps = append(d.TimelineSteps, TimelineStep{
			Index:    stepIndex,
			Phase:    "generation",
			StepType: "complete",
			Icon:     "✅",
			Title:    "Generation complete",
			Content:  summary,
		})
	}

	d.Reasoning = strings.Join(reasoningParts, "\n\n")
	d.FinalReply = strings.Join(messageParts, "\n\n")
	return d
}

// readFileContents reads file contents from the code directory for display in the HTML report.
// If starterFiles is non-empty, only files NOT in the starter set are included.
func readFileContents(codeDir string, files []string, starterFiles []string) map[string]string {
	contents := make(map[string]string)
	starterSet := make(map[string]bool, len(starterFiles))
	for _, f := range starterFiles {
		starterSet[f] = true
	}
	for _, f := range files {
		if len(starterFiles) > 0 && starterSet[f] {
			continue // skip unchanged starter project files
		}
		path := filepath.Join(codeDir, f)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if len(data) > 512*1024 {
			contents[f] = "(file too large to display)"
			continue
		}
		contents[f] = string(data)
	}
	return contents
}

func htmlFuncMap() template.FuncMap {
	return template.FuncMap{
		"scoreColor": func(score int) string {
			switch {
			case score >= 8:
				return "#22c55e" // green
			case score >= 6:
				return "#eab308" // yellow
			case score >= 4:
				return "#f97316" // orange
			default:
				return "#ef4444" // red
			}
		},
		"statusIcon": func(success bool) string {
			if success {
				return "✅"
			}
			return "❌"
		},
		"join": func(items []string, sep string) string {
			return strings.Join(items, sep)
		},
		"fmtDuration": func(d float64) string {
			return fmt.Sprintf("%.1fs", d)
		},
		"truncate": func(s string, n int) string {
			if len(s) <= n {
				return s
			}
			return s[:n] + "…"
		},
		"isReviewLine": func(line string) bool {
			trimmed := strings.TrimSpace(line)
			return strings.Contains(trimmed, "REVIEW:")
		},
		"highlightReviewLines": func(content string) template.HTML {
			lines := strings.Split(content, "\n")
			var b strings.Builder
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.Contains(trimmed, "REVIEW:") {
					b.WriteString(`<span class="review-comment">`)
					b.WriteString(template.HTMLEscapeString(line))
					b.WriteString("</span>\n")
				} else {
					b.WriteString(template.HTMLEscapeString(line))
					b.WriteString("\n")
				}
			}
			return template.HTML(b.String())
		},
		"langClass": func(filename string) string {
			ext := filepath.Ext(filename)
			switch ext {
			case ".py":
				return "python"
			case ".cs":
				return "csharp"
			case ".go":
				return "go"
			case ".js":
				return "javascript"
			case ".ts":
				return "typescript"
			case ".java":
				return "java"
			case ".json":
				return "json"
			case ".yaml", ".yml":
				return "yaml"
			case ".xml":
				return "xml"
			case ".md":
				return "markdown"
			case ".sh":
				return "bash"
			case ".ps1":
				return "powershell"
			default:
				return ""
			}
		},
		"hasPrefix": strings.HasPrefix,
		"contains":  strings.Contains,
		"boolStr": func(b *bool) string {
			if b == nil {
				return ""
			}
			if *b {
				return "✅"
			}
			return "❌"
		},
		"reportLink": func(r *EvalReport) string {
			service, _ := r.PromptMeta["service"].(string)
			plane, _ := r.PromptMeta["plane"].(string)
			language, _ := r.PromptMeta["language"].(string)
			category, _ := r.PromptMeta["category"].(string)
			if service == "" || plane == "" || language == "" || category == "" {
				return ""
			}
			return filepath.Join("results", service, plane, language, category, r.ConfigName, "report.html")
		},
	}
}

const reportTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Eval Report: {{.PromptID}} / {{.ConfigName}}</title>
<style>
  :root {
    --green: #22c55e; --red: #ef4444; --yellow: #eab308; --orange: #f97316;
    --gray: #6b7280; --bg: #f8fafc; --card-bg: #fff; --border: #e2e8f0;
    --text: #0f172a; --text-muted: #64748b; --purple: #7c3aed; --blue: #2563eb;
    --indigo: #4f46e5;
  }
  * { box-sizing: border-box; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 1000px; margin: 0 auto; padding: 2rem 1rem; color: var(--text); background: var(--bg); line-height: 1.6; }

  /* Navigation */
  .nav { margin-bottom: 1.5rem; }
  .nav a { color: var(--blue); text-decoration: none; font-size: 0.9rem; }
  .nav a:hover { text-decoration: underline; }

  /* Header */
  .report-header { margin-bottom: 2rem; }
  .report-header h1 { font-size: 1.5rem; margin: 0 0 0.25rem 0; }
  .report-header .subtitle { color: var(--text-muted); font-size: 0.95rem; }

  /* Result banner */
  .result-banner { display: flex; align-items: center; gap: 1.5rem; flex-wrap: wrap; padding: 1rem 1.25rem; border-radius: 8px; margin-bottom: 1.5rem; font-size: 0.95rem; }
  .result-banner.pass { background: #f0fdf4; border: 1px solid #bbf7d0; }
  .result-banner.fail { background: #fef2f2; border: 1px solid #fecaca; }
  .result-banner .verdict { font-size: 1.25rem; font-weight: 700; }
  .result-banner .meta-item { color: var(--text-muted); }
  .result-banner .meta-item strong { color: var(--text); }

  /* Badges */
  .badge { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 0.8em; font-weight: 600; }
  .badge-pass { background: #dcfce7; color: #166534; }
  .badge-fail { background: #fef2f2; color: #991b1b; }
  .badge-stub { background: #fef3c7; color: #92400e; }

  /* Sections */
  .section { background: var(--card-bg); border: 1px solid var(--border); border-radius: 8px; margin-bottom: 1.25rem; overflow: hidden; }
  .section-header { display: flex; align-items: center; gap: 0.5rem; padding: 0.75rem 1rem; border-bottom: 1px solid var(--border); background: #f8fafc; }
  .section-header h2 { font-size: 1rem; margin: 0; font-weight: 600; }
  .section-header .icon { font-size: 1.1rem; }
  .section-body { padding: 1rem; }

  /* Collapsible content */
  details { margin: 0.5rem 0; }
  details > summary { cursor: pointer; font-weight: 600; padding: 0.4rem 0; color: var(--text); user-select: none; }
  details > summary:hover { color: var(--blue); }
  details[open] > summary { margin-bottom: 0.5rem; }

  /* Code / pre */
  pre { background: #1e293b; color: #e2e8f0; padding: 1rem; border-radius: 6px; overflow-x: auto; font-size: 0.85rem; line-height: 1.5; margin: 0.5rem 0; white-space: pre-wrap; word-break: break-word; }
  code { font-family: 'SF Mono', 'Fira Code', Consolas, monospace; font-size: 0.85em; }
  p code, li code { background: #f1f5f9; color: var(--indigo); padding: 1px 5px; border-radius: 3px; }

  /* File cards */
  .file-card { background: var(--card-bg); border: 1px solid var(--border); border-radius: 8px; margin-bottom: 0.75rem; overflow: hidden; }
  .file-card-header { display: flex; align-items: center; gap: 0.5rem; padding: 0.5rem 1rem; background: #1e293b; color: #e2e8f0; font-family: monospace; font-size: 0.85rem; }
  .file-card-header .file-icon { opacity: 0.7; }
  .file-card pre { margin: 0; border-radius: 0; }
  .file-card details { margin: 0; }
  .file-card details > summary { padding: 0.4rem 1rem; font-size: 0.85rem; font-weight: 500; color: var(--text-muted); }

  /* Scores grid */
  .scores-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(130px, 1fr)); gap: 0.75rem; margin: 0.75rem 0; }
  .score-card { text-align: center; padding: 0.75rem 0.5rem; border-radius: 6px; background: #f8fafc; border: 1px solid var(--border); }
  .score-card .value { font-size: 1.4rem; font-weight: 700; }
  .score-card .label { font-size: 0.75rem; color: var(--text-muted); margin-top: 0.15rem; }
  .overall-score { text-align: center; margin: 1rem 0; }
  .overall-score .value { font-size: 2.5rem; font-weight: 700; }
  .overall-score .label { font-size: 0.85rem; color: var(--text-muted); }

  /* Strengths / issues */
  .findings-list { padding-left: 1.25rem; margin: 0.5rem 0; }
  .findings-list li { padding: 0.2rem 0; font-size: 0.9rem; }

  /* Error box */
  .error-box { background: #fef2f2; border: 1px solid #fecaca; border-radius: 8px; padding: 1rem; margin-bottom: 1.25rem; }
  .error-box h3 { margin: 0 0 0.5rem 0; color: #991b1b; }

  /* Meta table */
  .meta-table { width: 100%; border-collapse: collapse; }
  .meta-table td { padding: 0.3rem 0.5rem; border-bottom: 1px solid #f1f5f9; font-size: 0.9rem; }
  .meta-table td:first-child { font-weight: 600; width: 120px; color: var(--text-muted); }

  /* Verification */
  .verify-banner { padding: 1rem; border-radius: 6px; margin-bottom: 0.75rem; }
  .verify-pass { background: #f0fdf4; border: 1px solid #bbf7d0; }
  .verify-fail { background: #fef2f2; border: 1px solid #fecaca; }

  /* Review comment highlighting */
  .review-comment { background: #fef3c7; color: #92400e; display: block; border-left: 3px solid #f59e0b; padding-left: 0.5rem; margin: 1px 0; }
  .reviewed-file pre { white-space: pre-wrap; word-break: break-word; }

  /* Tags */
  .tag { display: inline-block; background: #f3f0ff; color: var(--purple); padding: 1px 8px; border-radius: 12px; font-size: 0.78rem; margin: 1px 2px; }

  /* ━━ Timeline ━━ */
  .phase { margin-bottom: 1.5rem; }
  .phase-header { display: flex; align-items: center; gap: 0.5rem; padding: 0.75rem 1rem; font-weight: 700; font-size: 1rem; border-radius: 8px 8px 0 0; }
  .phase-gen .phase-header { background: #eff6ff; color: #1e40af; border: 1px solid #bfdbfe; }
  .phase-verify .phase-header { background: #f0fdf4; color: #166534; border: 1px solid #bbf7d0; }
  .phase-review .phase-header { background: #faf5ff; color: #6b21a8; border: 1px solid #d8b4fe; }
  .timeline { position: relative; padding: 1.25rem 1rem 0.5rem 3.5rem; border: 1px solid var(--border); border-top: none; border-radius: 0 0 8px 8px; background: var(--card-bg); }
  .timeline::before { content: ''; position: absolute; left: 1.6rem; top: 0; bottom: 0; width: 2px; }
  .phase-gen .timeline::before { background: #93c5fd; }
  .phase-verify .timeline::before { background: #86efac; }
  .phase-review .timeline::before { background: #c4b5fd; }

  .tl-step { position: relative; margin-bottom: 1.25rem; }
  .tl-step:last-child { margin-bottom: 0.5rem; }
  .tl-marker { position: absolute; left: -2.65rem; top: 0.1rem; width: 1.75rem; height: 1.75rem; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 0.85rem; background: #fff; border: 2px solid var(--border); z-index: 1; }
  .phase-gen .tl-marker { border-color: #93c5fd; }
  .phase-verify .tl-marker { border-color: #86efac; }
  .phase-review .tl-marker { border-color: #c4b5fd; }

  .tl-card { padding: 0.6rem 0.85rem; border-radius: 6px; border: 1px solid var(--border); background: #fafbfc; }
  .tl-card-tool { border-left: 3px solid var(--indigo); }
  .tl-card-reasoning { border-left: 3px solid #93c5fd; }
  .tl-card-prompt { border-left: 3px solid var(--blue); }
  .tl-card-message { border-left: 3px solid #6366f1; }
  .tl-card-complete { border-left: 3px solid var(--green); background: #f0fdf4; }

  .tl-title { font-weight: 600; font-size: 0.9rem; display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap; }
  .tl-title .tool-name { font-family: monospace; color: var(--purple); }
  .tl-title .mcp-server { font-size: 0.75rem; color: var(--text-muted); font-weight: 400; }
  .tl-meta { display: flex; gap: 0.75rem; align-items: center; font-size: 0.8rem; color: var(--text-muted); margin-top: 0.2rem; }
  .tl-card details { margin: 0.4rem 0 0 0; }
  .tl-card details summary { font-size: 0.85rem; font-weight: 500; }
  .tl-card pre { font-size: 0.8rem; margin: 0.25rem 0 0 0; max-height: 300px; overflow-y: auto; }
  .tl-error { color: var(--red); font-size: 0.85rem; margin-top: 0.3rem; }
</style>
</head>
<body>

<div class="nav"><a href="../../../../../../summary.html">← Back to Summary</a></div>

<div class="report-header">
  <h1>📋 {{.PromptID}}</h1>
  <div class="subtitle">Config: <strong>{{.ConfigName}}</strong> · {{.Timestamp}}</div>
</div>

{{if .IsStub}}
<div style="background:#fffbeb;border:1px solid #fde68a;border-radius:8px;padding:1rem;margin-bottom:1.25rem">
  ⚠️ <strong>Stub Mode</strong> — no Copilot session was created. Results are placeholders.
</div>
{{end}}

<!-- Result banner -->
<div class="result-banner {{if .Success}}pass{{else}}fail{{end}}">
  <span class="verdict">{{if .Success}}✅ PASSED{{else}}❌ FAILED{{end}}</span>
  {{if .Review}}<span class="meta-item">Score: <strong style="color:{{scoreColor .Review.OverallScore}}">{{.Review.OverallScore}}/10</strong></span>{{end}}
  <span class="meta-item">Duration: <strong>{{fmtDuration .Duration}}</strong></span>
  <span class="meta-item">Files: <strong>{{.FileCount}}</strong></span>
  <span class="meta-item">Tool Calls: <strong>{{len .ToolActions}}</strong></span>
  {{if .IsStub}}<span class="badge badge-stub">STUB</span>{{end}}
</div>

{{if .Error}}
<div class="error-box">
  <h3>❌ Error</h3>
  <p>{{.Error}}</p>
  {{if .ErrorDetails}}<details><summary>Full error details</summary><pre>{{.ErrorDetails}}</pre></details>{{end}}
</div>
{{end}}

<!-- ━━ Prompt Details (Issue 6) ━━ -->
{{if .PromptMeta}}
<div class="section">
  <div class="section-header"><span class="icon">📋</span><h2>Prompt Details</h2></div>
  <div class="section-body">
    <table class="meta-table">
      {{with index .PromptMeta "description"}}{{if .}}<tr><td>Description</td><td>{{.}}</td></tr>{{end}}{{end}}
      {{with index .PromptMeta "difficulty"}}{{if .}}<tr><td>Difficulty</td><td>{{.}}</td></tr>{{end}}{{end}}
      <tr><td>Service</td><td>{{index .PromptMeta "service"}}</td></tr>
      <tr><td>Plane</td><td>{{index .PromptMeta "plane"}}</td></tr>
      <tr><td>Language</td><td>{{index .PromptMeta "language"}}</td></tr>
      <tr><td>Category</td><td>{{index .PromptMeta "category"}}</td></tr>
      {{with index .PromptMeta "sdk_package"}}{{if .}}<tr><td>SDK Package</td><td><code>{{.}}</code></td></tr>{{end}}{{end}}
      {{with index .PromptMeta "tags"}}{{if .}}<tr><td>Tags</td><td>{{.}}</td></tr>{{end}}{{end}}
    </table>
  </div>
</div>
{{end}}

<!-- ━━ Generation Timeline ━━ -->
{{if .TimelineSteps}}
<div class="phase phase-gen">
  <div class="phase-header"><span>🧪</span> Generation Timeline</div>
  <div class="timeline">
    {{range .TimelineSteps}}{{if eq .Phase "generation"}}
    <div class="tl-step">
      <div class="tl-marker">{{.Icon}}</div>
      <div class="tl-card tl-card-{{.StepType}}">
        <div class="tl-title">
          <span>{{.Index}}.</span>
          {{if eq .StepType "tool_call"}}<span class="tool-name">Tool call: {{.ToolName}}</span>{{if .MCPServer}}<span class="mcp-server">via {{.MCPServer}}</span>{{end}}{{else}}{{.Title}}{{end}}
        </div>
        {{if eq .StepType "tool_call"}}
        <div class="tl-meta">
          {{if .Success}}{{boolStr .Success}}{{end}}
          {{if gt .Duration 0.0}}<span>{{printf "%.0fms" .Duration}}</span>{{end}}
        </div>
        {{end}}
        {{if and (eq .StepType "prompt") .Content}}
        <details open><summary>Show prompt</summary><pre>{{.Content}}</pre></details>
        {{end}}
        {{if and (eq .StepType "reasoning") .Content}}
        <details><summary>Show reasoning</summary><pre>{{.Content}}</pre></details>
        {{end}}
        {{if and (eq .StepType "message") .Content}}
        <details><summary>Show reply</summary><pre>{{.Content}}</pre></details>
        {{end}}
        {{if eq .StepType "tool_call"}}
          {{if .Detail}}
          <details><summary>Show arguments</summary><pre>{{.Detail}}</pre></details>
          {{end}}
          {{if .Content}}
          <details><summary>Show result</summary><pre>{{.Content}}</pre></details>
          {{end}}
        {{end}}
        {{if .Error}}<div class="tl-error">❌ {{.Error}}</div>{{end}}
      </div>
    </div>
    {{end}}{{end}}
  </div>
</div>
{{end}}

<!-- ━━ Verification Timeline ━━ -->
{{if .Verification}}
<div class="phase phase-verify">
  <div class="phase-header"><span>🔍</span> Verification {{if .Verification.Pass}}<span class="badge badge-pass" style="margin-left:auto">PASS</span>{{else}}<span class="badge badge-fail" style="margin-left:auto">FAIL</span>{{end}}</div>
  <div class="timeline">
    <div class="tl-step">
      <div class="tl-marker">{{if .Verification.Pass}}✅{{else}}❌{{end}}</div>
      <div class="tl-card tl-card-complete">
        <div class="tl-title">{{if .Verification.Summary}}{{.Verification.Summary}}{{else}}{{if .Verification.Pass}}Verification passed{{else}}Verification failed{{end}}{{end}}</div>
      </div>
    </div>
    {{if .Verification.Reasoning}}
    <div class="tl-step">
      <div class="tl-marker">🤔</div>
      <div class="tl-card tl-card-reasoning">
        <div class="tl-title">Verifier's Reasoning</div>
        <details open><summary>Show reasoning</summary><pre>{{.Verification.Reasoning}}</pre></details>
      </div>
    </div>
    {{end}}
  </div>
</div>
{{end}}

<!-- ━━ Tool Usage Evaluation ━━ -->
{{if .ToolUsage}}
<div class="section">
  <div class="section-header"><span class="icon">🔧</span><h2>Tool Usage Evaluation</h2><span style="margin-left:auto">{{if .ToolUsage.Match}}<span class="badge badge-pass">MATCH</span>{{else}}<span class="badge badge-fail">MISMATCH</span>{{end}}</span></div>
  <div class="section-body">
    <div class="verify-banner {{if .ToolUsage.Match}}verify-pass{{else}}verify-fail{{end}}">
      {{if .ToolUsage.Match}}✅ All expected tools were used{{else}}⚠️ Some expected tools were not used during generation{{end}}
    </div>
    <table class="meta-table">
      <tr><td>Expected</td><td>{{join .ToolUsage.ExpectedTools ", "}}</td></tr>
      <tr><td>Actual</td><td>{{join .ToolUsage.ActualTools ", "}}</td></tr>
      {{if .ToolUsage.MatchedTools}}<tr><td>Matched</td><td style="color:var(--green)">{{join .ToolUsage.MatchedTools ", "}}</td></tr>{{end}}
      {{if .ToolUsage.MissingTools}}<tr><td>Missing</td><td style="color:var(--red);font-weight:600">{{join .ToolUsage.MissingTools ", "}}</td></tr>{{end}}
      {{if .ToolUsage.ExtraTools}}<tr><td>Extra</td><td style="color:var(--text-muted)">{{join .ToolUsage.ExtraTools ", "}}</td></tr>{{end}}
    </table>
  </div>
</div>
{{end}}

<!-- ━━ Code Review (Issue 1: show review session events) ━━ -->
{{if .Review}}
<div class="phase phase-review">
  <div class="phase-header"><span>📊</span> Code Review <span style="margin-left:auto;font-size:0.85rem">Score: {{.Review.OverallScore}}/10</span></div>
  <div class="timeline">
    <div class="tl-step">
      <div class="tl-marker">📊</div>
      <div class="tl-card">
        <div class="overall-score">
          <div class="value" style="color:{{scoreColor .Review.OverallScore}}">{{.Review.OverallScore}}/10</div>
          <div class="label">Overall Score</div>
        </div>
        <div class="scores-grid">
          <div class="score-card"><div class="value" style="color:{{scoreColor .Review.Scores.Correctness}}">{{.Review.Scores.Correctness}}</div><div class="label">Correctness</div></div>
          <div class="score-card"><div class="value" style="color:{{scoreColor .Review.Scores.Completeness}}">{{.Review.Scores.Completeness}}</div><div class="label">Completeness</div></div>
          <div class="score-card"><div class="value" style="color:{{scoreColor .Review.Scores.BestPractices}}">{{.Review.Scores.BestPractices}}</div><div class="label">Best Practices</div></div>
          <div class="score-card"><div class="value" style="color:{{scoreColor .Review.Scores.ErrorHandling}}">{{.Review.Scores.ErrorHandling}}</div><div class="label">Error Handling</div></div>
          <div class="score-card"><div class="value" style="color:{{scoreColor .Review.Scores.PackageUsage}}">{{.Review.Scores.PackageUsage}}</div><div class="label">Package Usage</div></div>
          <div class="score-card"><div class="value" style="color:{{scoreColor .Review.Scores.CodeQuality}}">{{.Review.Scores.CodeQuality}}</div><div class="label">Code Quality</div></div>
          {{if .Review.Scores.ReferenceSimilarity}}<div class="score-card"><div class="value" style="color:{{scoreColor .Review.Scores.ReferenceSimilarity}}">{{.Review.Scores.ReferenceSimilarity}}</div><div class="label">Ref Similarity</div></div>{{end}}
        </div>
        {{if .Review.Summary}}<p>{{.Review.Summary}}</p>{{end}}
      </div>
    </div>
    {{if .Review.Strengths}}
    <div class="tl-step">
      <div class="tl-marker">💪</div>
      <div class="tl-card">
        <div class="tl-title">Strengths</div>
        <ul class="findings-list">{{range .Review.Strengths}}<li>{{.}}</li>{{end}}</ul>
      </div>
    </div>
    {{end}}
    {{if .Review.Issues}}
    <div class="tl-step">
      <div class="tl-marker">⚠️</div>
      <div class="tl-card">
        <div class="tl-title">Issues</div>
        <ul class="findings-list">{{range .Review.Issues}}<li>{{.}}</li>{{end}}</ul>
      </div>
    </div>
    {{end}}
    {{if .Review.Events}}
    <div class="tl-step">
      <div class="tl-marker">🔍</div>
      <div class="tl-card">
        <div class="tl-title">Review Session Activity</div>
        <details><summary>Show reviewer tool calls and analysis ({{len .Review.Events}} events)</summary>
        <div style="margin-top:0.5rem">
        {{range .Review.Events}}
          {{if eq .Type "tool.execution_complete"}}
          <div style="margin:0.5rem 0;padding:0.5rem;border-left:3px solid var(--purple);background:#faf5ff;border-radius:0 4px 4px 0">
            <div style="font-weight:600;font-size:0.85rem;font-family:monospace;color:var(--purple)">🔧 {{.ToolName}}{{if gt .Duration 0.0}} <span style="font-weight:400;color:var(--text-muted)">({{printf "%.0fms" .Duration}})</span>{{end}}</div>
            {{if .Result}}<pre style="font-size:0.78rem;max-height:200px;overflow-y:auto;margin:0.25rem 0 0 0">{{truncate .Result 2000}}</pre>{{end}}
            {{if .Error}}<div style="color:var(--red);font-size:0.8rem;margin-top:0.25rem">❌ {{.Error}}</div>{{end}}
          </div>
          {{end}}
          {{if eq .Type "assistant.message"}}
          {{if .Content}}
          <div style="margin:0.5rem 0;padding:0.5rem;border-left:3px solid #93c5fd;background:#eff6ff;border-radius:0 4px 4px 0">
            <div style="font-size:0.8rem;color:var(--text-muted)">💬 Reviewer</div>
            <pre style="font-size:0.78rem;max-height:200px;overflow-y:auto;margin:0.25rem 0 0 0">{{truncate .Content 2000}}</pre>
          </div>
          {{end}}
          {{end}}
        {{end}}
        </div>
        </details>
      </div>
    </div>
    {{end}}
  </div>
</div>
{{end}}

<!-- ━━ Generated Files (Issue 3: expandable contents) ━━ -->
{{if .GeneratedFiles}}
<div class="section">
  <div class="section-header"><span class="icon">📁</span><h2>Generated Files ({{.FileCount}})</h2></div>
  <div class="section-body">
    <p style="font-size:0.85rem;color:var(--text-muted)">Files are saved in the <code>generated-code/</code> subdirectory alongside this report.</p>
    {{range .GeneratedFiles}}
    <div class="file-card">
      <div class="file-card-header"><span class="file-icon">📄</span> {{.}}</div>
      {{with index $.FileContents .}}
      <details>
        <summary>Show contents</summary>
        <pre>{{.}}</pre>
      </details>
      {{end}}
    </div>
    {{end}}
  </div>
</div>
{{end}}

<!-- ━━ Reviewed Code (Annotated) ━━ -->
{{if .ReviewedFiles}}
<div class="section">
  <div class="section-header"><span class="icon">📝</span><h2>Reviewed Code ({{len .ReviewedFiles}} files with annotations)</h2></div>
  <div class="section-body">
    <p style="font-size:0.85rem;color:var(--text-muted)">Code with inline <code>REVIEW:</code> comments highlighted. Annotated files saved in <code>reviewed-code/</code>.</p>
    {{range .ReviewedFiles}}
    <div class="file-card reviewed-file">
      <div class="file-card-header"><span class="file-icon">📝</span> {{.Path}}</div>
      <pre>{{highlightReviewLines .Content}}</pre>
    </div>
    {{end}}
  </div>
</div>
{{end}}

<!-- ━━ Build Verification (optional) ━━ -->
{{if .Build}}
<div class="section">
  <div class="section-header"><span class="icon">🔨</span><h2>Build Verification</h2><span style="margin-left:auto">{{if .Build.Success}}<span class="badge badge-pass">PASS</span>{{else}}<span class="badge badge-fail">FAIL</span>{{end}}</span></div>
  <div class="section-body">
    <table class="meta-table">
      <tr><td>Language</td><td>{{.Build.Language}}</td></tr>
      <tr><td>Command</td><td><code>{{.Build.Command}}</code></td></tr>
      <tr><td>Exit Code</td><td>{{.Build.ExitCode}}</td></tr>
    </table>
    {{if .Build.Stdout}}<details><summary>Stdout</summary><pre>{{.Build.Stdout}}</pre></details>{{end}}
    {{if .Build.Stderr}}<details><summary>Stderr</summary><pre>{{.Build.Stderr}}</pre></details>{{end}}
  </div>
</div>
{{end}}

</body>
</html>`

const summaryTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Evaluation Summary — {{.Summary.RunID}}</title>
<style>
  :root { --green: #22c55e; --red: #ef4444; --yellow: #eab308; --bg: #f8fafc; --text: #0f172a; --text-muted: #64748b; --border: #e2e8f0; --blue: #2563eb; --purple: #7c3aed; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 1200px; margin: 0 auto; padding: 2rem 1rem; color: var(--text); background: var(--bg); line-height: 1.6; }
  h1 { margin: 0 0 0.25rem 0; }
  h2 { margin: 1.5rem 0 0.75rem 0; }
  .subtitle { color: var(--text-muted); margin-bottom: 1.5rem; }
  a { color: var(--blue); text-decoration: none; }
  a:hover { text-decoration: underline; }
  .stats { display: flex; gap: 1rem; flex-wrap: wrap; margin: 1.25rem 0; }
  .stat { background: #fff; border: 1px solid var(--border); border-radius: 8px; padding: 1rem 1.25rem; text-align: center; min-width: 110px; }
  .stat-value { font-size: 1.5rem; font-weight: 700; }
  .stat-label { font-size: 0.8rem; color: var(--text-muted); }
  .analysis { background: #eff6ff; border: 1px solid #bfdbfe; border-radius: 10px; padding: 1.5rem; margin: 1.5rem 0; }
  .analysis h2 { margin: 0 0 0.75rem; font-size: 1.1rem; }
  .analysis-content { white-space: pre-wrap; font-size: 0.9rem; line-height: 1.6; }
  table { width: 100%; border-collapse: collapse; background: #fff; border: 1px solid var(--border); border-radius: 8px; overflow: hidden; }
  th { background: #f8fafc; padding: 0.75rem; text-align: center; font-size: 0.85rem; color: var(--text-muted); border-bottom: 2px solid var(--border); }
  th:first-child { text-align: left; }
  td { padding: 0.75rem; border-bottom: 1px solid #f1f5f9; text-align: center; vertical-align: top; }
  td:first-child { text-align: left; }
  .cell-pass { color: var(--green); }
  .cell-fail { color: var(--red); }
  .cell-icon { font-size: 1.1rem; }
  .cell-score { font-weight: 700; font-size: 0.9rem; }
  .cell-error { color: #991b1b; font-size: 0.8rem; }
  .cell-duration { font-size: 0.75rem; color: var(--text-muted); }
  .cell-files { font-size: 0.75rem; color: var(--text-muted); }
  .cell-tools { font-size: 0.7rem; color: var(--purple); font-family: monospace; }
  .cell-link { font-size: 0.75rem; margin-top: 0.25rem; }
  .detail-table { width: 100%; border-collapse: collapse; background: #fff; border: 1px solid var(--border); border-radius: 8px; overflow: hidden; margin-bottom: 2rem; }
  .detail-table th { background: #f8fafc; padding: 0.6rem 0.75rem; text-align: left; font-size: 0.8rem; color: var(--text-muted); border-bottom: 2px solid var(--border); }
  .detail-table td { padding: 0.6rem 0.75rem; border-bottom: 1px solid #f1f5f9; font-size: 0.85rem; vertical-align: top; }
  .tool-tag { display: inline-block; background: #f3f0ff; color: var(--purple); padding: 1px 6px; border-radius: 3px; font-size: 0.75rem; font-family: monospace; margin: 1px; }
</style>
</head>
<body>
<h1>📊 Evaluation Summary</h1>
<div class="subtitle">Run: <strong>{{.Summary.RunID}}</strong> — {{.Summary.Timestamp}}</div>

<div class="stats">
  <div class="stat"><div class="stat-value">{{.Summary.TotalEvals}}</div><div class="stat-label">Evaluations</div></div>
  <div class="stat"><div class="stat-value" style="color:var(--green)">{{.Summary.Passed}}</div><div class="stat-label">Passed</div></div>
  <div class="stat"><div class="stat-value" style="color:var(--red)">{{.Summary.Failed}}</div><div class="stat-label">Failed</div></div>
  <div class="stat"><div class="stat-value" style="color:#f97316">{{.Summary.Errors}}</div><div class="stat-label">Errors</div></div>
  <div class="stat"><div class="stat-value">{{fmtDuration .Summary.Duration}}</div><div class="stat-label">Duration</div></div>
</div>

<!-- ━━ AI Analysis (Issue 7) ━━ -->
{{if .Summary.Analysis}}
<div class="analysis">
  <h2>🤖 AI Analysis</h2>
  <div class="analysis-content">{{.Summary.Analysis}}</div>
</div>
{{end}}

{{if .Matrix}}
<h2>Prompt × Config Matrix</h2>
<table>
  <thead>
    <tr>
      <th>Prompt</th>
      {{range .Matrix.Configs}}<th>{{.}}</th>{{end}}
    </tr>
  </thead>
  <tbody>
    {{range $prompt := .Matrix.Prompts}}
    <tr>
      <td><code>{{$prompt}}</code></td>
      {{range $config := $.Matrix.Configs}}
      <td>
        {{with index (index $.Matrix.Cells $prompt) $config}}
          <div class="cell-icon">{{statusIcon .Success}}</div>
          {{if .Error}}<div class="cell-error" style="font-size:0.7rem">⚠️ Error</div>{{end}}
          {{if .HasReview}}<div class="cell-score" style="color:{{scoreColor .Score}}">{{.Score}}/10</div>{{end}}
          <div class="cell-duration">{{fmtDuration .Duration}}</div>
          <div class="cell-files">{{.FileCount}} files</div>
          {{if .ReportLink}}<div class="cell-link"><a href="{{.ReportLink}}">View Report →</a></div>{{end}}
        {{else}}<span style="color:#d1d5db">—</span>{{end}}
      </td>
      {{end}}
    </tr>
    {{end}}
  </tbody>
</table>
{{end}}

<!-- ━━ Detailed Results ━━ -->
{{if .Summary.Results}}
<h2>Detailed Results</h2>
<table class="detail-table">
  <thead>
    <tr>
      <th>Prompt</th>
      <th>Config</th>
      <th>Status</th>
      <th>Score</th>
      <th>Duration</th>
      <th>Files</th>
      <th>Tool Calls</th>
      <th>Report</th>
    </tr>
  </thead>
  <tbody>
    {{range .Summary.Results}}
    <tr>
      <td><code>{{.PromptID}}</code></td>
      <td>{{.ConfigName}}</td>
      <td>{{if .Error}}⚠️{{else}}{{statusIcon .Success}}{{end}}</td>
      <td>{{if .Review}}<span style="color:{{scoreColor .Review.OverallScore}};font-weight:700">{{.Review.OverallScore}}/10</span>{{else}}—{{end}}</td>
      <td>{{fmtDuration .Duration}}</td>
      <td>{{len .GeneratedFiles}}</td>
      <td>{{range .ToolCalls}}<span class="tool-tag">{{.}}</span>{{end}}</td>
      <td>{{with reportLink .}}<a href="{{.}}">View →</a>{{end}}</td>
    </tr>
    {{end}}
  </tbody>
</table>
{{end}}

<!-- ━━ Duration Analysis ━━ -->
{{if .Stats}}
{{if .Stats.DurationByConfig}}
<h2>Duration Analysis</h2>
<table class="detail-table">
  <thead><tr><th>Config</th><th>Min</th><th>Avg</th><th>Max</th></tr></thead>
  <tbody>
    {{range $cfg, $d := .Stats.DurationByConfig}}
    <tr><td>{{$cfg}}</td><td>{{fmtDuration $d.Min}}</td><td>{{fmtDuration $d.Avg}}</td><td>{{fmtDuration $d.Max}}</td></tr>
    {{end}}
  </tbody>
</table>
{{if .Stats.SlowestEval}}<p style="color:var(--text-muted);font-size:0.85rem">⏱ Slowest: <strong>{{.Stats.SlowestEval}}</strong> · Fastest: <strong>{{.Stats.FastestEval}}</strong></p>{{end}}
{{end}}

<!-- ━━ Config Comparison ━━ -->
{{if .Stats.ConfigPassRates}}
<h2>Config Comparison</h2>
<table class="detail-table">
  <thead><tr><th>Config</th><th>Total</th><th>Passed</th><th>Failed</th><th>Pass Rate</th></tr></thead>
  <tbody>
    {{range .Stats.ConfigPassRates}}
    <tr>
      <td>{{.Config}}</td>
      <td>{{.Total}}</td>
      <td style="color:var(--green)">{{.Passed}}</td>
      <td style="color:var(--red)">{{.Failed}}</td>
      <td><strong>{{printf "%.1f" .Rate}}%</strong></td>
    </tr>
    {{end}}
  </tbody>
</table>
{{if .Stats.PromptDeltas}}
<h3>Prompt Deltas (differ between configs)</h3>
<table class="detail-table">
  <thead><tr><th>Prompt</th><th>Passes On</th><th>Fails On</th></tr></thead>
  <tbody>
    {{range .Stats.PromptDeltas}}
    <tr><td><code>{{.PromptID}}</code></td><td style="color:var(--green)">{{.PassConfig}}</td><td style="color:var(--red)">{{.FailConfig}}</td></tr>
    {{end}}
  </tbody>
</table>
{{end}}
{{end}}

<!-- ━━ Tool Usage ━━ -->
{{if .Stats.ToolStats}}
<h2>Tool Usage</h2>
<table class="detail-table">
  <thead><tr><th>Tool</th><th>Calls</th><th>Successes</th><th>Failures</th><th>Success Rate</th></tr></thead>
  <tbody>
    {{range .Stats.ToolStats}}
    <tr>
      <td><span class="tool-tag">{{.Name}}</span></td>
      <td>{{.Count}}</td>
      <td style="color:var(--green)">{{.Successes}}</td>
      <td style="color:var(--red)">{{.Failures}}</td>
      <td>{{printf "%.1f" .Rate}}%</td>
    </tr>
    {{end}}
  </tbody>
</table>
{{end}}
{{end}}

</body>
</html>`
