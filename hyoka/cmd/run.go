package cmd

import (
"context"
"fmt"
"log/slog"
"path/filepath"
"strings"
"time"

copilot "github.com/github/copilot-sdk/go"
"github.com/ronniegeraghty/hyoka/internal/config"
"github.com/ronniegeraghty/hyoka/internal/eval"
"github.com/ronniegeraghty/hyoka/internal/prompt"
"github.com/ronniegeraghty/hyoka/internal/report"
"github.com/ronniegeraghty/hyoka/internal/review"
"github.com/ronniegeraghty/hyoka/internal/trends"
"github.com/spf13/cobra"
)

type runFlags struct {
prompts      string
service      string
language     string
plane        string
category     string
tags         string
promptID     string
configName   string
configFile   string
configDir    string
workers         int
maxSessions     int
model           string
output       string
progressMode string
skipTests    bool
skipReview   bool
skipTrends   bool
dryRun       bool
useStub      bool
// Fan-out visibility (#34)
autoConfirm bool
allConfigs  bool
// Generator guardrails (#35)
maxSessionActions int
maxFiles          int
maxOutputSize string
// Generator safety (#36)
allowCloud bool
// Resource monitoring (#45)
monitorResources bool
// Process lifecycle (#46)
strictCleanup bool
// Tiered criteria (#30)
criteriaDir string
// Directory exclusion (#63)
excludeDirs string
// Session timeout
sessionTimeout string
}

func addFilterFlags(cmd *cobra.Command, f *runFlags) {
cmd.Flags().StringVar(&f.prompts, "prompts", "./prompts", "Path to prompt library directory")
cmd.Flags().StringVar(&f.service, "service", "", "Filter by Azure service")
cmd.Flags().StringVar(&f.language, "language", "", "Filter by programming language")
cmd.Flags().StringVar(&f.plane, "plane", "", "Filter by data-plane/management-plane")
cmd.Flags().StringVar(&f.category, "category", "", "Filter by use-case category")
cmd.Flags().StringVar(&f.tags, "tags", "", "Filter by tags (comma-separated)")
cmd.Flags().StringVar(&f.promptID, "prompt-id", "", "Run a single prompt by ID")
cmd.Flags().StringVar(&f.configName, "config", "", "Config name(s) from config file (comma-separated)")
cmd.Flags().StringVar(&f.configFile, "config-file", "", "Path to a specific configuration YAML file (default: load all from configs/)")
cmd.Flags().StringVar(&f.configDir, "config-dir", "./configs", "Directory containing configuration YAML files")
cmd.Flags().IntVar(&f.workers, "workers", 0, "Parallel evaluation workers (default: number of CPUs, max 8)")
cmd.Flags().IntVar(&f.maxSessions, "max-sessions", 0, "Maximum concurrent Copilot sessions (default: workers \u00d7 3)")
cmd.Flags().StringVar(&f.model, "model", "", "Override model for all configs")
cmd.Flags().StringVar(&f.output, "output", "./reports", "Report output directory")
cmd.Flags().BoolVar(&f.skipTests, "skip-tests", false, "Skip test generation")
cmd.Flags().BoolVar(&f.skipReview, "skip-review", false, "Skip code review")
cmd.Flags().StringVar(&f.progressMode, "progress", "auto", "Progress display mode: auto, live, log, off")
cmd.Flags().BoolVar(&f.dryRun, "dry-run", false, "List matching prompts without running")
cmd.Flags().BoolVar(&f.useStub, "stub", false, "Use stub evaluator (no Copilot SDK)")
cmd.Flags().BoolVar(&f.skipTrends, "skip-trends", false, "Skip automatic trend analysis after run")
// Fan-out visibility (#34)
cmd.Flags().BoolVarP(&f.autoConfirm, "yes", "y", false, "Skip confirmation prompt for large runs (>10 evaluations)")
cmd.Flags().BoolVar(&f.allConfigs, "all-configs", false, "Run all configs when no --config filter is specified (required for multi-config runs)")
// Generator guardrails (#35)
cmd.Flags().IntVar(&f.maxSessionActions, "max-session-actions", 50, "Maximum actions per Copilot session (reasoning, response, or tool call each count as 1)")
cmd.Flags().IntVar(&f.maxFiles, "max-files", 50, "Maximum generated files per evaluation before aborting")
cmd.Flags().StringVar(&f.maxOutputSize, "max-output-size", "1MB", "Maximum total output size per evaluation (e.g., 1MB, 512KB)")
// Generator safety (#36)
cmd.Flags().BoolVar(&f.allowCloud, "allow-cloud", false, "Allow generated code to provision real Azure resources (disables safety boundaries)")
cmd.Flags().Bool("sandbox", true, "Enforce safety boundaries preventing real Azure resource provisioning (default, opposite of --allow-cloud)")
cmd.Flags().MarkHidden("sandbox") // sandbox is the default; --allow-cloud is the opt-out
// Resource monitoring (#45)
cmd.Flags().BoolVar(&f.monitorResources, "monitor-resources", false, "Monitor CPU and memory usage of Copilot sessions during evaluation")
// Process lifecycle (#46)
cmd.Flags().BoolVar(&f.strictCleanup, "strict-cleanup", false, "Fail run with non-zero exit if orphaned Copilot processes remain after cleanup")
// Tiered criteria (#30)
cmd.Flags().StringVar(&f.criteriaDir, "criteria-dir", "", "Directory containing attribute-matched criteria YAML files (e.g., criteria/)")
// Directory exclusion (#63)
cmd.Flags().StringVar(&f.excludeDirs, "exclude-dirs", "", "Comma-separated directories to exclude from generated_files output (e.g., node_modules,target,dist)")
// Session timeout
cmd.Flags().StringVar(&f.sessionTimeout, "session-timeout", "10m", "Maximum duration for a single generation or review session (e.g., 10m, 30m, 1h)")
}

