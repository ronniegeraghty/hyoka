package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	copilot "github.com/github/copilot-sdk/go"
	"github.com/ronniegeraghty/hyoka/internal/checkenv"
	"github.com/ronniegeraghty/hyoka/internal/clean"
	"github.com/ronniegeraghty/hyoka/internal/config"
	"github.com/ronniegeraghty/hyoka/internal/eval"
	"github.com/ronniegeraghty/hyoka/internal/logging"
	"github.com/ronniegeraghty/hyoka/internal/plugin"
	"github.com/ronniegeraghty/hyoka/internal/prompt"
	"github.com/ronniegeraghty/hyoka/internal/report"
	"github.com/ronniegeraghty/hyoka/internal/rerender"
	"github.com/ronniegeraghty/hyoka/internal/review"
	"github.com/ronniegeraghty/hyoka/internal/serve"
	"github.com/ronniegeraghty/hyoka/internal/trends"
	"github.com/ronniegeraghty/hyoka/internal/validate"
	"github.com/spf13/cobra"
)

var version = "0.2.0"

func main() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	var logLevel, logFile string

	root := &cobra.Command{
		Use:   "hyoka",
		Short: "Azure SDK Prompt Evaluation Tool — test AI agent code generation quality",
		Long:  "A tool for evaluating AI agent code generation quality by running prompts through the Copilot SDK, running build verification, and generating reports.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			closer, err := logging.Setup(logging.Options{
				Level:    logLevel,
				FilePath: logFile,
			})
			if err != nil {
				return err
			}
			// Store closer on the context so it can be called at shutdown.
			// For simplicity we use a runtime finalizer via cobra's PostRun;
			// in practice the process exits right after Execute returns.
			cmd.Root().PersistentPostRun = func(*cobra.Command, []string) { closer() }
			return nil
		},
	}

	root.PersistentFlags().StringVar(&logLevel, "log-level", "warn", "Log level: debug, info, warn, error")
	root.PersistentFlags().StringVar(&logFile, "log-file", "", "Redirect log output to a file (stderr stays clean)")

	root.AddCommand(runCmd())
	root.AddCommand(listCmd())
	root.AddCommand(configsCmd())
	root.AddCommand(versionCmd())

	root.AddCommand(validateCmd())
	root.AddCommand(checkEnvCmd())
	root.AddCommand(trendsCmd())
	root.AddCommand(reportCmd())
	root.AddCommand(newPromptCmd())
	root.AddCommand(serveCmd())
	root.AddCommand(pluginsCmd())
	root.AddCommand(cleanCmd())

	return root
}

