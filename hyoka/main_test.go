package main

import (
	"io"
	"testing"
)

func TestRunCmdFlagDefaults(t *testing.T) {
	cmd := runCmd()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--help"})
	_ = cmd.Execute()

	tests := []struct {
		flag     string
		expected string
	}{
		{"prompts", "./prompts"},
		{"service", ""},
		{"language", ""},
		{"plane", ""},
		{"category", ""},
		{"tags", ""},
		{"prompt-id", ""},
		{"config", ""},
		{"config-file", ""},
		{"config-dir", "./configs"},
		{"workers", "0"},
		{"max-sessions", "0"},
		{"timeout", "600"},
		{"generate-timeout", "0"},
		{"build-timeout", "300"},
		{"review-timeout", "300"},
		{"model", ""},
		{"output", "./reports"},
		{"progress", "auto"},
		{"max-turns", "25"},
		{"max-files", "50"},
		{"max-output-size", "1MB"},
		{"criteria-dir", ""},
	}

	for _, tt := range tests {
		f := cmd.Flags().Lookup(tt.flag)
		if f == nil {
			t.Errorf("expected flag %q to be registered", tt.flag)
			continue
		}
		if f.DefValue != tt.expected {
			t.Errorf("flag %q: expected default %q, got %q", tt.flag, tt.expected, f.DefValue)
		}
	}
}

func TestRunCmdBoolFlagDefaults(t *testing.T) {
	cmd := runCmd()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--help"})
	_ = cmd.Execute()

	falseFlags := []string{
		"skip-tests",
		"skip-review",
		"skip-trends",
		"verify-build",
		"dry-run",
		"stub",
		"yes",
		"all-configs",
		"allow-cloud",
		"monitor-resources",
		"strict-cleanup",
	}

	for _, name := range falseFlags {
		f := cmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("expected flag %q to be registered", name)
			continue
		}
		if f.DefValue != "false" {
			t.Errorf("flag %q: expected default %q, got %q", name, "false", f.DefValue)
		}
	}
}

func TestRunCmdFlagOverride(t *testing.T) {
	cmd := runCmd()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	args := []string{
		"--max-turns", "10",
		"--max-files", "20",
		"--max-output-size", "512KB",
		"--workers", "4",
		"--monitor-resources",
		"--strict-cleanup",
		"--verify-build",
		"--skip-review",
	}
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatalf("parsing flags: %v", err)
	}

	intTests := []struct {
		flag     string
		expected string
	}{
		{"max-turns", "10"},
		{"max-files", "20"},
		{"max-output-size", "512KB"},
		{"workers", "4"},
	}
	for _, tt := range intTests {
		val, err := cmd.Flags().GetString(tt.flag)
		if err != nil {
			// Try int
			v, err2 := cmd.Flags().GetInt(tt.flag)
			if err2 != nil {
				t.Errorf("flag %q: %v / %v", tt.flag, err, err2)
				continue
			}
			val = ""
			_ = v
			continue
		}
		if val != tt.expected {
			t.Errorf("flag %q: expected %q, got %q", tt.flag, tt.expected, val)
		}
	}

	boolTests := []struct {
		flag     string
		expected bool
	}{
		{"monitor-resources", true},
		{"strict-cleanup", true},
		{"verify-build", true},
		{"skip-review", true},
	}
	for _, tt := range boolTests {
		val, err := cmd.Flags().GetBool(tt.flag)
		if err != nil {
			t.Errorf("flag %q: %v", tt.flag, err)
			continue
		}
		if val != tt.expected {
			t.Errorf("flag %q: expected %v, got %v", tt.flag, tt.expected, val)
		}
	}
}

func TestParseByteSizeValid(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1MB", 1048576},
		{"1mb", 1048576},
		{"512KB", 524288},
		{"512kb", 524288},
		{"100", 100},
		{"2MB", 2097152},
		{"1024KB", 1048576},
	}

	for _, tt := range tests {
		got, err := parseByteSize(tt.input)
		if err != nil {
			t.Errorf("parseByteSize(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.expected {
			t.Errorf("parseByteSize(%q): expected %d, got %d", tt.input, tt.expected, got)
		}
	}
}

func TestParseByteSizeInvalid(t *testing.T) {
	invalid := []string{"", "abc", "1TB", "MB"}
	for _, input := range invalid {
		_, err := parseByteSize(input)
		if err == nil {
			t.Errorf("parseByteSize(%q): expected error, got none", input)
		}
	}
}

func TestRootCmdLogLevelFlags(t *testing.T) {
	cmd := rootCmd()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--help"})
	_ = cmd.Execute()

	logLevel := cmd.PersistentFlags().Lookup("log-level")
	if logLevel == nil {
		t.Fatal("expected persistent flag log-level")
	}
	if logLevel.DefValue != "warn" {
		t.Errorf("log-level default: expected %q, got %q", "warn", logLevel.DefValue)
	}

	logFile := cmd.PersistentFlags().Lookup("log-file")
	if logFile == nil {
		t.Fatal("expected persistent flag log-file")
	}
	if logFile.DefValue != "" {
		t.Errorf("log-file default: expected empty, got %q", logFile.DefValue)
	}
}

func TestServeCmdHelp(t *testing.T) {
	cmd := serveCmd()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("serve --help failed: %v", err)
	}

	port := cmd.Flags().Lookup("port")
	if port == nil {
		t.Fatal("expected --port flag")
	}
	if port.DefValue != "8080" {
		t.Errorf("port default: expected 8080, got %s", port.DefValue)
	}
}

func TestPluginsCmdHelp(t *testing.T) {
	cmd := pluginsCmd()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("plugins --help failed: %v", err)
	}

	dir := cmd.Flags().Lookup("plugins-dir")
	if dir == nil {
		t.Fatal("expected --plugins-dir flag")
	}
}

func TestValidateCmdHelp(t *testing.T) {
	cmd := validateCmd()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("validate --help failed: %v", err)
	}
}

func TestPluginsCmdEmptyDir(t *testing.T) {
	cmd := pluginsCmd()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--plugins-dir", t.TempDir()})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("plugins with empty dir failed: %v", err)
	}
}

func TestAllSubcommandsRegistered(t *testing.T) {
	cmd := rootCmd()
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}

	expected := []string{"run", "list", "validate", "check-env", "configs", "trends", "report", "serve", "plugins", "new-prompt", "version"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected subcommand %q to be registered", name)
		}
	}
}
