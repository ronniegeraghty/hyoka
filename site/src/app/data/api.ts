import type { RunSummary, EvalReport, DocEntry, DocDetail } from "./types";

const API_BASE = "";

async function fetchJSON<T>(url: string): Promise<T> {
  const res = await fetch(`${API_BASE}${url}`);
  if (!res.ok) {
    throw new Error(`API error ${res.status}: ${res.statusText}`);
  }
  return res.json();
}

export interface PromptInfo {
  id: string;
  service: string;
  plane: string;
  language: string;
  category: string;
  difficulty: string;
  description: string;
  sdk_package: string;
  doc_url: string;
  tags: string[];
  created: string;
  author: string;
  prompt_text: string;
  evaluation_criteria: string;
  file_path: string;
}

export async function fetchRuns(): Promise<RunSummary[]> {
  return fetchJSON<RunSummary[]>("/api/runs");
}

export async function fetchRun(runId: string): Promise<RunSummary> {
  return fetchJSON<RunSummary>(`/api/runs/${encodeURIComponent(runId)}`);
}

export async function fetchEval(runId: string, evalPath: string): Promise<EvalReport> {
  return fetchJSON<EvalReport>(
    `/api/runs/${encodeURIComponent(runId)}/eval?path=${encodeURIComponent(evalPath)}`
  );
}

export async function fetchDocs(): Promise<DocEntry[]> {
  return fetchJSON<DocEntry[]>("/api/docs");
}

export async function fetchDoc(slug: string): Promise<DocDetail> {
  return fetchJSON<DocDetail>(`/api/docs/${encodeURIComponent(slug)}`);
}

export async function fetchPrompts(): Promise<PromptInfo[]> {
  return fetchJSON<PromptInfo[]>("/api/prompts");
}

export async function fetchPrompt(promptId: string): Promise<PromptInfo> {
  return fetchJSON<PromptInfo>(`/api/prompts/${encodeURIComponent(promptId)}`);
}
