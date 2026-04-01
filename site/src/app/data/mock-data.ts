// ── Types matching the real data shape ──────────────────────────────

export interface ReviewCriteria {
  name: string;
  passed: boolean;
  reason: string;
}

export interface Review {
  model: string;
  overall_score: number;
  max_score: number;
  summary: string;
  strengths: string[];
  issues: string[];
  criteria: ReviewCriteria[];
}

export interface SessionEvent {
  type: string;
  tool_name?: string;
  tool_args?: string;
  tool_result?: string;
  tool_success?: boolean;
  content?: string;
  duration_ms: number;
  turn_number: number;
  input_tokens?: number;
  output_tokens?: number;
  file_path?: string;
  file_operation?: string;
  file_size?: number;
}

export interface ToolUsage {
  expected_tools: string[];
  actual_tools: string[];
  matched: string[];
  missed: string[];
  extra: string[];
}

export interface Build {
  language: string;
  command: string;
  exit_code: number;
  stdout: string;
  stderr: string;
  success: boolean;
}

export interface PromptMetadata {
  service: string;
  plane: string;
  language: string;
  category: string;
  difficulty: string;
  tags: string[];
  sdk_package: string;
}

export interface Environment {
  model: string;
  skills_loaded: string[];
  skills_invoked: string[];
  available_tools: string[];
  mcp_servers: string[];
  total_input_tokens: number;
  total_output_tokens: number;
  turn_count: number;
}

export interface EvalReport {
  prompt_id: string;
  config_name: string;
  timestamp: string;
  success: boolean;
  duration_seconds: number;
  generation_duration_seconds: number;
  review_duration_seconds: number;
  build_duration_seconds: number;
  generated_files: string[];
  tool_calls: string[];
  event_count: number;
  error: string;
  error_category: string;
  failure_reason: string;
  rerun_command: string;
  prompt_metadata: PromptMetadata;
  environment: Environment;
  review: Review;
  review_panel: Review[];
  session_events: SessionEvent[];
  tool_usage: ToolUsage;
  build: Build;
  reviewed_files: { path: string; content: string }[];
}

export interface ConfigPassRate {
  config: string;
  total: number;
  passed: number;
  failed: number;
  rate: number;
}

export interface PromptPassRate {
  prompt: string;
  total: number;
  passed: number;
  failed: number;
  rate: number;
}

export interface ToolStat {
  name: string;
  count: number;
  successes: number;
  failures: number;
  success_rate: number;
}

export interface SummaryStats {
  duration_by_config: Record<string, { min: number; avg: number; max: number }>;
  duration_by_prompt: Record<string, { min: number; avg: number; max: number }>;
  slowest_eval: string;
  fastest_eval: string;
  config_pass_rates: ConfigPassRate[];
  prompt_pass_rates: PromptPassRate[];
  prompt_deltas: { prompt_id: string; pass_config: string; fail_config: string }[];
  tool_stats: ToolStat[];
}

export interface RunSummary {
  run_id: string;
  timestamp: string;
  total_prompts: number;
  total_configs: number;
  total_evaluations: number;
  passed: number;
  failed: number;
  errors: number;
  duration_seconds: number;
  avg_generation_duration_seconds: number;
  avg_review_duration_seconds: number;
  avg_build_duration_seconds: number;
  analysis: string;
  results: EvalReport[];
}

export interface HistoryReport {
  prompt_id: string;
  total_runs: number;
  passed: number;
  pass_rate: number;
  avg_duration_seconds: number;
  entries: { run_id: string; config_name: string; success: boolean; duration: number; file_count: number; score: number }[];
  configs: { config: string; runs: number; passed: number; pass_rate: number; avg_duration: number }[];
}

// ── Mock Data ───────────────────────────────────────────────────────

