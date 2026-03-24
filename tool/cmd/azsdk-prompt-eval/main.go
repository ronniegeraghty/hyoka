package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	copilot "github.com/github/copilot-sdk/go"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/checkenv"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/config"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/eval"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/prompt"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/rerender"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/report"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/review"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/trends"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/validate"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/verify"
	"github.com/spf13/cobra"
)

var version = "0.6.0"

func main() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "azsdk-prompt-eval",
		Short: "Azure SDK Prompt Evaluation Tool — test AI agent code generation quality",
		Long:  "A tool for evaluating AI agent code generation quality by running prompts through the Copilot SDK, verifying builds, and generating reports.",
	}

	root.AddCommand(runCmd())
	root.AddCommand(listCmd())
	root.AddCommand(configsCmd())
	root.AddCommand(versionCmd())

	root.AddCommand(validateCmd())
	root.AddCommand(checkEnvCmd())
	root.AddCommand(trendsCmd())
	root.AddCommand(reportCmd())
	root.AddCommand(newPromptCmd())

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
	configName   string
	configFile   string
	configDir    string
	workers         int
	timeout         int
	generateTimeout int
	verifyTimeout   int
	reviewTimeout   int
	model           string
	output       string
	progressMode string
	skipTests    bool
	skipReview   bool
	skipTrends   bool
	verifyBuild  bool
	debug        bool
	dryRun       bool
	useStub      bool
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
	cmd.Flags().IntVar(&f.workers, "workers", 4, "Parallel evaluation workers")
	cmd.Flags().IntVar(&f.timeout, "timeout", 600, "Per-prompt generation timeout in seconds (deprecated: use --generate-timeout)")
	cmd.Flags().IntVar(&f.generateTimeout, "generate-timeout", 0, "Generation phase timeout in seconds (default: --timeout value or 600)")
	cmd.Flags().IntVar(&f.verifyTimeout, "verify-timeout", 300, "Verification phase timeout in seconds")
	cmd.Flags().IntVar(&f.reviewTimeout, "review-timeout", 300, "Review phase timeout in seconds")
	cmd.Flags().StringVar(&f.model, "model", "", "Override model for all configs")
	cmd.Flags().StringVar(&f.output, "output", "./reports", "Report output directory")
	cmd.Flags().BoolVar(&f.skipTests, "skip-tests", false, "Skip test generation")
	cmd.Flags().BoolVar(&f.skipReview, "skip-review", false, "Skip code review")
	cmd.Flags().BoolVar(&f.verifyBuild, "verify-build", false, "Also run build verification (in addition to Copilot verification)")
	cmd.Flags().BoolVar(&f.debug, "debug", false, "Verbose output")
	cmd.Flags().StringVar(&f.progressMode, "progress", "auto", "Progress display mode: auto, live, log, off")
	cmd.Flags().BoolVar(&f.dryRun, "dry-run", false, "List matching prompts without running")
	cmd.Flags().BoolVar(&f.useStub, "stub", false, "Use stub evaluator (no Copilot SDK)")
	cmd.Flags().BoolVar(&f.skipTrends, "skip-trends", false, "Skip automatic trend analysis after run")
}

// resolveSkillsDirs finds the skills directory relative to the prompts directory.
// It checks multiple candidate paths to work from both repo root and tool/ directory.
// resolveSkillsDirs resolves skill directories for generator and reviewer sessions.
// It looks for skills/generator/ and skills/reviewer/ subdirectories.
// Falls back to the parent skills/ directory for both if subdirs don't exist.
func resolveSkillsDirs(promptsDir string) (generatorDirs, reviewerDirs []string) {
	var baseDir string
	for _, candidate := range []string{
		filepath.Join(filepath.Dir(promptsDir), "skills"),
		"./skills",
		"../skills",
	} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(candidate)
			baseDir = abs
			break
		}
	}
	if baseDir == "" {
		return nil, nil
	}

	genDir := filepath.Join(baseDir, "generator")
	revDir := filepath.Join(baseDir, "reviewer")

	if info, err := os.Stat(genDir); err == nil && info.IsDir() {
		generatorDirs = []string{genDir}
	}
	if info, err := os.Stat(revDir); err == nil && info.IsDir() {
		reviewerDirs = []string{revDir}
	}

	// If neither subdir exists, fall back to the base skills dir for both
	if generatorDirs == nil && reviewerDirs == nil {
		return []string{baseDir}, []string{baseDir}
	}
	return generatorDirs, reviewerDirs
}

