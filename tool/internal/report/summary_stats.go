package report

import (
	"math"
	"sort"
)

// DurationStats holds min/avg/max duration statistics.
type DurationStats struct {
	Min float64 `json:"min"`
	Avg float64 `json:"avg"`
	Max float64 `json:"max"`
}

// ConfigPassRate holds pass rate info for a single config.
type ConfigPassRate struct {
	Config string  `json:"config"`
	Total  int     `json:"total"`
	Passed int     `json:"passed"`
	Failed int     `json:"failed"`
	Rate   float64 `json:"rate"` // 0-100
}

// PromptDelta identifies a prompt that differs between two configs.
type PromptDelta struct {
	PromptID   string `json:"prompt_id"`
	PassConfig string `json:"pass_config"`
	FailConfig string `json:"fail_config"`
}

// ToolStat holds aggregate stats for a single tool.
type ToolStat struct {
	Name      string  `json:"name"`
	Count     int     `json:"count"`
	Successes int     `json:"successes"`
	Failures  int     `json:"failures"`
	Rate      float64 `json:"success_rate"` // 0-100
}

// SummaryStats holds computed aggregate statistics for a run summary.
type SummaryStats struct {
	// Duration analysis
	DurationByConfig map[string]DurationStats `json:"duration_by_config"`
	DurationByPrompt map[string]DurationStats `json:"duration_by_prompt"`
	SlowestEval      string                   `json:"slowest_eval"`
	FastestEval      string                   `json:"fastest_eval"`

	// Config comparison
	ConfigPassRates []ConfigPassRate `json:"config_pass_rates"`
	PromptDeltas    []PromptDelta    `json:"prompt_deltas"`

	// Tool usage
	ToolStats []ToolStat `json:"tool_stats"`
}

// ComputeSummaryStats computes aggregate statistics from a RunSummary.
func ComputeSummaryStats(s *RunSummary) *SummaryStats {
	stats := &SummaryStats{
		DurationByConfig: make(map[string]DurationStats),
		DurationByPrompt: make(map[string]DurationStats),
	}

	// Group durations by config and prompt
	configDurations := make(map[string][]float64)
	promptDurations := make(map[string][]float64)
	configCounts := make(map[string]struct{ total, passed int })
	// Track per-prompt-per-config pass/fail for delta analysis
	promptConfigPass := make(map[string]map[string]bool)

	var slowestDur float64
	var fastestDur float64 = math.MaxFloat64

	for _, r := range s.Results {
		configDurations[r.ConfigName] = append(configDurations[r.ConfigName], r.Duration)
		promptDurations[r.PromptID] = append(promptDurations[r.PromptID], r.Duration)

		cc := configCounts[r.ConfigName]
		cc.total++
		if r.Success {
			cc.passed++
		}
		configCounts[r.ConfigName] = cc

		if promptConfigPass[r.PromptID] == nil {
			promptConfigPass[r.PromptID] = make(map[string]bool)
		}
		promptConfigPass[r.PromptID][r.ConfigName] = r.Success

		label := r.PromptID + "/" + r.ConfigName
		if r.Duration > slowestDur {
			slowestDur = r.Duration
			stats.SlowestEval = label
		}
		if r.Duration < fastestDur {
			fastestDur = r.Duration
			stats.FastestEval = label
		}
	}

	// Compute duration stats
	for cfg, durs := range configDurations {
		stats.DurationByConfig[cfg] = calcDurationStats(durs)
	}
	for pid, durs := range promptDurations {
		stats.DurationByPrompt[pid] = calcDurationStats(durs)
	}

	// Config pass rates
	for cfg, cc := range configCounts {
		rate := 0.0
		if cc.total > 0 {
			rate = float64(cc.passed) / float64(cc.total) * 100
		}
		stats.ConfigPassRates = append(stats.ConfigPassRates, ConfigPassRate{
			Config: cfg,
			Total:  cc.total,
			Passed: cc.passed,
			Failed: cc.total - cc.passed,
			Rate:   math.Round(rate*10) / 10,
		})
	}
	sort.Slice(stats.ConfigPassRates, func(i, j int) bool {
		return stats.ConfigPassRates[i].Config < stats.ConfigPassRates[j].Config
	})

	// Prompt deltas: find prompts that pass on one config but fail on another
	configs := make([]string, 0)
	for cfg := range configCounts {
		configs = append(configs, cfg)
	}
	sort.Strings(configs)

	if len(configs) >= 2 {
		for pid, cfgMap := range promptConfigPass {
			for i := 0; i < len(configs); i++ {
				for j := i + 1; j < len(configs); j++ {
					passA, okA := cfgMap[configs[i]]
					passB, okB := cfgMap[configs[j]]
					if okA && okB && passA != passB {
						delta := PromptDelta{PromptID: pid}
						if passA {
							delta.PassConfig = configs[i]
							delta.FailConfig = configs[j]
						} else {
							delta.PassConfig = configs[j]
							delta.FailConfig = configs[i]
						}
						stats.PromptDeltas = append(stats.PromptDeltas, delta)
					}
				}
			}
		}
		sort.Slice(stats.PromptDeltas, func(i, j int) bool {
			return stats.PromptDeltas[i].PromptID < stats.PromptDeltas[j].PromptID
		})
	}

	// Tool usage stats
	toolCounts := make(map[string]struct{ total, success, fail int })
	for _, r := range s.Results {
		for _, ev := range r.SessionEvents {
			if ev.Type == "tool.execution_complete" && ev.ToolName != "" {
				tc := toolCounts[ev.ToolName]
				tc.total++
				if ev.ToolSuccess != nil {
					if *ev.ToolSuccess {
						tc.success++
					} else {
						tc.fail++
					}
				}
				toolCounts[ev.ToolName] = tc
			}
		}
	}
	for name, tc := range toolCounts {
		rate := 0.0
		denom := tc.success + tc.fail
		if denom > 0 {
			rate = float64(tc.success) / float64(denom) * 100
		}
		stats.ToolStats = append(stats.ToolStats, ToolStat{
			Name:      name,
			Count:     tc.total,
			Successes: tc.success,
			Failures:  tc.fail,
			Rate:      math.Round(rate*10) / 10,
		})
	}
	sort.Slice(stats.ToolStats, func(i, j int) bool {
		return stats.ToolStats[i].Count > stats.ToolStats[j].Count
	})

	return stats
}

func calcDurationStats(durs []float64) DurationStats {
	if len(durs) == 0 {
		return DurationStats{}
	}
	mn := durs[0]
	mx := durs[0]
	sum := 0.0
	for _, d := range durs {
		sum += d
		if d < mn {
			mn = d
		}
		if d > mx {
			mx = d
		}
	}
	return DurationStats{
		Min: math.Round(mn*10) / 10,
		Avg: math.Round(sum/float64(len(durs))*10) / 10,
		Max: math.Round(mx*10) / 10,
	}
}
