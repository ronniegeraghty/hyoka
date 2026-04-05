package graders

import (
"testing"
)

func TestParseValidConfig(t *testing.T) {
yaml := `
graders:
  - kind: file
    name: "main_exists"
    config:
      path: "main.py"
      must_exist: true
    weight: 1.0
    gate: true
    when:
      language: python

  - kind: program
    name: "builds"
    config:
      command: "python"
      args: ["-m", "py_compile", "main.py"]
      timeout: 30
    weight: 1.0
    gate: true

  - kind: prompt
    name: "review_opus"
    config:
      model: "claude-opus-4.6"
      rubric: "Evaluate correctness."
    weight: 0.7

  - kind: prompt
    name: "review_sonnet"
    config:
      model: "claude-sonnet-4.5"
      rubric: "Evaluate correctness."
    weight: 0.3

  - kind: behavior
    name: "tool_usage"
    config:
      required_tools: ["azure-mcp"]
      forbidden_tools: ["rm"]
      max_turns: 25
    weight: 1.0

  - kind: action_sequence
    name: "read_before_write"
    config:
      expected_actions: ["read_file", "edit_file"]
    weight: 0.5

  - kind: tool_constraint
    name: "bounded_tools"
    config:
      required: ["azure-mcp"]
      forbidden: ["dangerous"]
      min_calls: 1
      max_calls: 50
    weight: 0.5
`
gcf, err := Parse([]byte(yaml))
if err != nil {
t.Fatalf("Parse failed: %v", err)
}
if len(gcf.Graders) != 7 {
t.Fatalf("expected 7 graders, got %d", len(gcf.Graders))
}

// Verify file grader
g := gcf.Graders[0]
if g.Kind != KindFile || g.Name != "main_exists" || !g.Gate {
t.Errorf("file grader: unexpected values: kind=%s name=%s gate=%v", g.Kind, g.Name, g.Gate)
}
cfg, err := g.DecodeConfig()
if err != nil {
t.Fatalf("DecodeConfig file: %v", err)
}
fc := cfg.(*FileConfig)
if fc.Path != "main.py" {
t.Errorf("file config path: expected main.py, got %s", fc.Path)
}
if fc.MustExist == nil || !*fc.MustExist {
t.Error("file config must_exist: expected true")
}

// Verify prompt grader
g = gcf.Graders[2]
cfg, err = g.DecodeConfig()
if err != nil {
t.Fatalf("DecodeConfig prompt: %v", err)
}
pc := cfg.(*PromptConfig)
if pc.Model != "claude-opus-4.6" {
t.Errorf("prompt config model: expected claude-opus-4.6, got %s", pc.Model)
}

// Verify program grader
g = gcf.Graders[1]
cfg, err = g.DecodeConfig()
if err != nil {
t.Fatalf("DecodeConfig program: %v", err)
}
pgc := cfg.(*ProgramConfig)
if pgc.Command != "python" || len(pgc.Args) != 3 {
t.Errorf("program config: command=%s args=%v", pgc.Command, pgc.Args)
}

// Verify behavior grader
g = gcf.Graders[4]
cfg, err = g.DecodeConfig()
if err != nil {
t.Fatalf("DecodeConfig behavior: %v", err)
}
bc := cfg.(*BehaviorConfig)
if len(bc.RequiredTools) != 1 || bc.RequiredTools[0] != "azure-mcp" {
t.Errorf("behavior required_tools: %v", bc.RequiredTools)
}
if bc.MaxTurns != 25 {
t.Errorf("behavior max_turns: expected 25, got %d", bc.MaxTurns)
}

// Verify action_sequence grader
g = gcf.Graders[5]
cfg, err = g.DecodeConfig()
if err != nil {
t.Fatalf("DecodeConfig action_sequence: %v", err)
}
asc := cfg.(*ActionSequenceConfig)
if len(asc.ExpectedActions) != 2 {
t.Errorf("action_sequence expected_actions: %v", asc.ExpectedActions)
}

// Verify tool_constraint grader
g = gcf.Graders[6]
cfg, err = g.DecodeConfig()
if err != nil {
t.Fatalf("DecodeConfig tool_constraint: %v", err)
}
tc := cfg.(*ToolConstraintConfig)
if tc.MinCalls != 1 || tc.MaxCalls != 50 {
t.Errorf("tool_constraint: min=%d max=%d", tc.MinCalls, tc.MaxCalls)
}
}

func TestParseRejectsDuplicateNames(t *testing.T) {
yaml := `
graders:
  - kind: file
    name: "dupe"
    config:
      path: "a.py"
  - kind: file
    name: "dupe"
    config:
      path: "b.py"
`
_, err := Parse([]byte(yaml))
if err == nil {
t.Fatal("expected error for duplicate names")
}
}

func TestParseRejectsUnknownKind(t *testing.T) {
yaml := `
graders:
  - kind: unknown_kind
    name: "bad"
    config:
      path: "a.py"
`
_, err := Parse([]byte(yaml))
if err == nil {
t.Fatal("expected error for unknown kind")
}
}

func TestParseRejectsInvalidWeight(t *testing.T) {
yaml := `
graders:
  - kind: file
    name: "bad_weight"
    config:
      path: "a.py"
    weight: 1.5
`
_, err := Parse([]byte(yaml))
if err == nil {
t.Fatal("expected error for weight > 1.0")
}
}

func TestWhenMapMatches(t *testing.T) {
tests := []struct {
name  string
when  WhenMap
props map[string]string
want  bool
}{
{"empty when matches all", WhenMap{}, map[string]string{"language": "python"}, true},
{"nil when matches all", nil, map[string]string{"language": "python"}, true},
{"exact match", WhenMap{"language": "python"}, map[string]string{"language": "python"}, true},
{"case insensitive", WhenMap{"language": "Python"}, map[string]string{"language": "python"}, true},
{"missing property", WhenMap{"language": "python"}, map[string]string{}, false},
{"wrong value", WhenMap{"language": "python"}, map[string]string{"language": "go"}, false},
{"multi match", WhenMap{"language": "python", "service": "kv"}, map[string]string{"language": "python", "service": "kv"}, true},
{"partial match fails", WhenMap{"language": "python", "service": "kv"}, map[string]string{"language": "python"}, false},
}
for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
if got := tt.when.Matches(tt.props); got != tt.want {
t.Errorf("WhenMap.Matches() = %v, want %v", got, tt.want)
}
})
}
}

func TestApplicableGraders(t *testing.T) {
graders := []GraderConfig{
{Kind: KindFile, Name: "py_only", When: WhenMap{"language": "python"}},
{Kind: KindFile, Name: "all", When: nil},
{Kind: KindFile, Name: "go_only", When: WhenMap{"language": "go"}},
}

applicable := ApplicableGraders(graders, map[string]string{"language": "python"})
if len(applicable) != 2 {
t.Fatalf("expected 2 applicable graders, got %d", len(applicable))
}
if applicable[0].Name != "py_only" || applicable[1].Name != "all" {
t.Errorf("unexpected graders: %s, %s", applicable[0].Name, applicable[1].Name)
}
}

func TestEffectiveWeight(t *testing.T) {
g := GraderConfig{Weight: 0}
if g.EffectiveWeight() != 1.0 {
t.Errorf("expected default weight 1.0, got %f", g.EffectiveWeight())
}
g.Weight = 0.5
if g.EffectiveWeight() != 0.5 {
t.Errorf("expected weight 0.5, got %f", g.EffectiveWeight())
}
}
