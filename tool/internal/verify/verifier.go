// Package verify provides functionality to verify generated code using Copilot.
package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	copilot "github.com/github/copilot-sdk/go"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/report"
	"github.com/ronniegeraghty/azure-sdk-prompts/tool/internal/utils"
)

// CopilotVerifier uses a separate Copilot session to verify generated code against requirements.
type CopilotVerifier struct {
	clientOpts       *copilot.ClientOptions
	model            string
	debug            bool
	skillDirectories []string
}

// NewCopilotVerifier creates a verifier that spawns its own Copilot client per verification.
func NewCopilotVerifier(clientOpts *copilot.ClientOptions, model string, debug bool) *CopilotVerifier {
	if model == "" {
		model = "claude-sonnet-4.5"
	}
	return &CopilotVerifier{clientOpts: clientOpts, model: model, debug: debug}
}

// SetSkillDirectories configures skill directories for the verification session.
func (v *CopilotVerifier) SetSkillDirectories(dirs []string) {
	v.skillDirectories = dirs
}

// minVerifyTimeout is the minimum time budget for verification, ensuring it
// doesn't get starved when the parent eval context is nearly exhausted.
const minVerifyTimeout = 2 * time.Minute

// Verify creates a separate Copilot session to evaluate whether generated code meets requirements.
func (v *CopilotVerifier) Verify(ctx context.Context, originalPrompt string, workDir string, evaluationCriteria string) (*report.VerifyResult, error) {
	// Ensure verification has adequate time even if the parent eval context
	// is nearly expired (e.g. generation consumed most of the shared timeout).
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining < minVerifyTimeout {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.Background(), minVerifyTimeout)
			defer cancel()
		}
	}

	generatedFiles, err := utils.ReadDirFiles(workDir)
	if err != nil {
		return nil, fmt.Errorf("reading generated files: %w", err)
	}
	if len(generatedFiles) == 0 {
		return &report.VerifyResult{
			Pass:      false,
			Reasoning: "No files were generated.",
			Summary:   "FAIL — no output files found",
		}, nil
	}

	verifyPrompt := buildVerifyPrompt(originalPrompt, generatedFiles, evaluationCriteria)

	// Create a fresh Copilot client for verification
	opts := *v.clientOpts
	opts.Cwd = workDir
	client := copilot.NewClient(&opts)
	if err := client.Start(ctx); err != nil {
		return nil, fmt.Errorf("starting verification client: %w", err)
	}
	defer func() {
		done := make(chan struct{})
		go func() { client.Stop(); close(done) }()
		select {
		case <-done:
		case <-time.After(30 * time.Second):
			client.ForceStop()
		}
	}()

	session, err := client.CreateSession(ctx, &copilot.SessionConfig{
		Model: v.model,
		SystemMessage: &copilot.SystemMessageConfig{
			Mode:    "append",
			Content: "You are a code verification judge. Respond with ONLY valid JSON. No markdown, no explanation.",
		},
		WorkingDirectory:    workDir,
		OnPermissionRequest: copilot.PermissionHandler.ApproveAll,
		SkillDirectories:    v.skillDirectories,
	})
	if err != nil {
		return nil, fmt.Errorf("creating verification session: %w", err)
	}
	defer func() {
		done := make(chan struct{})
		go func() { session.Disconnect(); close(done) }()
		select {
		case <-done:
		case <-time.After(15 * time.Second):
		}
	}()

	var assistantContent strings.Builder
	var mu sync.Mutex
	unsub := session.On(func(event copilot.SessionEvent) {
		if event.Type == copilot.SessionEventTypeAssistantMessage && event.Data.Content != nil {
			mu.Lock()
			assistantContent.WriteString(*event.Data.Content)
			mu.Unlock()
		}
	})
	defer unsub()

	_, err = session.SendAndWait(ctx, copilot.MessageOptions{
		Prompt: verifyPrompt,
	})
	if err != nil {
		return nil, fmt.Errorf("verification session send: %w", err)
	}

	mu.Lock()
	responseText := assistantContent.String()
	mu.Unlock()

	return parseVerifyResponse(responseText)
}

func buildVerifyPrompt(originalPrompt string, generatedFiles map[string]string, evaluationCriteria string) string {
	var b strings.Builder

	b.WriteString("You are a code verification judge. Evaluate whether the generated code meets the requirements.\n\n")

	b.WriteString("## Original Prompt\n\n")
	b.WriteString(originalPrompt)
	b.WriteString("\n\n")

	if evaluationCriteria != "" {
		b.WriteString("## Evaluation Criteria\n\n")
		b.WriteString(evaluationCriteria)
		b.WriteString("\n\n")
	}

	b.WriteString("## Generated Code\n\n")
	for name, content := range generatedFiles {
		fmt.Fprintf(&b, "### %s\n```\n%s\n```\n\n", name, content)
	}

	b.WriteString(`## Verification Checks

Determine if this code meets the prompt's requirements:
1. Does it address the main task described in the prompt?
2. Does it use the correct SDK/packages mentioned?
3. Is it syntactically valid and likely to compile/run?
4. Does it handle the key scenarios mentioned?

## Output Format

Respond with ONLY a JSON object:
{"pass": true/false, "reasoning": "Detailed explanation of what was checked and why it passes or fails", "summary": "One-line summary"}
`)

	return b.String()
}

func parseVerifyResponse(text string) (*report.VerifyResult, error) {
	jsonStr := utils.ExtractJSON(text)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in verification response: %.200s", text)
	}

	var result report.VerifyResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("parsing verification JSON: %w (response: %.200s)", err, jsonStr)
	}
	return &result, nil
}