const services = ["storage", "key-vault", "cosmos-db", "event-hubs", "service-bus", "identity"];
const planes = ["data-plane", "management-plane"];
const languages = ["dotnet", "python", "java", "js-ts", "go"];
const categories = ["authentication", "crud", "pagination", "streaming", "error-handling", "configuration", "monitoring"];
const difficulties: ("basic" | "intermediate" | "advanced")[] = ["basic", "intermediate", "advanced"];
const configs = ["baseline-opus", "baseline-sonnet", "copilot-gpt4o", "copilot-gpt4o-mini", "claude-3.5-haiku"];
const models = ["claude-3-opus", "claude-3.5-sonnet", "gpt-4o", "gpt-4o-mini", "claude-3.5-haiku"];
const reviewModels = ["gpt-4o-review", "claude-3.5-review", "gemini-1.5-review"];

const toolNames = [
  "read_file", "write_file", "list_directory", "search_code", "run_terminal",
  "find_references", "get_documentation", "create_file", "edit_file", "delete_file"
];

const sdkPackages: Record<string, Record<string, string>> = {
  storage: { dotnet: "Azure.Storage.Blobs", python: "azure-storage-blob", java: "azure-storage-blob", "js-ts": "@azure/storage-blob", go: "azblob" },
  "key-vault": { dotnet: "Azure.Security.KeyVault.Secrets", python: "azure-keyvault-secrets", java: "azure-security-keyvault-secrets", "js-ts": "@azure/keyvault-secrets", go: "azsecrets" },
  "cosmos-db": { dotnet: "Microsoft.Azure.Cosmos", python: "azure-cosmos", java: "azure-cosmos", "js-ts": "@azure/cosmos", go: "azcosmos" },
  "event-hubs": { dotnet: "Azure.Messaging.EventHubs", python: "azure-eventhub", java: "azure-messaging-eventhubs", "js-ts": "@azure/event-hubs", go: "azeventhubs" },
  "service-bus": { dotnet: "Azure.Messaging.ServiceBus", python: "azure-servicebus", java: "azure-messaging-servicebus", "js-ts": "@azure/service-bus", go: "azservicebus" },
  identity: { dotnet: "Azure.Identity", python: "azure-identity", java: "azure-identity", "js-ts": "@azure/identity", go: "azidentity" },
};

function pick<T>(arr: T[]): T {
  return arr[Math.floor(Math.random() * arr.length)];
}

function pickN<T>(arr: T[], n: number): T[] {
  const shuffled = [...arr].sort(() => Math.random() - 0.5);
  return shuffled.slice(0, n);
}

function randBetween(min: number, max: number, decimals = 1): number {
  return parseFloat((min + Math.random() * (max - min)).toFixed(decimals));
}

let evalCounter = 0;

function generatePromptId(service: string, plane: string, lang: string, cat: string): string {
  const p = plane === "data-plane" ? "dp" : "mp";
  return `${service}-${p}-${lang}-${cat}`;
}

function generateSessionEvents(success: boolean): SessionEvent[] {
  const events: SessionEvent[] = [];
  let turn = 1;

  events.push({
    type: "prompt",
    content: "Generate Azure SDK code for the given task...",
    duration_ms: 0,
    turn_number: turn,
    input_tokens: randBetween(200, 800, 0),
    output_tokens: 0,
  });

  events.push({
    type: "reasoning",
    content: "I need to analyze the requirements and determine which Azure SDK packages to use...",
    duration_ms: randBetween(500, 2000),
    turn_number: turn,
    input_tokens: 0,
    output_tokens: randBetween(100, 400, 0),
  });

  const numToolCalls = randBetween(3, 8, 0);
  for (let i = 0; i < numToolCalls; i++) {
    const tool = pick(toolNames);
    const toolSuccess = Math.random() > 0.1;
    turn++;
    events.push({
      type: "tool_call",
      tool_name: tool,
      tool_args: JSON.stringify({ path: `/src/example_${i}.py` }),
      duration_ms: randBetween(100, 1500),
      turn_number: turn,
      input_tokens: randBetween(50, 200, 0),
      output_tokens: randBetween(100, 500, 0),
    });
    events.push({
      type: "tool_result",
      tool_name: tool,
      tool_success: toolSuccess,
      tool_result: toolSuccess ? "Operation completed successfully" : "Error: file not found",
      duration_ms: randBetween(50, 500),
      turn_number: turn,
    });
    if (tool === "write_file" || tool === "create_file") {
      events.push({
        type: "file_write",
        file_path: `/src/generated_${i}.py`,
        file_operation: tool === "create_file" ? "create" : "write",
        file_size: randBetween(500, 5000, 0),
        duration_ms: randBetween(10, 50),
        turn_number: turn,
      });
    }
  }

  if (!success && Math.random() > 0.5) {
    events.push({
      type: "warning",
      content: "Potential issue detected: missing error handling for ResourceNotFoundError",
      duration_ms: 0,
      turn_number: turn,
    });
  }

  events.push({
    type: "reply",
    content: "I've generated the requested Azure SDK code. Here's a summary of what was created...",
    duration_ms: randBetween(200, 1000),
    turn_number: turn + 1,
    input_tokens: randBetween(100, 300, 0),
    output_tokens: randBetween(200, 800, 0),
  });

  return events;
}

