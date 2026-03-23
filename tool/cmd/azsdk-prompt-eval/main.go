package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	copilot "github.com/github/copilot-sdk/go"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/checkenv"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/config"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/eval"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/prompt"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/rerender"
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
		if info, err := os.Stat(c); err == nil && info.IsDir() || err == nil {
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
		"./configs/all.yaml", "../configs/all.yaml",
		"./configs.yaml", "../configs.yaml",
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
	workers      int
	timeout      int
	model        string
	output       string
	progressMode string
	skipTests    bool
	skipReview   bool
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
	cmd.Flags().StringVar(&f.configFile, "config-file", "./configs/all.yaml", "Path to configuration YAML")
	cmd.Flags().IntVar(&f.workers, "workers", 4, "Parallel evaluation workers")
	cmd.Flags().IntVar(&f.timeout, "timeout", 300, "Per-prompt timeout in seconds")
	cmd.Flags().StringVar(&f.model, "model", "", "Override model for all configs")
	cmd.Flags().StringVar(&f.output, "output", "./reports", "Report output directory")
	cmd.Flags().BoolVar(&f.skipTests, "skip-tests", false, "Skip test generation")
	cmd.Flags().BoolVar(&f.skipReview, "skip-review", false, "Skip code review")
	cmd.Flags().BoolVar(&f.verifyBuild, "verify-build", false, "Also run build verification (in addition to Copilot verification)")
	cmd.Flags().BoolVar(&f.debug, "debug", false, "Verbose output")
	cmd.Flags().StringVar(&f.progressMode, "progress", "auto", "Progress display mode: auto, live, log, off")
	cmd.Flags().BoolVar(&f.dryRun, "dry-run", false, "List matching prompts without running")
	cmd.Flags().BoolVar(&f.useStub, "stub", false, "Use stub evaluator (no Copilot SDK)")
}

// resolveSkillsDirs finds the skills directory relative to the prompts directory.
func resolveSkillsDirs(promptsDir string) []string {
	for _, candidate := range []string{
		filepath.Join(filepath.Dir(promptsDir), "skills"),
		"./skills",
		"../skills",
	} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(candidate)
			return []string{abs}
		}
	}
	return nil
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
			f.configFile = resolveConfigFile(cmd)
			f.output = resolveOutputDir(cmd)

			// Load config
			cfgFile, err := config.Load(f.configFile)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
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

			// Override model if specified
			if f.model != "" {
				for i := range configs {
					configs[i].Model = f.model
				}
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

					// Wire skills directory for reviewer/verifier sessions
					skillsDirs := resolveSkillsDirs(f.prompts)
					if len(skillsDirs) > 0 {
						copilotVerifier.SetSkillDirectories(skillsDirs)
					}
					verifier = copilotVerifier
				}
			}

			if f.skipReview {
				reviewer = nil
			}

			// Create and run engine
			engine := eval.NewEngineWithReviewer(evaluator, verifier, reviewer, eval.EngineOptions{
				Workers:      f.workers,
				Timeout:      time.Duration(f.timeout) * time.Second,
				OutputDir:    f.output,
				SkipTests:    f.skipTests,
				SkipReview:   f.skipReview,
				VerifyBuild:  f.verifyBuild,
				Debug:        f.debug,
				DryRun:       f.dryRun,
				ProgressMode: f.progressMode,
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

			return nil
		},
	}

	addFilterFlags(cmd, f)
	return cmd
}

func listCmd() *cobra.Command {
	f := &runFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List matching prompts",
		Long:  "List prompts matching the given filters (dry-run equivalent).",
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
	return cmd
}

func configsCmd() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "configs",
		Short: "List available configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile = resolveConfigFile(cmd)

			cfgFile, err := config.Load(configFile)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
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

	cmd.Flags().StringVar(&configFile, "config-file", "./configs/all.yaml", "Path to configuration YAML")
	return cmd
}

func validateCmd() *cobra.Command {
	var promptsDir string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate prompt frontmatter",
		Long:  "Validate all prompt files against schema rules and naming conventions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			promptsDir = resolvePromptsDir(cmd)

			result, err := validate.Validate(promptsDir)
			if err != nil {
				return fmt.Errorf("validation: %w", err)
			}

			fmt.Print(validate.FormatResult(result))

			if !result.OK() {
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
		Use:   "check-env",
		Short: "Check for required language toolchains and tools",
		Long:  "Tests if language toolchains (dotnet, python, go, node, java, rust, cargo, cmake, etc.), Copilot CLI, and MCP prerequisites are installed.",
		Run: func(cmd *cobra.Command, args []string) {
			checkenv.Run()
		},
	}
}

func trendsCmd() *cobra.Command {
	var promptID, service, language, reportsDir, output string
	var analyze bool

	cmd := &cobra.Command{
		Use:   "trends",
		Short: "Generate historical trend reports with time-series performance data",
		Long:  "Scans all past runs in reports/ directory and generates a trend report with pass-rate timelines, duration trends, config comparisons, and regression detection. Use --analyze for AI-powered insights.",
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

			return nil
		},
	}

	cmd.Flags().StringVar(&promptID, "prompt-id", "", "Filter trends by prompt ID")
	cmd.Flags().StringVar(&service, "service", "", "Filter trends by Azure service")
	cmd.Flags().StringVar(&language, "language", "", "Filter trends by programming language")
	cmd.Flags().StringVar(&reportsDir, "reports-dir", "./reports", "Directory containing past evaluation reports")
	cmd.Flags().StringVar(&output, "output", "./reports/trends", "Output directory for trend reports")
	cmd.Flags().BoolVar(&analyze, "analyze", false, "Run Copilot-powered AI analysis of trends")

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
