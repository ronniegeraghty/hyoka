package cmd

import (
"github.com/ronniegeraghty/hyoka/internal/checkenv"
"github.com/spf13/cobra"
)

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