// resolvePathFlag returns the flag value if explicitly set by the user,
// otherwise tries the candidate paths in order, falling back to the default.
func resolvePathFlag(cmd *cobra.Command, flagName string, candidates []string) string {
	if cmd.Flags().Changed(flagName) {
		val, _ := cmd.Flags().GetString(flagName)
		return val
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	val, _ := cmd.Flags().GetString(flagName)
	return val
}

func resolvePromptsDir(cmd *cobra.Command) string {
	return resolvePathFlag(cmd, "prompts", []string{"./prompts", "../prompts"})
}

func resolveConfigFile(cmd *cobra.Command) string {
	return resolvePathFlag(cmd, "config-file", []string{
		"./configs", "../configs",
	})
}

func resolveConfigDir(cmd *cobra.Command) string {
	return resolvePathFlag(cmd, "config-dir", []string{
		"./configs", "../configs",
	})
}

func resolveOutputDir(cmd *cobra.Command) string {
	return resolvePathFlag(cmd, "output", []string{"./reports", "../reports"})
}

func resolveOutputFile(cmd *cobra.Command, candidates []string) string {
	return resolvePathFlag(cmd, "output", candidates)
}

type runFlags struct {
	prompts      string
	service      string
	language     string
	plane        string
	category     string
	tags         string
	promptID     string
	filter       []string
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
	cmd.Flags().StringVar(&f.service, "service", "", "Filter by Azure service (shorthand for --filter service=VALUE)")
	cmd.Flags().StringVar(&f.language, "language", "", "Filter by programming language (shorthand for --filter language=VALUE)")
	cmd.Flags().StringVar(&f.plane, "plane", "", "Filter by data-plane/management-plane (shorthand for --filter plane=VALUE)")
	cmd.Flags().StringVar(&f.category, "category", "", "Filter by use-case category (shorthand for --filter category=VALUE)")
	cmd.Flags().StringVar(&f.tags, "tags", "", "Filter by tags (comma-separated)")
	cmd.Flags().StringVar(&f.promptID, "prompt-id", "", "Run a single prompt by ID")
	cmd.Flags().StringArrayVar(&f.filter, "filter", nil, "Filter by arbitrary property (key=value, repeatable)")
	cmd.Flags().StringVar(&f.configName, "config", "", "Config name(s) from config file (comma-separated)")
	cmd.Flags().StringVar(&f.configFile, "config-file", "", "Path to a specific configuration YAML file (default: load all from configs/)")
	cmd.Flags().StringVar(&f.configDir, "config-dir", "./configs", "Directory containing configuration YAML files")
	cmd.Flags().IntVar(&f.workers, "workers", 0, "Parallel evaluation workers (default: number of CPUs, max 8)")
	cmd.Flags().IntVar(&f.maxSessions, "max-sessions", 0, "Maximum concurrent Copilot sessions (default: workers × 3)")
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


// resolveConfigSkillDirs resolves relative skill_directories in loaded configs
// to absolute paths so they work regardless of which directory the tool is invoked from.
// Handles both legacy top-level fields and new Generator/Reviewer sub-struct skills.
func resolveConfigSkillDirs(configs []config.ToolConfig, promptsDir string) {

	resolveSkills := func(skills []config.Skill) {
		for j := range skills {
			if skills[j].Type == "local" && skills[j].Path != "" && !filepath.IsAbs(skills[j].Path) {
				candidates := []string{
					skills[j].Path,
					filepath.Join(filepath.Dir(promptsDir), skills[j].Path),
				}
				for _, c := range candidates {
					if info, err := os.Stat(c); err == nil && info.IsDir() {
						abs, absErr := filepath.Abs(c)
						if absErr != nil {
							slog.Warn("Failed to resolve absolute skill path", "path", c, "error", absErr)
						}
						skills[j].Path = abs
						break
					}
				}
			}
		}
	}

	for i := range configs {
		// Resolve new sub-struct skill paths
		if configs[i].Generator != nil {
			resolveSkills(configs[i].Generator.Skills)
		}
		if configs[i].Reviewer != nil {
			resolveSkills(configs[i].Reviewer.Skills)
		}
	}
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
	for _, kv := range f.filter {
		k, v, ok := strings.Cut(kv, "=")
		if ok && k != "" {
			filters[k] = v
		}
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
				fmt.Printf("⚠️  Found %d configs but no --config filter specified.\n", len(configs))
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
				fmt.Println("✗ No prompts matched the given filters.")
				if len(prompts) > 0 {
					fmt.Printf("  (%d prompt(s) were loaded but none matched the specified filters)\n", len(prompts))
				}
				return fmt.Errorf("no prompts matched the given filters")
			}

			fmt.Printf("Found %d prompt(s), %d config(s) → %d evaluation(s)\n",
				len(filtered), len(configs), len(filtered)*len(configs))

			// Select evaluator and reviewer
			var evaluator eval.CopilotEvaluator
			var reviewerFactory eval.ReviewerFactory

			// Parse session-timeout flag early — needed for reviewer setup.
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
					fmt.Printf("⚠️  Copilot SDK unavailable (%v), falling back to stub evaluator\n", err)
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
						var reviewerSystemPrompt string
						if cfg.Reviewer != nil {
							reviewerModels = cfg.Reviewer.Models
							reviewerSystemPrompt = cfg.Reviewer.SystemPrompt
						}
						if len(reviewerModels) == 0 {
							return nil, nil, nil
						}

						if len(reviewerModels) > 1 {
							// Multi-model panel
							panelReviewer := review.NewPanelReviewer(clientOpts, reviewerModels, f.maxSessionActions)
							panelReviewer.SetSessionTimeout(sessionTimeout)
							if reviewerSystemPrompt != "" {
								panelReviewer.SetSystemPrompt(reviewerSystemPrompt)
							}
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
						if reviewerSystemPrompt != "" {
							copilotReviewer.SetSystemPrompt(reviewerSystemPrompt)
						}
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
				fmt.Printf("\n%s\n", strings.Repeat("─", 60))
				fmt.Println("📊 Generating trend analysis...")

				trendsOutputDir := filepath.Join(f.output, "trends")
				tr, err := trends.Generate(trends.TrendOptions{
					ReportsDir: f.output,
					OutputDir:  trendsOutputDir,
					Analyze:    false, // generate data first, analyze below
				})
				if err != nil {
					slog.Warn("Trend generation failed", "error", err)
					fmt.Printf("⚠️  Trend generation failed: %v\n", err)
				} else if tr.TotalRuns > 0 {
					fmt.Println("🤖 Running AI-powered trend analysis...")
					analysis, aErr := trends.AnalyzeTrends(context.Background(), tr)
					if aErr != nil {
						slog.Warn("AI trend analysis failed", "error", aErr)
						fmt.Printf("⚠️  AI analysis failed: %v (continuing without analysis)\n", aErr)
					} else {
						tr.Analysis = analysis
						fmt.Println("\n--- AI Analysis ---")
						fmt.Println(analysis)
						fmt.Println("-------------------")

						// Re-write summary HTML with AI analysis included (Issue 7)
						summary.Analysis = analysis
						if _, err := report.WriteSummaryHTML(summary, f.output); err != nil {
							slog.Warn("Failed to update summary with analysis", "error", err)
							fmt.Printf("⚠️  Failed to update summary with analysis: %v\n", err)
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

func listCmd() *cobra.Command {
	f := &runFlags{}
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List matching prompts",
		Long:    "List prompts matching the given filters (dry-run equivalent).",
		RunE: func(cmd *cobra.Command, args []string) error {
			f.prompts = resolvePromptsDir(cmd)

			prompts, err := prompt.LoadPrompts(f.prompts)
			if err != nil {
				return fmt.Errorf("loading prompts: %w", err)
			}

			filter := buildFilter(f)
			filtered := prompt.FilterPrompts(prompts, filter)

			if len(filtered) == 0 {
				fmt.Println("No prompts matched the given filters.")
				return nil
			}

			if jsonOutput {
				data, err := json.MarshalIndent(filtered, "", "  ")
				if err != nil {
					return fmt.Errorf("marshaling prompts: %w", err)
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Printf("Found %d prompt(s):\n\n", len(filtered))
			for _, p := range filtered {
				fmt.Printf("  %-30s %s/%s/%s [%s]\n", p.ID, p.Service(), p.Plane(), p.Language(), p.Category())
				if p.Description() != "" {
					fmt.Printf("  %-30s %s\n", "", p.Description())
				}
			}
			return nil
		},
	}

	addFilterFlags(cmd, f)
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output prompts as JSON array")
	return cmd
}

func configsCmd() *cobra.Command {
	var configFile string
	var configDir string

	cmd := &cobra.Command{
		Use:   "configs",
		Short: "List available configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfgFile *config.ConfigFile
			if cmd.Flags().Changed("config-file") {
				configFile = resolveConfigFile(cmd)
				var err error
				cfgFile, err = config.Load(configFile)
				if err != nil {
					return fmt.Errorf("loading config: %w", err)
				}
			} else {
				configDir = resolveConfigDir(cmd)
				var err error
				cfgFile, err = config.LoadDir(configDir)
				if err != nil {
					return fmt.Errorf("loading configs from %s: %w", configDir, err)
				}
			}

			fmt.Printf("Available configurations (%d):\n\n", len(cfgFile.Configs))
			for _, c := range cfgFile.Configs {
				model := ""
				if c.Generator != nil {
					model = c.Generator.Model
				}
				fmt.Printf("  %-20s %s (model: %s)\n", c.Name, c.Description, model)
				var mcpServers map[string]*config.MCPServer
				if c.Generator != nil {
					mcpServers = c.Generator.MCPServers
				}
				if len(mcpServers) > 0 {
					fmt.Printf("  %-20s MCP servers: ", "")
					var names []string
					for name := range mcpServers {
						names = append(names, name)
					}
					fmt.Println(strings.Join(names, ", "))
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&configFile, "config-file", "", "Path to a specific configuration YAML file")
	cmd.Flags().StringVar(&configDir, "config-dir", "./configs", "Directory containing configuration YAML files")
	return cmd
}

func validateCmd() *cobra.Command {
	var promptsDir string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate prompts and configs",
		Long:  "Validate all prompt files against schema rules and naming conventions, and validate config files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			promptsDir = resolvePromptsDir(cmd)
			allOK := true

			// Validate prompts
			result, err := validate.Validate(promptsDir)
			if err != nil {
				// Zero prompts found — show near-miss suggestions
				nearMisses := prompt.ScanNearMisses(promptsDir)
				fmt.Printf("✗ No prompts found in %s\n", promptsDir)
				if len(nearMisses) > 0 {
					fmt.Println("\nDid you mean one of these?")
					for _, nm := range nearMisses {
						fmt.Printf("  %s\n", nm)
					}
				}
				os.Exit(1)
			}
			if result.TotalFiles == 0 {
				nearMisses := prompt.ScanNearMisses(promptsDir)
				fmt.Printf("✗ No prompts found in %s\n", promptsDir)
				if len(nearMisses) > 0 {
					fmt.Println("\nDid you mean one of these?")
					for _, nm := range nearMisses {
						fmt.Printf("  %s\n", nm)
					}
				}
				os.Exit(1)
			}
			fmt.Println(validate.FormatResult(result))
			if !result.OK() {
				allOK = false
			}

			// Validate config files
			configDir := filepath.Join(filepath.Dir(promptsDir), "configs")
			if entries, err := os.ReadDir(configDir); err == nil {
				configCount := 0
				configErrors := 0
				for _, e := range entries {
					if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
						continue
					}
					cfgPath := filepath.Join(configDir, e.Name())
					_, cfgErr := config.Load(cfgPath)
					configCount++
					if cfgErr != nil {
						fmt.Printf("✗ Config %s: %v\n", e.Name(), cfgErr)
						configErrors++
						allOK = false
					}
				}
				if configCount > 0 {
					if configErrors == 0 {
						fmt.Printf("✓ All %d config(s) are valid\n", configCount)
					} else {
						fmt.Printf("✗ %d of %d config(s) have errors\n", configErrors, configCount)
					}
				}
			}

			if !allOK {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&promptsDir, "prompts", "./prompts", "Path to prompt library directory")
	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("hyoka version %s\n", version)
		},
	}
}

func checkEnvCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "check-env",
		Aliases: []string{"env"},
		Short:   "Check for required language toolchains and tools",
		Long:    "Tests if language toolchains (dotnet, python, go, node, java, rust, cargo, cmake, etc.), Copilot CLI, and MCP prerequisites are installed.",
		Run: func(cmd *cobra.Command, args []string) {
			checkenv.Run()
		},
	}
}

func trendsCmd() *cobra.Command {
	var promptID, service, language, reportsDir, output string
	var analyze bool
	var openBrowser bool

	cmd := &cobra.Command{
		Use:   "trends",
		Short: "Generate historical trend reports with time-series performance data",
		Long:  "Scans all past runs in reports/ directory and generates a trend report with pass-rate timelines, duration trends, config comparisons, and regression detection. AI-powered insights are included by default; use --no-analyze to skip.",
		RunE: func(cmd *cobra.Command, args []string) error {
			reportsDir = resolvePathFlag(cmd, "reports-dir", []string{"../reports", "./reports"})
			if !cmd.Flags().Changed("output") {
				output = filepath.Join(reportsDir, "trends")
			}

			tr, err := trends.Generate(trends.TrendOptions{
				ReportsDir: reportsDir,
				PromptID:   promptID,
				Service:    service,
				Language:   language,
				OutputDir:  output,
				Analyze:    analyze,
			})
			if err != nil {
				return fmt.Errorf("generating trends: %w", err)
			}

			if tr.TotalRuns == 0 {
				fmt.Println("No historical data found matching the given filters.")
				return nil
			}

			// Run Copilot-powered analysis if requested
			if analyze {
				fmt.Println("🤖 Running AI-powered trend analysis...")
				analysis, err := trends.AnalyzeTrends(context.Background(), tr)
				if err != nil {
					slog.Warn("AI trend analysis failed", "error", err)
					fmt.Printf("⚠️  AI analysis failed: %v (continuing without analysis)\n", err)
				} else {
					tr.Analysis = analysis
					fmt.Println("\n--- AI Analysis ---")
					fmt.Println(analysis)
					fmt.Println("-------------------")
				}
			}

			mdPath, err := trends.WriteMarkdown(tr, output)
			if err != nil {
				return fmt.Errorf("writing markdown trends: %w", err)
			}
			fmt.Printf("Markdown trend report: %s\n", mdPath)

			htmlPath, err := trends.WriteHTML(tr, output)
			if err != nil {
				return fmt.Errorf("writing HTML trends: %w", err)
			}
			fmt.Printf("HTML trend report:     %s\n", htmlPath)
			fmt.Printf("\nAnalyzed %d historical evaluation(s) across %d prompt(s)\n", tr.TotalRuns, len(tr.PromptTrends))

			if openBrowser && htmlPath != "" {
				openInBrowser(htmlPath)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&promptID, "prompt-id", "", "Filter trends by prompt ID")
	cmd.Flags().StringVar(&service, "service", "", "Filter trends by Azure service")
	cmd.Flags().StringVar(&language, "language", "", "Filter trends by programming language")
	cmd.Flags().StringVar(&reportsDir, "reports-dir", "./reports", "Directory containing past evaluation reports")
	cmd.Flags().StringVar(&output, "output", "./reports/trends", "Output directory for trend reports")
	cmd.Flags().BoolVar(&analyze, "analyze", true, "Run AI-powered analysis of trends (enabled by default)")
	cmd.Flags().BoolVar(&openBrowser, "open", false, "Auto-open the HTML trend report in the browser")

	// --no-analyze opt-out: cobra doesn't auto-generate negation flags,
	// so we register a separate bool and reconcile in RunE.
	var noAnalyze bool
	cmd.Flags().BoolVar(&noAnalyze, "no-analyze", false, "Skip AI-powered trend analysis")
	// Wire no-analyze into analyze before RunE executes
	origRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if noAnalyze {
			analyze = false
		}
		return origRunE(cmd, args)
	}

	return cmd
}

func reportCmd() *cobra.Command {
	var reportsDir string
	var all bool

	cmd := &cobra.Command{
		Use:   "report [run-id]",
		Short: "Re-render HTML/MD reports from existing report.json files",
		Long:  "Re-generates report.html, report.md, summary.html, and summary.md using current templates without re-running evaluations. Useful after template improvements.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reportsDir = resolvePathFlag(cmd, "reports-dir", []string{"../reports", "./reports"})

			var runID string
			if len(args) > 0 {
				runID = args[0]
			}

			if !all && runID == "" {
				return fmt.Errorf("specify a run ID or use --all to re-render all runs")
			}

			return rerender.Run(rerender.Options{
				ReportsDir: reportsDir,
				RunID:      runID,
				All:        all,
			})
		},
	}

	cmd.Flags().StringVar(&reportsDir, "reports-dir", "./reports", "Directory containing evaluation reports")
	cmd.Flags().BoolVar(&all, "all", false, "Re-render all runs")

	return cmd
}

func serveCmd() *cobra.Command {
	var port int
	var reportsDir string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start a local web server to browse evaluation reports",
		Long:  "Starts an HTTP server that provides a web UI for browsing past evaluation runs, viewing summaries, and individual report pages.",
		RunE: func(cmd *cobra.Command, args []string) error {
			reportsDir = resolveOutputFile(cmd, []string{"./reports", "../reports"})
			return serve.Start(serve.Options{
				ReportsDir: reportsDir,
				Port:       port,
			})
		},
	}

	cmd.Flags().IntVar(&port, "port", 8080, "Port to serve on")
	cmd.Flags().StringVar(&reportsDir, "output", "./reports", "Directory containing evaluation reports")

	return cmd
}

func pluginsCmd() *cobra.Command {
	var pluginsDir string

	cmd := &cobra.Command{
		Use:   "plugins",
		Short: "List available plugins",
		Long:  "Scans the plugins directory and lists all available plugin definitions with their skills and MCP servers.",
		RunE: func(cmd *cobra.Command, args []string) error {
			reg := plugin.NewRegistry()
			if err := reg.LoadDir(pluginsDir); err != nil {
				return fmt.Errorf("loading plugins: %w", err)
			}

			plugins := reg.All()
			if len(plugins) == 0 {
				fmt.Printf("No plugins found in %s\n", pluginsDir)
				return nil
			}

			fmt.Printf("Found %d plugin(s) in %s:\n\n", len(plugins), pluginsDir)
			for _, p := range plugins {
				fmt.Printf("  %s", p.Name)
				if p.Description != "" {
					fmt.Printf(" — %s", p.Description)
				}
				fmt.Println()
				if len(p.Skills) > 0 {
					fmt.Printf("    Skills: %d\n", len(p.Skills))
				}
				if len(p.MCPServers) > 0 {
					fmt.Printf("    MCP Servers: %d\n", len(p.MCPServers))
				}
				if p.Hooks != nil {
					hooks := len(p.Hooks.PreToolUse) + len(p.Hooks.PostToolUse)
					if hooks > 0 {
						fmt.Printf("    Hooks: %d\n", hooks)
					}
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&pluginsDir, "plugins-dir", "./plugins", "Directory containing plugin YAML files")

	return cmd
}

func cleanCmd() *cobra.Command {
	var dryRun bool
	var keepLogs int
	var all bool
	var yes bool

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Remove stale Copilot CLI session state and orphaned processes from past eval runs",
		Long: `Scans for orphaned hyoka-spawned Copilot processes (tagged with HYOKA_SESSION)
and stale ~/.copilot/session-state/ directories. Lists any found processes and
asks for confirmation before killing them. Session state accumulates over many
eval runs and can grow to gigabytes, so periodic cleanup is recommended.

By default only cleans hyoka-created sessions and processes. Use --all to also
clean Copilot CLI log files and non-hyoka session state.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			// Phase 1: Scan for orphaned hyoka processes (#70)
			procs, scanErr := clean.ScanHyokaProcesses(out)
			if scanErr != nil {
				fmt.Fprintf(out, "Warning: process scan: %v\n", scanErr)
			}

			killOrphans := false
			if len(procs) > 0 && !dryRun {
				if yes {
					killOrphans = true
				} else {
					fmt.Fprintf(out, "\nKill these %d process(es)? [y/N] ", len(procs))
					var answer string
					fmt.Fscanln(cmd.InOrStdin(), &answer)
					answer = strings.TrimSpace(strings.ToLower(answer))
					killOrphans = answer == "y" || answer == "yes"
				}
				if killOrphans {
					killed := clean.KillHyokaProcesses(procs, out)
					fmt.Fprintf(out, "Killed %d process(es)\n", killed)
				} else {
					fmt.Fprintln(out, "Skipped process cleanup.")
				}
			}

			// Phase 2: Session state and log cleanup
			result, err := clean.Run(clean.Options{
				DryRun:      dryRun,
				KeepLogs:    keepLogs,
				All:         all,
				KillOrphans: false, // already handled above
				Out:         out,
			})
			if err != nil {
				return err
			}

			if dryRun {
				fmt.Fprintf(out, "\nDry run: found %d session(s) to remove", result.SessionsFound)
				if len(procs) > 0 {
					fmt.Fprintf(out, ", %d process(es) to kill", len(procs))
				}
				fmt.Fprintln(out)
			} else {
				parts := []string{fmt.Sprintf("%d session(s)", result.SessionsRemoved)}
				if result.LogsRemoved > 0 {
					parts = append(parts, fmt.Sprintf("%d log(s)", result.LogsRemoved))
				}
				fmt.Fprintf(out, "\nCleaned %s — freed %s\n",
					strings.Join(parts, ", "), humanSize(result.BytesFreed))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be cleaned without deleting or killing")
	cmd.Flags().IntVar(&keepLogs, "keep-logs", 50, "Number of recent log files to keep (only with --all)")
	cmd.Flags().BoolVar(&all, "all", false, "Also clean Copilot CLI logs and non-hyoka session state")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt and kill orphaned processes automatically")

	return cmd
}

func humanSize(b int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case b >= gb:
		return fmt.Sprintf("%.1fGB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1fMB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1fKB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%dB", b)
	}
}

// openInBrowser opens the given file path in the default browser.
func openInBrowser(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", path)
	default:
		fmt.Printf("Open the report manually: %s\n", path)
		return
	}
	if err := cmd.Start(); err != nil {
		fmt.Printf("Could not open browser: %v\nOpen manually: %s\n", err, path)
	}
}

var validServices = validate.ValidServices
var validLanguages = validate.ValidLanguages
var validPlanes = validate.ValidPlanes
var validCategories = validate.ValidCategories
var validDifficulties = validate.ValidDifficulties

func newPromptCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new-prompt",
		Short: "Scaffold a new prompt file interactively",
		Long:  "Asks for service, language, plane, category, and difficulty, then generates a prompt file with populated frontmatter at the correct path.",
		RunE: func(cmd *cobra.Command, args []string) error {
			promptsDir := resolvePromptsDir(cmd)

			service := askChoice("Service", validServices)
			plane := askChoice("Plane", validPlanes)
			language := askChoice("Language", validLanguages)
			category := askChoice("Category", validCategories)
			difficulty := askChoice("Difficulty", validDifficulties)
			description := askFreeText("Description (what this prompt tests)")

			// Build the prompt ID
			planeAbbrev := "dp"
			if plane == "management-plane" {
				planeAbbrev = "mp"
			}

			// Ask for a slug to make the ID unique
			slug := askFreeText("Short slug for filename (e.g. 'list-blobs')")
			slug = strings.ReplaceAll(strings.TrimSpace(slug), " ", "-")
			slug = strings.ToLower(slug)

			id := fmt.Sprintf("%s-%s-%s-%s", service, planeAbbrev, language, slug)

			dir := filepath.Join(promptsDir, service, plane, language)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("creating directory: %w", err)
			}

			filename := slug + ".prompt.md"
			filePath := filepath.Join(dir, filename)

			if _, err := os.Stat(filePath); err == nil {
				return fmt.Errorf("file already exists: %s", filePath)
			}

			today := time.Now().Format("2006-01-02")

			content := fmt.Sprintf(`---
id: %s
service: %s
plane: %s
language: %s
category: %s
difficulty: %s
description: >
  %s
sdk_package: ""
doc_url: ""
tags: []
created: %s
author: ""
---

# TODO: Title — %s (%s)

## Prompt

TODO: Write your prompt here.

## Expected Coverage

The generated code should demonstrate:
- TODO: List key aspects to test

## Context

TODO: Why this prompt matters.
`, id, service, plane, language, category, difficulty, description, today, service, language)

			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("writing prompt file: %w", err)
			}

			fmt.Printf("\n✅ Created prompt file: %s\n", filePath)
			fmt.Printf("   Prompt ID: %s\n", id)
			fmt.Println("\nNext steps:")
			fmt.Println("  1. Edit the file to add your prompt text")
			fmt.Println("  2. Run: go run ./hyoka validate")
			return nil
		},
	}
}

// parseByteSize parses a human-readable byte size string (e.g., "1MB", "512KB", "1048576").
func parseByteSize(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	multipliers := map[string]int64{
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
	}
	for suffix, mult := range multipliers {
		if strings.HasSuffix(s, suffix) {
			numStr := strings.TrimSuffix(s, suffix)
			num, err := strconv.ParseFloat(strings.TrimSpace(numStr), 64)
			if err != nil {
				return 0, fmt.Errorf("invalid number %q", numStr)
			}
			return int64(num * float64(mult)), nil
		}
	}
	// Plain number (bytes)
	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q — use a number with optional KB/MB/GB suffix", s)
	}
	return num, nil
}

func askChoice(label string, options []string) string {
	fmt.Printf("\n%s:\n", label)
	for i, opt := range options {
		fmt.Printf("  %d) %s\n", i+1, opt)
	}
	for {
		fmt.Printf("Choose [1-%d]: ", len(options))
		var choice int
		_, err := fmt.Scanln(&choice)
		if err == nil && choice >= 1 && choice <= len(options) {
			return options[choice-1]
		}
		fmt.Println("Invalid choice, try again.")
	}
}

func askFreeText(label string) string {
	fmt.Printf("\n%s: ", label)
	var input string
	fmt.Scanln(&input)
	return strings.TrimSpace(input)
}