function generateReview(success: boolean): Review {
  const score = success ? randBetween(70, 100, 0) : randBetween(30, 65, 0);
  const allStrengths = [
    "Correct SDK package imports",
    "Proper credential management using DefaultAzureCredential",
    "Good error handling with specific exception types",
    "Clean code structure and separation of concerns",
    "Appropriate use of async/await patterns",
    "Comprehensive logging implementation",
    "Follows Azure SDK best practices for retry policies",
    "Proper resource cleanup with context managers",
  ];
  const allIssues = [
    "Missing retry policy configuration",
    "No error handling for ResourceNotFoundError",
    "Hardcoded connection string instead of using environment variables",
    "Missing proper pagination implementation",
    "No timeout configuration for long-running operations",
    "Incomplete error handling for transient failures",
    "Missing input validation for user-provided parameters",
    "No logging or telemetry instrumentation",
  ];

  const criteriaNames = [
    "SDK Package Usage", "Authentication", "Error Handling", "Resource Cleanup",
    "Code Structure", "Best Practices", "Documentation", "Security",
  ];

  return {
    model: pick(reviewModels),
    overall_score: score,
    max_score: 100,
    summary: success
      ? "The generated code demonstrates solid understanding of the Azure SDK patterns with minor areas for improvement."
      : "The code has significant gaps in error handling and does not follow recommended Azure SDK patterns.",
    strengths: pickN(allStrengths, randBetween(2, 4, 0)),
    issues: pickN(allIssues, success ? randBetween(1, 2, 0) : randBetween(3, 5, 0)),
    criteria: criteriaNames.map((name) => ({
      name,
      passed: success ? Math.random() > 0.2 : Math.random() > 0.5,
      reason: `${name} evaluation: ${success ? "meets" : "partially meets"} expectations`,
    })),
  };
}

