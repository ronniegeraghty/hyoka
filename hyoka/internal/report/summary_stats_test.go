package report

import (
	"testing"
)

func TestComputeSummaryStats(t *testing.T) {
	boolTrue := true
	boolFalse := false

	s := &RunSummary{
		Results: []*EvalReport{
			{
				PromptID:   "p1",
				ConfigName: "baseline",
				Duration:   10.0,
				Success:    true,
				SessionEvents: []SessionEventRecord{
					{Type: "tool.execution_complete", ToolName: "create", ToolSuccess: &boolTrue},
					{Type: "tool.execution_complete", ToolName: "edit", ToolSuccess: &boolTrue},
				},
			},
			{
				PromptID:   "p1",
				ConfigName: "azure-mcp",
				Duration:   15.0,
				Success:    true,
				SessionEvents: []SessionEventRecord{
					{Type: "tool.execution_complete", ToolName: "create", ToolSuccess: &boolTrue},
					{Type: "tool.execution_complete", ToolName: "azure_mcp", ToolSuccess: &boolFalse},
				},
			},
			{
				PromptID:   "p2",
				ConfigName: "baseline",
				Duration:   5.0,
				Success:    false,
				SessionEvents: []SessionEventRecord{
					{Type: "tool.execution_complete", ToolName: "create", ToolSuccess: &boolTrue},
				},
			},
			{
				PromptID:   "p2",
				ConfigName: "azure-mcp",
				Duration:   12.0,
				Success:    true,
				SessionEvents: []SessionEventRecord{
					{Type: "tool.execution_complete", ToolName: "create", ToolSuccess: &boolTrue},
				},
			},
		},
	}

	stats := ComputeSummaryStats(s)

	// Duration by config
	if len(stats.DurationByConfig) != 2 {
		t.Errorf("expected 2 config duration entries, got %d", len(stats.DurationByConfig))
	}
	bl := stats.DurationByConfig["baseline"]
	if bl.Min != 5.0 || bl.Max != 10.0 {
		t.Errorf("baseline duration: expected min=5 max=10, got min=%.1f max=%.1f", bl.Min, bl.Max)
	}

	// Duration by prompt
	if len(stats.DurationByPrompt) != 2 {
		t.Errorf("expected 2 prompt duration entries, got %d", len(stats.DurationByPrompt))
	}

	// Slowest/fastest
	if stats.SlowestEval != "p1/azure-mcp" {
		t.Errorf("expected slowest p1/azure-mcp, got %s", stats.SlowestEval)
	}
	if stats.FastestEval != "p2/baseline" {
		t.Errorf("expected fastest p2/baseline, got %s", stats.FastestEval)
	}

	// Config pass rates
	if len(stats.ConfigPassRates) != 2 {
		t.Fatalf("expected 2 config pass rates, got %d", len(stats.ConfigPassRates))
	}
	for _, cpr := range stats.ConfigPassRates {
		if cpr.Config == "baseline" && cpr.Rate != 50.0 {
			t.Errorf("baseline pass rate: expected 50%%, got %.1f%%", cpr.Rate)
		}
		if cpr.Config == "azure-mcp" && cpr.Rate != 100.0 {
			t.Errorf("azure-mcp pass rate: expected 100%%, got %.1f%%", cpr.Rate)
		}
	}

	// Prompt deltas (p2 passes on azure-mcp but fails on baseline)
	if len(stats.PromptDeltas) != 1 {
		t.Fatalf("expected 1 prompt delta, got %d", len(stats.PromptDeltas))
	}
	if stats.PromptDeltas[0].PromptID != "p2" {
		t.Errorf("expected delta for p2, got %s", stats.PromptDeltas[0].PromptID)
	}

	// Tool usage
	if len(stats.ToolStats) == 0 {
		t.Fatal("expected tool stats")
	}
	// "create" should be most used (4 times)
	if stats.ToolStats[0].Name != "create" || stats.ToolStats[0].Count != 4 {
		t.Errorf("expected create with count 4, got %s with %d", stats.ToolStats[0].Name, stats.ToolStats[0].Count)
	}
}

func TestComputeSummaryStatsEmpty(t *testing.T) {
	s := &RunSummary{Results: []*EvalReport{}}
	stats := ComputeSummaryStats(s)
	if len(stats.DurationByConfig) != 0 {
		t.Error("expected empty stats for empty results")
	}
	if len(stats.ToolStats) != 0 {
		t.Error("expected empty tool stats for empty results")
	}
}

func TestCalcDurationStats(t *testing.T) {
	ds := calcDurationStats([]float64{3.0, 7.0, 5.0})
	if ds.Min != 3.0 {
		t.Errorf("expected min 3, got %.1f", ds.Min)
	}
	if ds.Max != 7.0 {
		t.Errorf("expected max 7, got %.1f", ds.Max)
	}
	if ds.Avg != 5.0 {
		t.Errorf("expected avg 5, got %.1f", ds.Avg)
	}

	empty := calcDurationStats(nil)
	if empty.Min != 0 || empty.Avg != 0 || empty.Max != 0 {
		t.Error("expected zero stats for nil slice")
	}
}
