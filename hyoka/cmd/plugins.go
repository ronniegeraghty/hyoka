package cmd

import (
"fmt"

"github.com/ronniegeraghty/hyoka/internal/plugin"
"github.com/spf13/cobra"
)

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
fmt.Printf(" \u2014 %s", p.Description)
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