// resolveConfigSkillDirs resolves relative skill_directories in loaded configs
// to absolute paths so they work regardless of which directory the tool is invoked from.
func resolveConfigSkillDirs(configs []config.ToolConfig, promptsDir string) {
	resolve := func(dirs []string) []string {
		resolved := make([]string, 0, len(dirs))
		for _, dir := range dirs {
			if filepath.IsAbs(dir) {
				resolved = append(resolved, dir)
				continue
			}
			candidates := []string{
				dir,
				filepath.Join(filepath.Dir(promptsDir), dir),
			}
			found := false
			for _, c := range candidates {
				if info, err := os.Stat(c); err == nil && info.IsDir() {
					abs, _ := filepath.Abs(c)
					resolved = append(resolved, abs)
					found = true
					break
				}
			}
			if !found {
				abs, _ := filepath.Abs(dir)
				resolved = append(resolved, abs)
			}
		}
		return resolved
	}

	for i := range configs {
		configs[i].SkillDirectories = resolve(configs[i].SkillDirectories)
		configs[i].GeneratorSkillDirectories = resolve(configs[i].GeneratorSkillDirectories)
		configs[i].ReviewerSkillDirectories = resolve(configs[i].ReviewerSkillDirectories)
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
	return prompt.Filter{
		Service:  f.service,
		Plane:    f.plane,
		Language: f.language,
		Category: f.category,
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

			// Override model if specified via CLI flag
			if f.model != "" {
				for i := range configs {
					configs[i].Model = f.model
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
				fmt.Println("No prompts matched the given filters.")
				return nil
			}

			fmt.Printf("Found %d prompt(s), %d config(s) → %d evaluation(s)\n",
				len(filtered), len(configs), len(filtered)*len(configs))

			// Select evaluator, verifier, and reviewer
			var evaluator eval.CopilotEvaluator
			var reviewer review.Reviewer
			var verifier eval.Verifier
			var panelReviewer *review.PanelReviewer

			if f.useStub {
				fmt.Println("Using stub evaluator (--stub flag)")
				evaluator = &eval.StubEvaluator{}
				reviewer = &review.StubReviewer{}
				verifier = &eval.StubVerifier{}
			} else {
				// Try to create a real Copilot SDK evaluator
				sdkEval := eval.NewCopilotSDKEvaluator(eval.CopilotEvalOptions{
					Debug: f.debug,
				})
				evaluator = sdkEval

				// Verify Copilot CLI is available
				client := copilot.NewClient(nil)
				if err := client.Start(context.Background()); err != nil {
					fmt.Printf("⚠️  Copilot SDK unavailable (%v), falling back to stub evaluator\n", err)
					evaluator = &eval.StubEvaluator{}
					reviewer = &review.StubReviewer{}
					verifier = &eval.StubVerifier{}
				} else {
					client.Stop()
					fmt.Println("Using Copilot SDK evaluator")

					// Create a real CopilotVerifier with its own client options
					clientOpts := &copilot.ClientOptions{}
					if f.debug {
						clientOpts.LogLevel = "debug"
					}
					copilotVerifier := verify.NewCopilotVerifier(clientOpts, f.model, f.debug)

					// Wire skills directories separately for generator, reviewer, and verifier
					generatorSkillsDirs, reviewerSkillsDirs := resolveSkillsDirs(f.prompts)
					if len(reviewerSkillsDirs) > 0 {
						copilotVerifier.SetSkillDirectories(reviewerSkillsDirs)
					}
					verifier = copilotVerifier

					// Create reviewer(s) — use panel if reviewer_models list configured
					var reviewerModels []string
					for _, c := range configs {
						if models := c.EffectiveReviewerModels(); len(models) > 0 {
							reviewerModels = models
							break
						}
					}

					if len(reviewerModels) > 1 {
						// Multi-model panel
						panelReviewer = review.NewPanelReviewer(clientOpts, reviewerModels, f.debug)
						if len(reviewerSkillsDirs) > 0 {
							panelReviewer.SetSkillDirectories(reviewerSkillsDirs)
						}
						fmt.Printf("Using review panel: %v\n", reviewerModels)
					} else {
						// Single reviewer (backward compat)
						reviewClient := copilot.NewClient(clientOpts)
						if err := reviewClient.Start(context.Background()); err != nil {
							fmt.Printf("⚠️  Could not start reviewer client: %v, reviews will be skipped\n", err)
						} else {
							reviewerModel := f.model
							if len(reviewerModels) == 1 {
								reviewerModel = reviewerModels[0]
							}
							copilotReviewer := review.NewCopilotReviewer(reviewClient, reviewerModel)
							if len(reviewerSkillsDirs) > 0 {
								copilotReviewer.SetSkillDirectories(reviewerSkillsDirs)
							}
							reviewer = copilotReviewer
							defer reviewClient.Stop()
						}
					}

					// Override generator config's skill directories with generator-specific ones
					_ = generatorSkillsDirs // used below when config skill_directories are resolved
				}
			}

			if f.skipReview {
				reviewer = nil
			}

			// Create and run engine
			engine := eval.NewEngineWithReviewer(evaluator, verifier, reviewer, eval.EngineOptions{
				Workers:         f.workers,
				Timeout:         time.Duration(f.timeout) * time.Second,
				GenerateTimeout: time.Duration(f.generateTimeout) * time.Second,
				VerifyTimeout:   time.Duration(f.verifyTimeout) * time.Second,
				ReviewTimeout:   time.Duration(f.reviewTimeout) * time.Second,
				OutputDir:       f.output,
				SkipTests:       f.skipTests,
				SkipReview:      f.skipReview,
				VerifyBuild:     f.verifyBuild,
				Debug:           f.debug,
				DryRun:          f.dryRun,
				ProgressMode:    f.progressMode,
			})
			if panelReviewer != nil && !f.skipReview {
				engine.SetPanelReviewer(panelReviewer)
			}

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
					fmt.Printf("⚠️  Trend generation failed: %v\n", err)
				} else if tr.TotalRuns > 0 {
					fmt.Println("🤖 Running AI-powered trend analysis...")
					analysis, aErr := trends.AnalyzeTrends(context.Background(), tr)
					if aErr != nil {
						fmt.Printf("⚠️  AI analysis failed: %v (continuing without analysis)\n", aErr)
					} else {
						tr.Analysis = analysis
						fmt.Println("\n--- AI Analysis ---")
						fmt.Println(analysis)
						fmt.Println("-------------------")

						// Re-write summary HTML with AI analysis included (Issue 7)
						summary.Analysis = analysis
						if _, err := report.WriteSummaryHTML(summary, f.output); err != nil {
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
				fmt.Printf("  %-30s %s/%s/%s [%s]\n", p.ID, p.Service, p.Plane, p.Language, p.Category)
				if p.Description != "" {
					fmt.Printf("  %-30s %s\n", "", p.Description)
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
				fmt.Printf("  %-20s %s (model: %s)\n", c.Name, c.Description, c.Model)
				if len(c.MCPServers) > 0 {
					fmt.Printf("  %-20s MCP servers: ", "")
					var names []string
					for name := range c.MCPServers {
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
				return fmt.Errorf("validation: %w", err)
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
			fmt.Printf("azsdk-prompt-eval version %s\n", version)
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
			fmt.Println("  2. Run: go run ./tool/cmd/azsdk-prompt-eval validate")
			return nil
		},
	}
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
