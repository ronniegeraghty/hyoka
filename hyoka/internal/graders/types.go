// Package graders defines the configuration schema for the pluggable grader
// system that replaces the three-tier criteria approach.
//
// Each grader is a single-concern evaluator (file check, build verification,
// LLM review, etc.) defined in YAML config and executed independently.
// Results are aggregated by weighted scoring, with gate graders acting as
// hard pass/fail constraints (DM3).
//
// See docs/grader-config-schema.md for the full schema specification.
package graders

import (
"bytes"
"fmt"
"os"
"path/filepath"
"strings"

"gopkg.in/yaml.v3"
)

// Supported grader kinds.
const (
KindFile           = "file"
KindProgram        = "program"
KindPrompt         = "prompt"
KindBehavior       = "behavior"
KindActionSequence = "action_sequence"
KindToolConstraint = "tool_constraint"
)

// validKinds is the set of recognized grader kind values.
var validKinds = map[string]bool{
KindFile:           true,
KindProgram:        true,
KindPrompt:         true,
KindBehavior:       true,
KindActionSequence: true,
KindToolConstraint: true,
}

// GraderConfigFile is the top-level YAML structure containing a list of graders.
type GraderConfigFile struct {
Graders []GraderConfig `yaml:"graders" json:"graders"`
}

// GraderConfig defines a single grader instance in the evaluation pipeline.
type GraderConfig struct {
Kind   string    `yaml:"kind" json:"kind"`
Name   string    `yaml:"name" json:"name"`
Config yaml.Node `yaml:"config" json:"config"`
Weight float64   `yaml:"weight,omitempty" json:"weight,omitempty"`
Gate   bool      `yaml:"gate,omitempty" json:"gate,omitempty"`
When   WhenMap   `yaml:"when,omitempty" json:"when,omitempty"`
}

// WhenMap holds property-based applicability conditions.
// All entries must match for the grader to apply (AND logic).
// Matching is case-insensitive.
type WhenMap map[string]string

// Matches returns true if all conditions in the map match the given properties.
// An empty WhenMap matches everything.
func (w WhenMap) Matches(props map[string]string) bool {
for k, v := range w {
pv, ok := props[k]
if !ok || !strings.EqualFold(v, pv) {
return false
}
}
return true
}

// FileConfig holds configuration for the "file" grader kind.
type FileConfig struct {
Path      string `yaml:"path" json:"path"`
Pattern   string `yaml:"pattern,omitempty" json:"pattern,omitempty"`
MustExist *bool  `yaml:"must_exist,omitempty" json:"must_exist,omitempty"`
}

