import { useState } from "react";
import { Book, Terminal, Settings, Database, ChevronRight } from "lucide-react";

const sections = [
  {
    id: "getting-started",
    icon: Terminal,
    title: "Getting Started",
    content: `## Installation

\`\`\`bash
go install github.com/azure/hyoka@latest
\`\`\`

## Quick Start

1. Clone the repository and install dependencies
2. Configure your AI model credentials in \`config.yaml\`
3. Define your evaluation prompts in the \`prompts/\` directory
4. Run your first evaluation:

\`\`\`bash
hyoka run --config default --prompts prompts/storage-basic.yaml
\`\`\`

## Project Structure

\`\`\`
hyoka/
├── config.yaml          # Model & reviewer configuration
├── prompts/             # Evaluation prompt definitions
│   ├── storage-basic.yaml
│   ├── keyvault-intermediate.yaml
│   └── cosmosdb-advanced.yaml
├── results/             # Evaluation output (auto-generated)
└── analysis/            # AI-generated reports
\`\`\``,
  },
  {
    id: "configuration",
    icon: Settings,
    title: "Configuration",
    content: `## Config File

The \`config.yaml\` file defines model configurations, reviewer settings, and evaluation parameters.

\`\`\`yaml
models:
  - name: gpt-4o
    provider: openai
    temperature: 0.2
    max_tokens: 4096
  - name: claude-3.5-sonnet
    provider: anthropic
    temperature: 0.1
    max_tokens: 4096

reviewers:
  count: 3
  consolidation: weighted_average

languages:
  - python
  - go
  - dotnet
  - java
  - javascript

build_verification:
  enabled: true
  timeout: 60s
\`\`\`

## Environment Variables

| Variable | Description |
|----------|-------------|
| \`OPENAI_API_KEY\` | OpenAI API key for GPT models |
| \`ANTHROPIC_API_KEY\` | Anthropic API key for Claude models |
| \`GITHUB_TOKEN\` | Token for Copilot integration |
| \`AZURE_SUBSCRIPTION_ID\` | Azure subscription for live validation |`,
  },
  {
    id: "prompts",
    icon: Book,
    title: "Writing Prompts",
    content: `## Prompt Format

Prompts are YAML files describing the code generation task:

\`\`\`yaml
name: blob-storage-upload
service: storage
difficulty: basic
category: crud
language: python
description: >
  Create a Python function that uploads a file to Azure Blob Storage
  using the azure-storage-blob SDK. Include proper error handling
  and connection string configuration.

expected_imports:
  - azure.storage.blob
  - azure.core.exceptions

evaluation_criteria:
  - uses_defaultazurecredential: false
  - handles_resource_not_found: true
  - includes_retry_logic: true
\`\`\`

## Difficulty Levels

- **Basic**: Single-operation tasks with straightforward API usage
- **Intermediate**: Multi-step workflows, error handling, pagination
- **Advanced**: Complex scenarios like streaming, transactions, cross-service orchestration

## Categories

\`authentication\` · \`crud\` · \`pagination\` · \`streaming\` · \`error-handling\` · \`configuration\` · \`monitoring\``,
  },
  {
    id: "results",
    icon: Database,
    title: "Understanding Results",
    content: `## Evaluation Output

Each evaluation produces a JSON result file containing:

\`\`\`json
{
  "id": "EVL-0042",
  "prompt": "blob-storage-upload",
  "config": "gpt-4o-default",
  "language": "python",
  "status": "pass",
  "score": 92,
  "generation": {
    "duration_ms": 8300,
    "files_generated": 3,
    "tool_calls": 7,
    "token_usage": { "input": 1240, "output": 3850 }
  },
  "build": {
    "success": true,
    "duration_ms": 4100
  },
  "reviews": [
    {
      "reviewer": "reviewer-1",
      "score": 90,
      "strengths": ["Correct SDK usage", "Good error handling"],
      "issues": ["Missing retry configuration"]
    }
  ],
  "consolidated_score": 92,
  "reasoning_steps": [...]
}
\`\`\`

## Metrics Tracked

- **Pass/Fail**: Binary outcome based on consolidated score threshold
- **Score**: Numeric 0-100 from consolidated reviewer assessments
- **Durations**: Generation, build, and review times
- **Token Usage**: Input/output tokens consumed per generation
- **Tool Calls**: Number and type of tools the AI agent invoked`,
  },
];

export function DocsPage() {
  const [active, setActive] = useState("getting-started");
  const activeSection = sections.find((s) => s.id === active)!;

  return (
    <div className="min-h-screen bg-[#0a0a0f]" style={{ fontFamily: "'Inter', sans-serif" }}>
      <div className="mx-auto flex max-w-7xl flex-col md:flex-row">
        {/* Sidebar */}
        <aside className="border-b border-white/10 p-6 md:w-64 md:border-b-0 md:border-r md:py-10">
          <h2 className="mb-4 text-white/30" style={{ fontSize: 11, letterSpacing: "0.1em", textTransform: "uppercase" }}>
            Documentation
          </h2>
          <nav className="flex gap-1 overflow-x-auto md:flex-col">
            {sections.map((s) => (
              <button
                key={s.id}
                onClick={() => setActive(s.id)}
                className={`flex items-center gap-2.5 whitespace-nowrap rounded-lg px-3 py-2.5 text-left transition ${
                  active === s.id ? "bg-emerald-500/10 text-emerald-400" : "text-white/50 hover:bg-white/5 hover:text-white/70"
                }`}
                style={{ fontSize: 14 }}
              >
                <s.icon className="h-4 w-4 flex-shrink-0" />
                {s.title}
                {active === s.id && <ChevronRight className="ml-auto hidden h-3 w-3 md:block" />}
              </button>
            ))}
          </nav>
        </aside>

        {/* Content */}
        <main className="flex-1 p-6 md:p-10">
          <div className="prose-invert max-w-3xl">
            <div className="mb-6 flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-500/10">
                <activeSection.icon className="h-5 w-5 text-emerald-400" />
              </div>
              <h1 className="text-white" style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 24 }}>
                {activeSection.title}
              </h1>
            </div>
            <div className="text-white/60" style={{ fontSize: 14, lineHeight: 1.8 }}>
              {activeSection.content.split("\n").map((line, i) => {
                if (line.startsWith("## ")) {
                  return <h2 key={i} className="mb-3 mt-8 text-white" style={{ fontSize: 18 }}>{line.replace("## ", "")}</h2>;
                }
                if (line.startsWith("- **")) {
                  const match = line.match(/- \*\*(.+?)\*\*: (.+)/);
                  if (match) {
                    return (
                      <p key={i} className="ml-4 mb-1">
                        <strong className="text-white/80">{match[1]}</strong>: {match[2]}
                      </p>
                    );
                  }
                }
                if (line.startsWith("```")) {
                  return null; // handled below
                }
                if (line.startsWith("| ")) {
                  return null; // simplified
                }
                if (line.trim() === "") return <div key={i} className="h-3" />;
                return <p key={i} className="mb-1">{line}</p>;
              })}

              {/* Code blocks */}
              {activeSection.content.split("```").filter((_, i) => i % 2 === 1).map((block, i) => {
                const lines = block.split("\n");
                const lang = lines[0];
                const code = lines.slice(1).join("\n");
                return (
                  <pre key={i} className="my-4 overflow-x-auto rounded-xl border border-white/8 bg-white/[0.04] p-4">
                    <div className="mb-2 text-emerald-400/50" style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 11 }}>{lang}</div>
                    <code className="text-emerald-300/70" style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 13 }}>
                      {code}
                    </code>
                  </pre>
                );
              })}
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}
