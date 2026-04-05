package cmd

import (
"github.com/ronniegeraghty/hyoka/internal/serve"
"github.com/spf13/cobra"
)

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
