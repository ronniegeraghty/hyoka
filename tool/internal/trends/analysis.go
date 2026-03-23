package trends

import (
"context"
"fmt"
"strings"
"sync"

copilot "github.com/github/copilot-sdk/go"
)

// AnalyzeTrends uses a Copilot SDK session to perform AI-powered trend analysis.
// It returns the analysis text to be included in the report.
func AnalyzeTrends(ctx context.Context, tr *TrendReport) (string, error) {
prompt := formatTrendPrompt(tr)

client := copilot.NewClient(nil)
if err := client.Start(ctx); err != nil {
return "", fmt.Errorf("starting copilot client: %w", err)
}
defer client.Stop()

session, err := client.CreateSession(ctx, &copilot.SessionConfig{
Model: "gpt-4.1",
SystemMessage: &copilot.SystemMessageConfig{
Mode:    "append",
Content: "You are an expert at analyzing software evaluation trends. Be concise and actionable.",
},
OnPermissionRequest: copilot.PermissionHandler.ApproveAll,
})
if err != nil {
return "", fmt.Errorf("creating analysis session: %w", err)
}
defer session.Disconnect()

// Capture assistant messages via event subscription
var assistantContent strings.Builder
var mu sync.Mutex
unsub := session.On(func(event copilot.SessionEvent) {
if event.Type == copilot.SessionEventTypeAssistantMessage && event.Data.Content != nil {
mu.Lock()
assistantContent.WriteString(*event.Data.Content)
mu.Unlock()
}
})
defer unsub()

_, err = session.SendAndWait(ctx, copilot.MessageOptions{
Prompt: prompt,
})
if err != nil {
return "", fmt.Errorf("analysis request failed: %w", err)
}

mu.Lock()
result := strings.TrimSpace(assistantContent.String())
mu.Unlock()

if result == "" {
return "", fmt.Errorf("empty analysis response")
}

return result, nil
}

// formatTrendPrompt builds a structured prompt for the AI to analyze.
func formatTrendPrompt(tr *TrendReport) string {
var b strings.Builder

b.WriteString("Analyze the following AI code generation evaluation performance data.\n\n")
b.WriteString("Each prompt tests whether an AI agent can generate correct code for a specific Azure SDK scenario. ")
b.WriteString("Configs represent different tool setups (e.g., baseline vs azure-mcp).\n\n")
b.WriteString("Provide:\n")
b.WriteString("1. **Patterns** — consistent behaviors across prompts or configs\n")
b.WriteString("2. **Regressions** — prompts that started failing after previously passing\n")
b.WriteString("3. **Recommendations** — actionable suggestions to improve pass rates\n")
b.WriteString("4. **Anomalies** — unusual durations, sudden changes, or outliers\n\n")
b.WriteString("Be concise and specific. Reference prompt IDs and config names. Use bullet points.\n\n")
b.WriteString("---\n\n")

// Overall stats
passed, failed := 0, 0
for _, e := range tr.Entries {
if e.Success {
passed++
} else {
failed++
}
}
fmt.Fprintf(&b, "## Overview\n- Total evaluations: %d\n- Passed: %d (%.0f%%)\n- Failed: %d\n- Unique prompts: %d\n- Run count: %d\n\n",
tr.TotalRuns, passed, pct(passed, tr.TotalRuns), failed, len(tr.PromptTrends), len(tr.RunIDs))

// Per-prompt detail
b.WriteString("## Per-Prompt Performance\n\n")
for _, pt := range tr.PromptTrends {
fmt.Fprintf(&b, "### %s\n", pt.PromptID)
for cfg, runs := range pt.Configs {
p, f := 0, 0
totalDur := 0.0
for _, r := range runs {
if r.Success {
p++
} else {
f++
}
totalDur += r.Duration
}
avgDur := 0.0
if len(runs) > 0 {
avgDur = totalDur / float64(len(runs))
}
trend := pt.Trend[cfg]
fmt.Fprintf(&b, "- **%s**: %d/%d passed (%.0f%%), avg duration: %s, trend: %s\n",
cfg, p, p+f, pct(p, p+f), formatDuration(avgDur), trend)

if f > 0 {
b.WriteString("  Run history: ")
for i, r := range runs {
if i > 0 {
b.WriteString(", ")
}
icon := "PASS"
if !r.Success {
icon = "FAIL"
}
fmt.Fprintf(&b, "%s(%s)", icon, r.RunID)
}
b.WriteString("\n")
}
}
b.WriteString("\n")
}

return b.String()
}