function generateEvalReport(promptId?: string, configName?: string): EvalReport {
  evalCounter++;
  const service = promptId ? promptId.split("-")[0] + (promptId.includes("key-vault") ? "-vault" : promptId.includes("cosmos-db") ? "-db" : promptId.includes("event-hubs") ? "-hubs" : promptId.includes("service-bus") ? "-bus" : "") : pick(services);
  const realService = services.find(s => service.startsWith(s.split("-")[0])) || pick(services);
  const lang = promptId ? (promptId.split("-").find(p => languages.includes(p)) || pick(languages)) : pick(languages);
  const cat = pick(categories);
  const difficulty = pick(difficulties);
  const plane = pick(planes);
  const config = configName || pick(configs);
  const modelIdx = configs.indexOf(config);
  const model = modelIdx >= 0 ? models[modelIdx] : pick(models);
  const success = Math.random() > 0.28;
  const pid = promptId || generatePromptId(realService, plane, lang, cat);

  const genDuration = randBetween(3, 25);
  const buildDuration = randBetween(1, 8);
  const reviewDuration = randBetween(4, 15);
  const totalDuration = genDuration + buildDuration + reviewDuration + randBetween(0.5, 3);

  const usedTools = pickN(toolNames, randBetween(3, 7, 0));
  const expectedTools = [...usedTools.slice(0, 3), ...pickN(toolNames, 2)];
  const matched = usedTools.filter(t => expectedTools.includes(t));
  const missed = expectedTools.filter(t => !usedTools.includes(t));
  const extra = usedTools.filter(t => !expectedTools.includes(t));

  const numFiles = randBetween(1, 6, 0);
  const langExt: Record<string, string> = { dotnet: ".cs", python: ".py", java: ".java", "js-ts": ".ts", go: ".go" };
  const ext = langExt[lang] || ".py";

  const review = generateReview(success);
  const reviewPanel = [review, generateReview(success), generateReview(success)];

  return {
    prompt_id: pid,
    config_name: config,
    timestamp: `2026-03-${randBetween(20, 29, 0).toString().padStart(2, "0")}T${randBetween(8, 22, 0).toString().padStart(2, "0")}:${randBetween(0, 59, 0).toString().padStart(2, "0")}:00Z`,
    success,
    duration_seconds: totalDuration,
    generation_duration_seconds: genDuration,
    review_duration_seconds: reviewDuration,
    build_duration_seconds: buildDuration,
    generated_files: Array.from({ length: numFiles }, (_, i) => `/src/generated_${i}${ext}`),
    tool_calls: usedTools,
    event_count: randBetween(8, 25, 0),
    error: success ? "" : pick(["Build failed", "Review score below threshold", "Generation timeout"]),
    error_category: success ? "" : pick(["build_failure", "low_score", "timeout"]),
    failure_reason: success ? "" : "Generated code did not meet quality threshold",
    rerun_command: `hyoka run --prompt ${pid} --config ${config}`,
    prompt_metadata: {
      service: realService,
      plane,
      language: lang,
      category: cat,
      difficulty,
      tags: [realService, lang, cat, difficulty, plane],
      sdk_package: sdkPackages[realService]?.[lang] || "azure-sdk",
    },
    environment: {
      model,
      skills_loaded: ["azure-sdk-knowledge", "code-generation", "code-review"],
      skills_invoked: pickN(["azure-sdk-knowledge", "code-generation"], randBetween(1, 2, 0)),
      available_tools: toolNames,
      mcp_servers: ["filesystem", "terminal"],
      total_input_tokens: randBetween(2000, 12000, 0),
      total_output_tokens: randBetween(3000, 15000, 0),
      turn_count: randBetween(4, 15, 0),
    },
    review,
    review_panel: reviewPanel,
    session_events: generateSessionEvents(success),
    tool_usage: { expected_tools: expectedTools, actual_tools: usedTools, matched, missed, extra },
    build: {
      language: lang,
      command: lang === "dotnet" ? "dotnet build" : lang === "python" ? "python -m py_compile" : lang === "go" ? "go build ./..." : lang === "java" ? "mvn compile" : "npx tsc --noEmit",
      exit_code: success ? 0 : 1,
      stdout: success ? "Build succeeded." : "",
      stderr: success ? "" : "error: type 'BlobClient' has no member 'upload_blob_async'",
      success: success || Math.random() > 0.3,
    },
    reviewed_files: Array.from({ length: numFiles }, (_, i) => ({
      path: `/src/generated_${i}${ext}`,
      content: `// Generated code for ${pid}\n// ... implementation ...`,
    })),
  };
}

// ── Generate prompt IDs deterministically ──────────────────────────

const promptIds: string[] = [];
for (const service of services) {
  for (const lang of languages) {
    for (const cat of categories.slice(0, 3)) {
      promptIds.push(generatePromptId(service, pick(planes), lang, cat));
    }
  }
}

// ── Runs ────────────────────────────────────────────────────────────

