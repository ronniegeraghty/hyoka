package cmd

import (
"fmt"
"strings"

"github.com/ronniegeraghty/hyoka/internal/config"
"github.com/spf13/cobra"
)

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
