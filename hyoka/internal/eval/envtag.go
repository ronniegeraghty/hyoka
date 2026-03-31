package eval

// Environment variable names injected into SDK-spawned Copilot processes
// so hyoka can identify and clean up its own child processes (#70).
const (
	// EnvHyokaSession is set to "true" on every SDK-spawned process.
	EnvHyokaSession = "HYOKA_SESSION"
	// EnvHyokaPromptID carries the prompt ID for the current evaluation.
	EnvHyokaPromptID = "HYOKA_PROMPT_ID"
	// EnvHyokaConfig carries the config name for the current evaluation.
	EnvHyokaConfig = "HYOKA_CONFIG"
)

// HyokaBaseEnv returns the baseline environment entries that tag a process
// as hyoka-managed. Callers can append prompt/config-specific entries.
func HyokaBaseEnv() []string {
	return []string{EnvHyokaSession + "=true"}
}

// HyokaEvalEnv returns environment entries that tag a process with
// the session marker plus prompt and config metadata.
func HyokaEvalEnv(promptID, configName string) []string {
	env := HyokaBaseEnv()
	if promptID != "" {
		env = append(env, EnvHyokaPromptID+"="+promptID)
	}
	if configName != "" {
		env = append(env, EnvHyokaConfig+"="+configName)
	}
	return env
}
