package cmd

import (
"encoding/json"
"fmt"

"github.com/ronniegeraghty/hyoka/internal/prompt"
"github.com/spf13/cobra"
)

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
