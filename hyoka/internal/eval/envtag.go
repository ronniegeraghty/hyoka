package eval

import "os"

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

// HyokaBaseEnv returns the current process environment with the hyoka
// session marker appended. The full environment is inherited so spawned
// processes retain PATH and other required variables.
func HyokaBaseEnv() []string {
	return append(os.Environ(), EnvHyokaSession+"=true")
}

// HyokaEvalEnv returns the current process environment with the session
// marker plus prompt and config metadata appended.
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