function generateRun(runId: string, timestamp: string): RunSummary {
  const numEvals = randBetween(15, 35, 0);
  const results: EvalReport[] = [];
  for (let i = 0; i < numEvals; i++) {
    results.push(generateEvalReport(pick(promptIds), pick(configs)));
  }
  const passed = results.filter(r => r.success).length;
  const failed = results.filter(r => !r.success).length;

  return {
    run_id: runId,
    timestamp,
    total_prompts: new Set(results.map(r => r.prompt_id)).size,
    total_configs: new Set(results.map(r => r.config_name)).size,
    total_evaluations: numEvals,
    passed,
    failed,
    errors: randBetween(0, 2, 0),
    duration_seconds: results.reduce((sum, r) => sum + r.duration_seconds, 0),
    avg_generation_duration_seconds: parseFloat((results.reduce((s, r) => s + r.generation_duration_seconds, 0) / numEvals).toFixed(1)),
    avg_review_duration_seconds: parseFloat((results.reduce((s, r) => s + r.review_duration_seconds, 0) / numEvals).toFixed(1)),
    avg_build_duration_seconds: parseFloat((results.reduce((s, r) => s + r.build_duration_seconds, 0) / numEvals).toFixed(1)),
    analysis: generateAnalysis(passed, failed, results),
    results,
  };
}

function generateAnalysis(passed: number, failed: number, results: EvalReport[]): string {
  const rate = ((passed / (passed + failed)) * 100).toFixed(1);
  const topService = [...new Set(results.filter(r => r.success).map(r => r.prompt_metadata.service))]
    .sort((a, b) => results.filter(r => r.success && r.prompt_metadata.service === b).length - results.filter(r => r.success && r.prompt_metadata.service === a).length)[0] || "identity";
  return `This run achieved a ${rate}% pass rate across ${passed + failed} evaluations. The ${topService} service showed the strongest results. Python and TypeScript continue to outperform other languages. Advanced difficulty prompts involving streaming and pagination remain the most challenging, with pass rates 15-20% below basic prompts. Build verification failures account for approximately 30% of all failures, suggesting generated code often has syntactic issues before reaching the review stage.`;
}

export const mockRuns: RunSummary[] = [
  generateRun("20260329-143200", "2026-03-29T14:32:00Z"),
  generateRun("20260328-091500", "2026-03-28T09:15:00Z"),
  generateRun("20260327-213400", "2026-03-27T21:34:00Z"),
  generateRun("20260326-160800", "2026-03-26T16:08:00Z"),
  generateRun("20260325-112200", "2026-03-25T11:22:00Z"),
  generateRun("20260324-083000", "2026-03-24T08:30:00Z"),
];

// ── Summary Stats ──────────────────────────────────────────────────

