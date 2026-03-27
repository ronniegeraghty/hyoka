package progress

// EventType classifies a progress event.
type EventType int

const (
	EventStarting      EventType = iota // Eval starting, waiting for session
	EventSendingPrompt                  // Sending prompt to Copilot
	EventReasoning                      // LLM is reasoning
	EventToolStart                      // Tool call initiated
	EventToolComplete                   // Tool call finished
	EventWritingFile                    // File write tool call
	EventWaiting                        // Waiting between events
	EventPhaseChange                    // Transition between eval phases
	EventPassed                         // Eval passed
	EventFailed                         // Eval failed
	EventError                          // Eval errored
)

// Phase identifies which stage an eval is in.
type Phase string

const (
	PhaseGenerating Phase = "generating"
	PhaseVerifying  Phase = "verifying"
	PhaseReviewing  Phase = "reviewing"
)

// ProgressEvent carries status from the eval engine or Copilot session to the display.
type ProgressEvent struct {
	EvalID      string    // Unique eval identifier (promptID/configName)
	PromptID    string    // Prompt ID
	ConfigName  string    // Config name
	Type        EventType // What happened
	Message     string    // Human-readable activity description
	FileCount   int       // Generated file count (for completion events)
	Phase       Phase     // Current phase (for EventPhaseChange)
	ReviewScore int       // Review score out of 10 (for EventPassed)
}

// ProgressFunc receives progress events from evaluators.
type ProgressFunc func(ProgressEvent)

// Reporter is implemented by evaluators that support live progress updates.
type Reporter interface {
	SetProgressFunc(fn ProgressFunc)
}