func buildFilter(f *runFlags) prompt.Filter {
var tags []string
if f.tags != "" {
tags = strings.Split(f.tags, ",")
for i := range tags {
tags[i] = strings.TrimSpace(tags[i])
}
}
filters := make(map[string]string)
if f.service != "" {
filters["service"] = f.service
}
if f.plane != "" {
filters["plane"] = f.plane
}
if f.language != "" {
filters["language"] = f.language
}
if f.category != "" {
filters["category"] = f.category
}
return prompt.Filter{
Filters:  filters,
Tags:     tags,
PromptID: f.promptID,
}
}

func runCmd() *cobra.Command {
f := &runFlags{}
cmd := &cobra.Command{
Use:   "run",
Short: "Run evaluations",
Long:  "Run evaluations with optional filters against the prompt library.",
RunE: func(cmd *cobra.Command, args []string) error {
// Backward compat: --debug upgrades log level to debug
// When log-level is debug or info and progress mode is auto,
// disable live progress so slog output is visible on stderr.
if f.progressMode == "auto" {
logLevel, _ := cmd.Root().PersistentFlags().GetString("log-level")
if logLevel == "debug" || logLevel == "info" {
f.progressMode = "log"
}
}

f.prompts = resolvePromptsDir(cmd)
f.output = resolveOutputDir(cmd)

// Load config(s)
var cfgFile *config.ConfigFile
if cmd.Flags().Changed("config-file") {
f.configFile = resolveConfigFile(cmd)
var err error
cfgFile, err = config.Load(f.configFile)
if err != nil {
return fmt.Errorf("loading config: %w", err)
}
} else {
configDir := resolveConfigDir(cmd)
var err error
cfgFile, err = config.LoadDir(configDir)
if err != nil {
return fmt.Errorf("loading configs from %s: %w", configDir, err)
}
}

// Get selected configs
var configNames []string
if f.configName != "" {
configNames = strings.Split(f.configName, ",")
for i := range configNames {
configNames[i] = strings.TrimSpace(configNames[i])
}
}
configs, err := cfgFile.GetConfigs(configNames)
if err != nil {
return err
}

// Require --all-configs when multiple configs exist and no --config filter is specified (#34)
if f.configName == "" && len(configs) > 1 && !f.allConfigs {
fmt.Printf("\u26a0\ufe0f  Found %d configs but no --config filter specified.\n", len(configs))
fmt.Println("   Use --all-configs to run all configs, or --config <name> to select specific ones.")
return fmt.Errorf("multiple configs found without --config or --all-configs flag")
}

// Override model if specified via CLI flag
if f.model != "" {
for i := range configs {
if configs[i].Generator == nil {
configs[i].Generator = &config.GeneratorConfig{}
}
configs[i].Generator.Model = f.model
}
}

// Resolve relative skill_directories in configs to absolute paths
resolveConfigSkillDirs(configs, f.prompts)

// Install declared skills and plugins (npx skills add)
if err := config.InstallSkillsAndPlugins(configs); err != nil {
return fmt.Errorf("installing skills/plugins: %w", err)
}

// Load and filter prompts
prompts, err := prompt.LoadPrompts(f.prompts)
if err != nil {
return fmt.Errorf("loading prompts: %w", err)
}

filter := buildFilter(f)
filtered := prompt.FilterPrompts(prompts, filter)

if len(filtered) == 0 {
fmt.Println("\u2717 No prompts matched the given filters.")
if len(prompts) > 0 {
fmt.Printf("  (%d prompt(s) were loaded but none matched the specified filters)\n", len(prompts))
}
return fmt.Errorf("no prompts matched the given filters")
}

fmt.Printf("Found %d prompt(s), %d config(s) \u2192 %d evaluation(s)\n",
len(filtered), len(configs), len(filtered)*len(configs))

// Select evaluator and reviewer
var evaluator eval.CopilotEvaluator
var reviewerFactory eval.ReviewerFactory

// Parse session-timeout flag early \u2014 needed for reviewer setup.
sessionTimeout, err := time.ParseDuration(f.sessionTimeout)
if err != nil {
return fmt.Errorf("invalid --session-timeout %q: %w", f.sessionTimeout, err)
}

if f.useStub {
slog.Info("Using stub evaluator", "reason", "--stub flag")
fmt.Println("Using stub evaluator (--stub flag)")
evaluator = &eval.StubEvaluator{}
} else {
// Try to create a real Copilot SDK evaluator
sdkEval := eval.NewCopilotSDKEvaluator(eval.CopilotEvalOptions{
AllowCloud:        f.allowCloud,
MaxSessionActions: f.maxSessionActions,
})
sdkEval.SetSessionTimeout(sessionTimeout)
evaluator = sdkEval

// Verify Copilot CLI is available
client := copilot.NewClient(&copilot.ClientOptions{
Env: eval.HyokaBaseEnv(), // Tag verification process (#70)
})
if err := client.Start(context.Background()); err != nil {
slog.Warn("Copilot SDK unavailable, falling back to stub", "error", err)
fmt.Printf("\u26a0\ufe0f  Copilot SDK unavailable (%v), falling back to stub evaluator\n", err)
evaluator = &eval.StubEvaluator{}
} else {
defer client.Stop() // #65: ensure cleanup on any exit path
slog.Info("Using Copilot SDK evaluator")
fmt.Println("Using Copilot SDK evaluator")

clientOpts := &copilot.ClientOptions{
Env: eval.HyokaBaseEnv(), // Tag reviewer processes (#70)
}
if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
clientOpts.LogLevel = "debug"
}

// Extract reviewer skill directories from configs
var reviewerSkillsDirs []string
for _, c := range configs {
if c.Reviewer != nil {
for _, s := range c.Reviewer.Skills {
if s.Type == "local" && s.Path != "" {
reviewerSkillsDirs = append(reviewerSkillsDirs, s.Path)
}
}
}
}

// Create reviewer factory that builds a reviewer per config (#92)
reviewerFactory = func(cfg *config.ToolConfig) (review.Reviewer, *review.PanelReviewer, error) {
var reviewerModels []string
if cfg.Reviewer != nil && len(cfg.Reviewer.Models) > 0 {
reviewerModels = cfg.Reviewer.Models
}
if len(reviewerModels) == 0 {
return nil, nil, nil
}

if len(reviewerModels) > 1 {
// Multi-model panel
panelReviewer := review.NewPanelReviewer(clientOpts, reviewerModels, f.maxSessionActions)
panelReviewer.SetSessionTimeout(sessionTimeout)
if len(reviewerSkillsDirs) > 0 {
panelReviewer.SetSkillDirectories(reviewerSkillsDirs)
}
slog.Debug("Created review panel for config", "config", cfg.Name, "models", reviewerModels)
return nil, panelReviewer, nil
}

// Single reviewer
reviewClient := copilot.NewClient(clientOpts)
if err := reviewClient.Start(context.Background()); err != nil {
return nil, nil, fmt.Errorf("could not start reviewer client: %w", err)
}
copilotReviewer := review.NewCopilotReviewer(reviewClient, reviewerModels[0], f.maxSessionActions)
copilotReviewer.SetSessionTimeout(sessionTimeout)
if len(reviewerSkillsDirs) > 0 {
copilotReviewer.SetSkillDirectories(reviewerSkillsDirs)
}
slog.Debug("Created single reviewer for config", "config", cfg.Name, "model", reviewerModels[0])
return copilotReviewer, nil, nil
}

}
}

if f.skipReview {
reviewerFactory = nil
}

// Parse max-output-size flag (#35)
maxOutputSize, err := parseByteSize(f.maxOutputSize)
if err != nil {
return fmt.Errorf("invalid --max-output-size %q: %w", f.maxOutputSize, err)
}

// Create and run engine
// Parse exclude-dirs (#63)
var excludeDirs []string
if f.excludeDirs != "" {
for _, d := range strings.Split(f.excludeDirs, ",") {
d = strings.TrimSpace(d)
if d != "" {
excludeDirs = append(excludeDirs, d)
}
}
}

engine := eval.NewEngineWithReviewerFactory(evaluator, reviewerFactory, eval.EngineOptions{
Workers:           f.workers,
MaxSessions:       f.maxSessions,
OutputDir:         f.output,
SkipTests:         f.skipTests,
SkipReview:        f.skipReview,
DryRun:            f.dryRun,
ProgressMode:      f.progressMode,
ConfirmLargeRuns:  true,
AutoConfirm:       f.autoConfirm,
MaxSessionActions: f.maxSessionActions,
MaxFiles:          f.maxFiles,
MaxOutputSize:     maxOutputSize,
MonitorResources:  f.monitorResources,
StrictCleanup:     f.strictCleanup,
CriteriaDir:       f.criteriaDir,
ExcludeDirs:       excludeDirs,
SessionTimeout:    sessionTimeout,
})

summary, err := engine.Run(context.Background(), filtered, configs)
if err != nil {
return fmt.Errorf("evaluation failed: %w", err)
}

fmt.Printf("\nRun Summary:\n")
fmt.Printf("  Run ID:      %s\n", summary.RunID)
fmt.Printf("  Evaluations: %d\n", summary.TotalEvals)
fmt.Printf("  Passed:      %d\n", summary.Passed)
fmt.Printf("  Failed:      %d\n", summary.Failed)
fmt.Printf("  Errors:      %d\n", summary.Errors)
fmt.Printf("  Duration:    %.2fs\n", summary.Duration)

// Auto-run trend analysis unless opted out
if !f.skipTrends && !f.dryRun {
fmt.Printf("\n%s\n", strings.Repeat("\u2500", 60))
fmt.Println("\U0001f4ca Generating trend analysis...")

trendsOutputDir := filepath.Join(f.output, "trends")
tr, err := trends.Generate(trends.TrendOptions{
ReportsDir: f.output,
OutputDir:  trendsOutputDir,
Analyze:    false, // generate data first, analyze below
})
if err != nil {
slog.Warn("Trend generation failed", "error", err)
fmt.Printf("\u26a0\ufe0f  Trend generation failed: %v\n", err)
} else if tr.TotalRuns > 0 {
fmt.Println("\U0001f916 Running AI-powered trend analysis...")
analysis, aErr := trends.AnalyzeTrends(context.Background(), tr)
if aErr != nil {
slog.Warn("AI trend analysis failed", "error", aErr)
fmt.Printf("\u26a0\ufe0f  AI analysis failed: %v (continuing without analysis)\n", aErr)
} else {
tr.Analysis = analysis
fmt.Println("\n--- AI Analysis ---")
fmt.Println(analysis)
fmt.Println("-------------------")

// Re-write summary HTML with AI analysis included (Issue 7)
summary.Analysis = analysis
if _, err := report.WriteSummaryHTML(summary, f.output); err != nil {
slog.Warn("Failed to update summary with analysis", "error", err)
fmt.Printf("\u26a0\ufe0f  Failed to update summary with analysis: %v\n", err)
}
}

mdPath, _ := trends.WriteMarkdown(tr, trendsOutputDir)
htmlPath, _ := trends.WriteHTML(tr, trendsOutputDir)
if mdPath != "" {
fmt.Printf("Trend report (MD):   %s\n", mdPath)
}
if htmlPath != "" {
fmt.Printf("Trend report (HTML): %s\n", htmlPath)
}
fmt.Printf("Analyzed %d evaluation(s) across %d prompt(s)\n", tr.TotalRuns, len(tr.PromptTrends))
} else {
fmt.Println("No historical data found for trend analysis.")
}
}

return nil
},
}

addFilterFlags(cmd, f)
return cmd
}