export function computeSummaryStats(results: EvalReport[]): SummaryStats {
  const byConfig: Record<string, number[]> = {};
  const byPrompt: Record<string, number[]> = {};
  const configCounts: Record<string, { total: number; passed: number; failed: number }> = {};
  const promptCounts: Record<string, { total: number; passed: number; failed: number }> = {};
  const toolCounts: Record<string, { count: number; successes: number; failures: number }> = {};

  for (const r of results) {
    if (!byConfig[r.config_name]) byConfig[r.config_name] = [];
    byConfig[r.config_name].push(r.duration_seconds);

    if (!byPrompt[r.prompt_id]) byPrompt[r.prompt_id] = [];
    byPrompt[r.prompt_id].push(r.duration_seconds);

    if (!configCounts[r.config_name]) configCounts[r.config_name] = { total: 0, passed: 0, failed: 0 };
    configCounts[r.config_name].total++;
    if (r.success) configCounts[r.config_name].passed++;
    else configCounts[r.config_name].failed++;

    if (!promptCounts[r.prompt_id]) promptCounts[r.prompt_id] = { total: 0, passed: 0, failed: 0 };
    promptCounts[r.prompt_id].total++;
    if (r.success) promptCounts[r.prompt_id].passed++;
    else promptCounts[r.prompt_id].failed++;

    for (const event of r.session_events) {
      if (event.type === "tool_call" && event.tool_name) {
        if (!toolCounts[event.tool_name]) toolCounts[event.tool_name] = { count: 0, successes: 0, failures: 0 };
        toolCounts[event.tool_name].count++;
      }
      if (event.type === "tool_result" && event.tool_name) {
        if (toolCounts[event.tool_name]) {
          if (event.tool_success) toolCounts[event.tool_name].successes++;
          else toolCounts[event.tool_name].failures++;
        }
      }
    }
  }

  const durationByConfig: Record<string, { min: number; avg: number; max: number }> = {};
  for (const [k, v] of Object.entries(byConfig)) {
    durationByConfig[k] = { min: Math.min(...v), avg: parseFloat((v.reduce((a, b) => a + b, 0) / v.length).toFixed(1)), max: Math.max(...v) };
  }

  const durationByPrompt: Record<string, { min: number; avg: number; max: number }> = {};
  for (const [k, v] of Object.entries(byPrompt)) {
    durationByPrompt[k] = { min: Math.min(...v), avg: parseFloat((v.reduce((a, b) => a + b, 0) / v.length).toFixed(1)), max: Math.max(...v) };
  }

  const sorted = [...results].sort((a, b) => b.duration_seconds - a.duration_seconds);

  return {
    duration_by_config: durationByConfig,
    duration_by_prompt: durationByPrompt,
    slowest_eval: sorted[0]?.prompt_id || "",
    fastest_eval: sorted[sorted.length - 1]?.prompt_id || "",
    config_pass_rates: Object.entries(configCounts).map(([config, c]) => ({
      config, total: c.total, passed: c.passed, failed: c.failed, rate: parseFloat(((c.passed / c.total) * 100).toFixed(1)),
    })),
    prompt_pass_rates: Object.entries(promptCounts).map(([prompt, c]) => ({
      prompt, total: c.total, passed: c.passed, failed: c.failed, rate: parseFloat(((c.passed / c.total) * 100).toFixed(1)),
    })),
    prompt_deltas: Object.keys(promptCounts).slice(0, 5).map(pid => ({
      prompt_id: pid,
      pass_config: pick(configs),
      fail_config: pick(configs.filter(c => c !== configs[0])),
    })),
    tool_stats: Object.entries(toolCounts).map(([name, c]) => ({
      name, count: c.count, successes: c.successes, failures: c.failures,
      success_rate: c.count > 0 ? parseFloat(((c.successes / c.count) * 100).toFixed(1)) : 0,
    })).sort((a, b) => b.count - a.count),
  };
}

// ── History Reports ────────────────────────────────────────────────

export function generateHistoryReport(promptId: string): HistoryReport {
  const entries: HistoryReport["entries"] = [];
  for (const run of mockRuns) {
    for (const config of pickN(configs, randBetween(2, 4, 0))) {
      const success = Math.random() > 0.3;
      entries.push({
        run_id: run.run_id,
        config_name: config,
        success,
        duration: randBetween(5, 30),
        file_count: randBetween(1, 6, 0),
        score: success ? randBetween(70, 98, 0) : randBetween(25, 60, 0),
      });
    }
  }

  const passed = entries.filter(e => e.success).length;
  const configMap: Record<string, { runs: number; passed: number; totalDuration: number }> = {};
  for (const e of entries) {
    if (!configMap[e.config_name]) configMap[e.config_name] = { runs: 0, passed: 0, totalDuration: 0 };
    configMap[e.config_name].runs++;
    if (e.success) configMap[e.config_name].passed++;
    configMap[e.config_name].totalDuration += e.duration;
  }

  return {
    prompt_id: promptId,
    total_runs: entries.length,
    passed,
    pass_rate: parseFloat(((passed / entries.length) * 100).toFixed(1)),
    avg_duration_seconds: parseFloat((entries.reduce((s, e) => s + e.duration, 0) / entries.length).toFixed(1)),
    entries,
    configs: Object.entries(configMap).map(([config, c]) => ({
      config,
      runs: c.runs,
      passed: c.passed,
      pass_rate: parseFloat(((c.passed / c.runs) * 100).toFixed(1)),
      avg_duration: parseFloat((c.totalDuration / c.runs).toFixed(1)),
    })),
  };
}

// ── All unique prompts from all runs ───────────────────────────────

