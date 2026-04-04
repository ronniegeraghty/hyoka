// ── Types matching the API data shapes ──────────────────────────────

export interface ReviewCriteria {
  name: string;
  passed: boolean;
  reason: string;
}

export interface ReviewScores {
  criteria: ReviewCriteria[];
}

export interface Review {
  overall_score: number;
  max_score: number;
  summary: string;
  strengths?: string[];
  issues?: string[];
  scores?: ReviewScores;
}

export interface ReviewPanelEntry {
  model: string;
  overall_score: number;
  max_score: number;
  summary: string;
  scores?: ReviewScores;
  events?: unknown[];
}

export interface SessionEvent {
  type: string;
  tool_name?: string;
  tool_args?: string;
  tool_result?: string;
  tool_success?: boolean;
  content?: string;
  duration_ms?: number;
  turnNumber?: number;
  input_tokens?: number;
  output_tokens?: number;
  inputTokens?: number;
  outputTokens?: number;
  file_path?: string;
  file_operation?: string;
  file_size?: number;
  mcp_server_name?: string;
  mcp_tool_name?: string;
  error?: string;
}

export interface PromptMetadata {
  service: string;
  plane: string;
  language: string;
  category: string;
  difficulty: string;
  tags?: string[];
  sdk_package?: string;
}

export interface Environment {
  model: string;
  skills_loaded?: string[];
  skills_invoked?: string[];
  available_tools?: string[];
  mcp_servers?: string[];
  totalInputTokens?: number;
  totalOutputTokens?: number;
  total_input_tokens?: number;
  total_output_tokens?: number;
  turnCount?: number;
  turn_count?: number;
}

export interface EvalResult {
  prompt_id: string;
  config_name: string;
  success: boolean;
  error?: string;
  review: Review;
  duration_seconds: number;
  generated_files?: string[];
  prompt_metadata: PromptMetadata;
}

export interface RunSummary {
  run_id: string;
  timestamp: string;
  total_prompts?: number;
  total_configs?: number;
  total_evaluations: number;
  passed: number;
  failed: number;
  errors: number;
  duration_seconds: number;
  avg_generation_duration_seconds?: number;
  avg_review_duration_seconds?: number;
  avg_build_duration_seconds?: number;
  analysis?: string;
  results: EvalResult[];
}

export interface EvalReport {
  prompt_id: string;
  config_name: string;
  timestamp: string;
  success: boolean;
  error?: string;
  duration_seconds: number;
  generation_duration_seconds?: number;
  review_duration_seconds?: number;
  generated_files?: string[];
  session_events?: SessionEvent[];
  event_count?: number;
  tool_calls?: string[];
  review: Review;
  review_panel?: ReviewPanelEntry[];
  prompt_metadata: PromptMetadata;
  environment?: Environment;
  config_used?: { model: string; name: string };
  rerunCommand?: string;
  guardrail_abort_reason?: string;
}

export interface DocEntry {
  slug: string;
  title: string;
}

export interface DocDetail {
  slug: string;
  title: string;
  content: string;
}
