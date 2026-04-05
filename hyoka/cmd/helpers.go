package cmd

import (
"fmt"
"log/slog"
"os"
"os/exec"
"path/filepath"
"runtime"
"strconv"
"strings"

"github.com/ronniegeraghty/hyoka/internal/config"
"github.com/spf13/cobra"
)

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
if configs[i].Generator != nil {
resolveSkills(configs[i].Generator.Skills)
}
if configs[i].Reviewer != nil {
resolveSkills(configs[i].Reviewer.Skills)
}
}
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
num, err := strconv.ParseInt(s, 10, 64)
if err != nil {
return 0, fmt.Errorf("invalid size %q \u2014 use a number with optional KB/MB/GB suffix", s)
}
return num, nil
}