export function getAllPrompts(): { prompt_id: string; metadata: PromptMetadata; evalCount: number; passRate: number }[] {
  const map: Record<string, { metadata: PromptMetadata; total: number; passed: number }> = {};
  for (const run of mockRuns) {
    for (const r of run.results) {
      if (!map[r.prompt_id]) map[r.prompt_id] = { metadata: r.prompt_metadata, total: 0, passed: 0 };
      map[r.prompt_id].total++;
      if (r.success) map[r.prompt_id].passed++;
    }
  }
  return Object.entries(map).map(([prompt_id, v]) => ({
    prompt_id,
    metadata: v.metadata,
    evalCount: v.total,
    passRate: parseFloat(((v.passed / v.total) * 100).toFixed(1)),
  }));
}

export function getEvalByIds(promptId: string, configName: string): EvalReport | undefined {
  for (const run of mockRuns) {
    const found = run.results.find(r => r.prompt_id === promptId && r.config_name === configName);
    if (found) return found;
  }
  // Generate one if not found
  return generateEvalReport(promptId, configName);
}

// ── Correlation analysis for a specific prompt ─────────────────────

export interface CorrelationStat {
  name: string;
  total: number;
  passed: number;
  failed: number;
  rate: number;
  avgScore: number;
  avgDuration: number;
}

export interface PromptCorrelations {
  byModel: CorrelationStat[];
  byTool: CorrelationStat[];
  bySkill: CorrelationStat[];
  byMcpServer: CorrelationStat[];
  overallRate: number;
  overallAvgScore: number;
}

export function getEvalsForPrompt(promptId: string): EvalReport[] {
  const results: EvalReport[] = [];
  for (const run of mockRuns) {
    for (const r of run.results) {
      if (r.prompt_id === promptId) results.push(r);
    }
  }
  // If we don't have enough data from runs, generate some extra evals
  if (results.length < 8) {
    for (const config of configs) {
      if (!results.find(r => r.config_name === config)) {
        results.push(generateEvalReport(promptId, config));
      }
    }
  }
  return results;
}

export function computePromptCorrelations(promptId: string): PromptCorrelations {
  const evals = getEvalsForPrompt(promptId);

  const overallPassed = evals.filter(e => e.success).length;
  const overallRate = evals.length > 0 ? parseFloat(((overallPassed / evals.length) * 100).toFixed(1)) : 0;
  const overallAvgScore = evals.length > 0 ? parseFloat((evals.reduce((s, e) => s + e.review.overall_score, 0) / evals.length).toFixed(1)) : 0;

  function computeGrouped(keyFn: (e: EvalReport) => string[]): CorrelationStat[] {
    const map: Record<string, { total: number; passed: number; scoreSum: number; durationSum: number }> = {};
    for (const e of evals) {
      for (const key of keyFn(e)) {
        if (!map[key]) map[key] = { total: 0, passed: 0, scoreSum: 0, durationSum: 0 };
        map[key].total++;
        if (e.success) map[key].passed++;
        map[key].scoreSum += e.review.overall_score;
        map[key].durationSum += e.duration_seconds;
      }
    }
    return Object.entries(map)
      .map(([name, v]) => ({
        name,
        total: v.total,
        passed: v.passed,
        failed: v.total - v.passed,
        rate: parseFloat(((v.passed / v.total) * 100).toFixed(1)),
        avgScore: parseFloat((v.scoreSum / v.total).toFixed(1)),
        avgDuration: parseFloat((v.durationSum / v.total).toFixed(1)),
      }))
      .sort((a, b) => b.rate - a.rate);
  }

  const byModel = computeGrouped(e => [e.environment.model]);

  // For tools, consider which tools were actually used in the eval
  const byTool = computeGrouped(e => [...new Set(e.tool_usage.actual_tools)]);

  const bySkill = computeGrouped(e => e.environment.skills_invoked);

  const byMcpServer = computeGrouped(e => e.environment.mcp_servers);

  return { byModel, byTool, bySkill, byMcpServer, overallRate, overallAvgScore };
}