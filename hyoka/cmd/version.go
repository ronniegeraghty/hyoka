package cmd

import (
"fmt"

"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
return &cobra.Command{
Use:   "version",
Short: "Print version",
Run: func(cmd *cobra.Command, args []string) {
fmt.Printf("hyoka version %s\n", Version)
},
}
}