// ProgramConfig holds configuration for the "program" grader kind.
type ProgramConfig struct {
Command string   `yaml:"command" json:"command"`
Args    []string `yaml:"args,omitempty" json:"args,omitempty"`
Timeout int      `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// PromptConfig holds configuration for the "prompt" grader kind.
// Each prompt grader runs exactly one model (DM19).
type PromptConfig struct {
Model  string `yaml:"model" json:"model"`
Rubric string `yaml:"rubric" json:"rubric"`
}

// BehaviorConfig holds configuration for the "behavior" grader kind.
type BehaviorConfig struct {
RequiredTools  []string `yaml:"required_tools,omitempty" json:"required_tools,omitempty"`
ForbiddenTools []string `yaml:"forbidden_tools,omitempty" json:"forbidden_tools,omitempty"`
MaxTurns       int      `yaml:"max_turns,omitempty" json:"max_turns,omitempty"`
}

// ActionSequenceConfig holds configuration for the "action_sequence" grader kind.
type ActionSequenceConfig struct {
ExpectedActions []string `yaml:"expected_actions" json:"expected_actions"`
}

// ToolConstraintConfig holds configuration for the "tool_constraint" grader kind.
type ToolConstraintConfig struct {
Required  []string `yaml:"required,omitempty" json:"required,omitempty"`
Forbidden []string `yaml:"forbidden,omitempty" json:"forbidden,omitempty"`
MinCalls  int      `yaml:"min_calls,omitempty" json:"min_calls,omitempty"`
MaxCalls  int      `yaml:"max_calls,omitempty" json:"max_calls,omitempty"`
}

// EffectiveWeight returns the grader's weight, defaulting to 1.0 if unset.
func (g *GraderConfig) EffectiveWeight() float64 {
if g.Weight == 0 {
return 1.0
}
return g.Weight
}

// DecodeConfig decodes the raw YAML config node into the appropriate
// kind-specific config struct. Returns an error if the kind is unknown
// or the config doesn't match the expected schema.
func (g *GraderConfig) DecodeConfig() (any, error) {
switch g.Kind {
case KindFile:
var c FileConfig
if err := g.Config.Decode(&c); err != nil {
return nil, fmt.Errorf("decoding file config for %q: %w", g.Name, err)
}
return &c, nil
case KindProgram:
var c ProgramConfig
if err := g.Config.Decode(&c); err != nil {
return nil, fmt.Errorf("decoding program config for %q: %w", g.Name, err)
}
return &c, nil
case KindPrompt:
var c PromptConfig
if err := g.Config.Decode(&c); err != nil {
return nil, fmt.Errorf("decoding prompt config for %q: %w", g.Name, err)
}
return &c, nil
case KindBehavior:
var c BehaviorConfig
if err := g.Config.Decode(&c); err != nil {
return nil, fmt.Errorf("decoding behavior config for %q: %w", g.Name, err)
}
return &c, nil
case KindActionSequence:
var c ActionSequenceConfig
if err := g.Config.Decode(&c); err != nil {
return nil, fmt.Errorf("decoding action_sequence config for %q: %w", g.Name, err)
}
return &c, nil
case KindToolConstraint:
var c ToolConstraintConfig
if err := g.Config.Decode(&c); err != nil {
return nil, fmt.Errorf("decoding tool_constraint config for %q: %w", g.Name, err)
}
return &c, nil
default:
return nil, fmt.Errorf("unknown grader kind %q for %q", g.Kind, g.Name)
}
}

// Parse decodes YAML bytes into a GraderConfigFile and validates it.
func Parse(data []byte) (*GraderConfigFile, error) {
var gcf GraderConfigFile
dec := yaml.NewDecoder(bytes.NewReader(data))
dec.KnownFields(true)
if err := dec.Decode(&gcf); err != nil {
return nil, fmt.Errorf("parsing grader config: %w", err)
}
if err := Validate(&gcf); err != nil {
return nil, err
}
return &gcf, nil
}

// LoadFile loads and parses a single grader config YAML file.
func LoadFile(path string) (*GraderConfigFile, error) {
data, err := os.ReadFile(path)
if err != nil {
return nil, fmt.Errorf("reading grader config %s: %w", path, err)
}
gcf, err := Parse(data)
if err != nil {
return nil, fmt.Errorf("in %s: %w", path, err)
}
return gcf, nil
}

// LoadDir loads all grader config YAML files from a directory tree.
func LoadDir(dir string) (*GraderConfigFile, error) {
merged := &GraderConfigFile{}

err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
if err != nil {
return err
}
if info.IsDir() {
return nil
}
ext := filepath.Ext(path)
if ext != ".yaml" && ext != ".yml" {
return nil
}
gcf, err := LoadFile(path)
if err != nil {
return err
}
merged.Graders = append(merged.Graders, gcf.Graders...)
return nil
})
if err != nil {
return nil, fmt.Errorf("walking grader config directory %s: %w", dir, err)
}
return merged, nil
}

// Validate checks a GraderConfigFile for structural correctness.
func Validate(gcf *GraderConfigFile) error {
if len(gcf.Graders) == 0 {
return fmt.Errorf("no graders defined")
}

names := make(map[string]bool, len(gcf.Graders))
for i, g := range gcf.Graders {
if g.Name == "" {
return fmt.Errorf("grader at index %d: name is required", i)
}
if g.Kind == "" {
return fmt.Errorf("grader %q: kind is required", g.Name)
}
if !validKinds[g.Kind] {
return fmt.Errorf("grader %q: unknown kind %q", g.Name, g.Kind)
}
if names[g.Name] {
return fmt.Errorf("grader %q: duplicate name", g.Name)
}
names[g.Name] = true
if g.Weight < 0 || g.Weight > 1 {
return fmt.Errorf("grader %q: weight must be between 0.0 and 1.0, got %f", g.Name, g.Weight)
}
if _, err := g.DecodeConfig(); err != nil {
return err
}
}
return nil
}

// ApplicableGraders filters graders by the given prompt properties,
// returning only those whose When conditions match.
func ApplicableGraders(graders []GraderConfig, props map[string]string) []GraderConfig {
var applicable []GraderConfig
for _, g := range graders {
if g.When.Matches(props) {
applicable = append(applicable, g)
}
}
return applicable
}
