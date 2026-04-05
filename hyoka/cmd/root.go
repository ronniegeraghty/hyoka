package cmd

import (
"github.com/ronniegeraghty/hyoka/internal/logging"
"github.com/spf13/cobra"
)

// Version is the CLI version, set at build time or defaulting here.
var Version = "0.2.0"

// Execute creates the root command tree and runs it.
func Execute() error {
return rootCmd().Execute()
}

func rootCmd() *cobra.Command {
var logLevel, logFile string

root := &cobra.Command{
Use:   "hyoka",
Short: "Azure SDK Prompt Evaluation Tool \u2014 test AI agent code generation quality",
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
