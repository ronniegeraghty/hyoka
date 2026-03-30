import { useParams, Link } from "react-router";
import { getEvalByIds } from "../data/mock-data";
import { useMemo, useState } from "react";
import {
  ArrowLeft, CheckCircle2, XCircle, Clock, FileCode2, Cpu,
  MessageSquare, Wrench, Terminal, ChevronDown, ChevronRight,
  AlertTriangle, Zap, Copy, Check
} from "lucide-react";

const mono = { fontFamily: "'JetBrains Mono', monospace" };

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <button
      onClick={() => { navigator.clipboard.writeText(text); setCopied(true); setTimeout(() => setCopied(false), 2000); }}
      className="text-white/30 transition hover:text-white/60"
    >
      {copied ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
    </button>
  );
}

export function EvalDetailPage() {
  const { promptId, configName } = useParams();
  const evalReport = useMemo(() => getEvalByIds(
    decodeURIComponent(promptId || ""),
    decodeURIComponent(configName || "")
  ), [promptId, configName]);

  const [expandedReviewer, setExpandedReviewer] = useState<number>(0);
  const [showTimeline, setShowTimeline] = useState(true);
  const [showBuild, setShowBuild] = useState(false);
  const [showFiles, setShowFiles] = useState(false);

  if (!evalReport) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[#0a0a0f]">
        <p className="text-white/50">Evaluation not found</p>
      </div>
    );
  }

  const r = evalReport;
  const scoreColor = r.review.overall_score >= 80 ? "text-emerald-400" : r.review.overall_score >= 60 ? "text-amber-400" : "text-red-400";

  return (
    <div className="min-h-screen bg-[#0a0a0f] px-4 py-8 sm:px-6" style={{ fontFamily: "'Inter', sans-serif" }}>
      <div className="mx-auto max-w-6xl">
        <Link to={`/prompts/${encodeURIComponent(r.prompt_id)}`} className="mb-6 inline-flex items-center gap-1.5 text-white/40 no-underline transition hover:text-emerald-400" style={{ fontSize: 13 }}>
          <ArrowLeft className="h-3.5 w-3.5" /> {r.prompt_id}
        </Link>

        {/* Header */}
        <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <div className="mb-2 flex items-center gap-3">
              {r.success ? <CheckCircle2 className="h-5 w-5 text-emerald-400" /> : <XCircle className="h-5 w-5 text-red-400" />}
              <h1 className="text-white" style={{ ...mono, fontSize: "clamp(1.1rem, 2.5vw, 1.5rem)" }}>
                {r.prompt_id}
              </h1>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <span className="rounded-md bg-blue-500/10 px-2.5 py-1 text-blue-400/80" style={{ ...mono, fontSize: 11 }}>{r.config_name}</span>
              <span className="rounded-md bg-white/5 px-2.5 py-1 text-white/50" style={{ fontSize: 11 }}>{r.prompt_metadata.service}</span>
              <span className="rounded-md bg-white/5 px-2.5 py-1 text-white/50" style={{ fontSize: 11 }}>{r.prompt_metadata.language}</span>
              <span className={`rounded-md px-2.5 py-1 ${
                r.prompt_metadata.difficulty === "basic" ? "bg-emerald-500/10 text-emerald-400/70" :
                r.prompt_metadata.difficulty === "intermediate" ? "bg-amber-500/10 text-amber-400/70" :
                "bg-red-500/10 text-red-400/70"
              }`} style={{ fontSize: 11 }}>{r.prompt_metadata.difficulty}</span>
            </div>
          </div>
          <div className={`rounded-xl border px-6 py-3 text-center ${r.success ? "border-emerald-500/20 bg-emerald-500/10" : "border-red-500/20 bg-red-500/10"}`}>
            <div className={`${scoreColor}`} style={{ ...mono, fontSize: 32 }}>{r.review.overall_score}</div>
            <div className="text-white/30" style={{ fontSize: 11 }}>/ {r.review.max_score}</div>
          </div>
        </div>

        {/* Error banner */}
        {!r.success && r.error && (
          <div className="mb-6 rounded-xl border border-red-500/20 bg-red-500/[0.05] p-4">
            <div className="mb-1 flex items-center gap-2">
              <AlertTriangle className="h-4 w-4 text-red-400" />
              <span className="text-red-400" style={{ fontSize: 13 }}>{r.error}</span>
            </div>
            <p className="ml-6 text-white/40" style={{ fontSize: 12 }}>{r.failure_reason}</p>
            <div className="mt-2 ml-6 flex items-center gap-2">
              <code className="rounded bg-white/5 px-2 py-0.5 text-white/40" style={{ ...mono, fontSize: 11 }}>{r.rerun_command}</code>
              <CopyButton text={r.rerun_command} />
            </div>
          </div>
        )}

        {/* Stat cards */}
        <div className="mb-6 grid grid-cols-2 gap-3 md:grid-cols-6">
          {[
            { label: "Total", value: `${r.duration_seconds.toFixed(1)}s`, icon: Clock, color: "text-white/60" },
            { label: "Generation", value: `${r.generation_duration_seconds.toFixed(1)}s`, icon: Cpu, color: "text-blue-400" },
            { label: "Build", value: `${r.build_duration_seconds.toFixed(1)}s`, icon: Terminal, color: "text-amber-400" },
            { label: "Review", value: `${r.review_duration_seconds.toFixed(1)}s`, icon: MessageSquare, color: "text-purple-400" },
            { label: "Files", value: r.generated_files.length, icon: FileCode2, color: "text-emerald-400" },
            { label: "Turns", value: r.environment.turn_count, icon: Zap, color: "text-pink-400" },
          ].map(s => (
            <div key={s.label} className="rounded-lg border border-white/8 bg-white/[0.03] p-3">
              <div className="mb-1 flex items-center gap-1.5">
                <s.icon className={`h-3 w-3 ${s.color}`} />
                <span className="text-white/30" style={{ fontSize: 10 }}>{s.label}</span>
              </div>
              <span className="text-white" style={{ ...mono, fontSize: 16 }}>{s.value}</span>
            </div>
          ))}
        </div>

        {/* Environment info */}
        <div className="mb-6 rounded-xl border border-white/8 bg-white/[0.03] p-5">
          <h3 className="mb-3 text-white" style={{ fontSize: 14 }}>Environment</h3>
          <div className="grid gap-x-8 gap-y-2 sm:grid-cols-2 md:grid-cols-3" style={{ fontSize: 12 }}>
            <div><span className="text-white/30">Model:</span> <span className="text-white/60" style={mono}>{r.environment.model}</span></div>
            <div><span className="text-white/30">Input Tokens:</span> <span className="text-white/60" style={mono}>{r.environment.total_input_tokens.toLocaleString()}</span></div>
            <div><span className="text-white/30">Output Tokens:</span> <span className="text-white/60" style={mono}>{r.environment.total_output_tokens.toLocaleString()}</span></div>
            <div><span className="text-white/30">MCP Servers:</span> <span className="text-white/60">{r.environment.mcp_servers.join(", ")}</span></div>
            <div><span className="text-white/30">Skills Invoked:</span> <span className="text-white/60">{r.environment.skills_invoked.join(", ")}</span></div>
            <div><span className="text-white/30">SDK Package:</span> <span className="text-blue-400/70" style={mono}>{r.prompt_metadata.sdk_package}</span></div>
          </div>
        </div>

        {/* Tool Usage */}
        <div className="mb-6 rounded-xl border border-white/8 bg-white/[0.03] p-5">
          <h3 className="mb-3 flex items-center gap-2 text-white" style={{ fontSize: 14 }}>
            <Wrench className="h-4 w-4 text-white/40" /> Tool Usage
          </h3>
          <div className="grid gap-4 sm:grid-cols-3" style={{ fontSize: 12 }}>
            <div>
              <p className="mb-1.5 text-white/30">Matched ({r.tool_usage.matched.length})</p>
              <div className="flex flex-wrap gap-1">
                {r.tool_usage.matched.map(t => (
                  <span key={t} className="rounded-md bg-emerald-500/10 px-2 py-0.5 text-emerald-400/70" style={{ ...mono, fontSize: 10 }}>{t}</span>
                ))}
              </div>
            </div>
            <div>
              <p className="mb-1.5 text-white/30">Missed ({r.tool_usage.missed.length})</p>
              <div className="flex flex-wrap gap-1">
                {r.tool_usage.missed.length > 0 ? r.tool_usage.missed.map(t => (
                  <span key={t} className="rounded-md bg-red-500/10 px-2 py-0.5 text-red-400/70" style={{ ...mono, fontSize: 10 }}>{t}</span>
                )) : <span className="text-white/20">None</span>}
              </div>
            </div>
            <div>
              <p className="mb-1.5 text-white/30">Extra ({r.tool_usage.extra.length})</p>
              <div className="flex flex-wrap gap-1">
                {r.tool_usage.extra.length > 0 ? r.tool_usage.extra.map(t => (
                  <span key={t} className="rounded-md bg-amber-500/10 px-2 py-0.5 text-amber-400/70" style={{ ...mono, fontSize: 10 }}>{t}</span>
                )) : <span className="text-white/20">None</span>}
              </div>
            </div>
          </div>
        </div>

        {/* Review Panel */}
        <div className="mb-6 rounded-xl border border-white/8 bg-white/[0.03] p-5">
          <h3 className="mb-4 text-white" style={{ fontSize: 14 }}>Review Panel ({r.review_panel.length} reviewers)</h3>

          {/* Consolidated */}
          <div className="mb-4 rounded-lg border border-white/8 bg-white/[0.03] p-4">
            <div className="mb-2 flex items-center justify-between">
              <span className="text-white/50" style={{ fontSize: 13 }}>Consolidated Review</span>
              <span className={scoreColor} style={{ ...mono, fontSize: 16 }}>{r.review.overall_score}/{r.review.max_score}</span>
            </div>
            <p className="mb-3 text-white/45" style={{ fontSize: 13, lineHeight: 1.6 }}>{r.review.summary}</p>
            <div className="grid gap-4 sm:grid-cols-2">
              <div>
                <p className="mb-1.5 text-emerald-400/60" style={{ fontSize: 11 }}>Strengths</p>
                {r.review.strengths.map((s, i) => (
                  <div key={i} className="mb-1 flex gap-1.5">
                    <CheckCircle2 className="mt-0.5 h-3 w-3 shrink-0 text-emerald-400/50" />
                    <span className="text-white/50" style={{ fontSize: 12 }}>{s}</span>
                  </div>
                ))}
              </div>
              <div>
                <p className="mb-1.5 text-red-400/60" style={{ fontSize: 11 }}>Issues</p>
                {r.review.issues.map((s, i) => (
                  <div key={i} className="mb-1 flex gap-1.5">
                    <XCircle className="mt-0.5 h-3 w-3 shrink-0 text-red-400/50" />
                    <span className="text-white/50" style={{ fontSize: 12 }}>{s}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Individual reviewers */}
          <div className="space-y-2">
            {r.review_panel.map((rev, i) => (
              <div key={i} className="rounded-lg border border-white/5 bg-white/[0.02]">
                <button
                  onClick={() => setExpandedReviewer(expandedReviewer === i ? -1 : i)}
                  className="flex w-full items-center justify-between p-3 text-left"
                >
                  <div className="flex items-center gap-3">
                    <span className="text-white/40" style={{ fontSize: 12 }}>Reviewer {i + 1}</span>
                    <span className="text-white/25" style={{ ...mono, fontSize: 11 }}>{rev.model}</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className={rev.overall_score >= 80 ? "text-emerald-400" : rev.overall_score >= 60 ? "text-amber-400" : "text-red-400"} style={{ ...mono, fontSize: 13 }}>
                      {rev.overall_score}/{rev.max_score}
                    </span>
                    {expandedReviewer === i ? <ChevronDown className="h-3.5 w-3.5 text-white/20" /> : <ChevronRight className="h-3.5 w-3.5 text-white/20" />}
                  </div>
                </button>
                {expandedReviewer === i && (
                  <div className="border-t border-white/5 p-4">
                    <p className="mb-3 text-white/40" style={{ fontSize: 12, lineHeight: 1.6 }}>{rev.summary}</p>
                    {/* Criteria */}
                    <div className="mb-3 flex flex-wrap gap-1.5">
                      {rev.criteria.map(c => (
                        <span key={c.name} className={`flex items-center gap-1 rounded-md px-2 py-0.5 ${c.passed ? "bg-emerald-500/10 text-emerald-400/70" : "bg-red-500/10 text-red-400/70"}`} style={{ fontSize: 10 }}>
                          {c.passed ? <CheckCircle2 className="h-2.5 w-2.5" /> : <XCircle className="h-2.5 w-2.5" />}
                          {c.name}
                        </span>
                      ))}
                    </div>
                    <div className="grid gap-3 sm:grid-cols-2" style={{ fontSize: 12 }}>
                      <div>
                        {rev.strengths.map((s, j) => (
                          <div key={j} className="mb-1 flex gap-1.5 text-white/40">
                            <span className="text-emerald-400/40">+</span> {s}
                          </div>
                        ))}
                      </div>
                      <div>
                        {rev.issues.map((s, j) => (
                          <div key={j} className="mb-1 flex gap-1.5 text-white/40">
                            <span className="text-red-400/40">−</span> {s}
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Session Timeline */}
        <div className="mb-6 rounded-xl border border-white/8 bg-white/[0.03]">
          <button
            onClick={() => setShowTimeline(!showTimeline)}
            className="flex w-full items-center justify-between p-5"
          >
            <h3 className="flex items-center gap-2 text-white" style={{ fontSize: 14 }}>
              <Zap className="h-4 w-4 text-amber-400" />
              Session Timeline ({r.session_events.length} events)
            </h3>
            {showTimeline ? <ChevronDown className="h-4 w-4 text-white/20" /> : <ChevronRight className="h-4 w-4 text-white/20" />}
          </button>
          {showTimeline && (
            <div className="border-t border-white/8 p-5">
              <div className="relative space-y-0">
                <div className="absolute left-[18px] top-2 h-[calc(100%-16px)] w-px bg-white/8" />
                {r.session_events.map((evt, i) => {
                  const typeColors: Record<string, string> = {
                    prompt: "bg-blue-500/20 text-blue-400",
                    reasoning: "bg-purple-500/20 text-purple-400",
                    tool_call: "bg-amber-500/20 text-amber-400",
                    tool_result: "bg-emerald-500/20 text-emerald-400",
                    reply: "bg-cyan-500/20 text-cyan-400",
                    warning: "bg-red-500/20 text-red-400",
                    file_write: "bg-pink-500/20 text-pink-400",
                  };
                  const colorClass = typeColors[evt.type] || "bg-white/10 text-white/50";

                  return (
                    <div key={i} className="relative flex gap-3 py-2">
                      <div className={`relative z-10 flex h-9 w-9 shrink-0 items-center justify-center rounded-lg ${colorClass}`}>
                        <span style={{ ...mono, fontSize: 9 }}>{evt.turn_number}</span>
                      </div>
                      <div className="min-w-0 flex-1 rounded-lg bg-white/[0.02] p-3">
                        <div className="mb-1 flex flex-wrap items-center gap-2">
                          <span className={`rounded px-1.5 py-0.5 ${colorClass}`} style={{ ...mono, fontSize: 10 }}>{evt.type}</span>
                          {evt.tool_name && <span className="text-white/50" style={{ ...mono, fontSize: 11 }}>{evt.tool_name}</span>}
                          {evt.duration_ms > 0 && <span className="text-white/25" style={{ fontSize: 10 }}>{evt.duration_ms.toFixed(0)}ms</span>}
                          {evt.tool_success !== undefined && (
                            evt.tool_success ? <CheckCircle2 className="h-3 w-3 text-emerald-400/60" /> : <XCircle className="h-3 w-3 text-red-400/60" />
                          )}
                        </div>
                        {evt.content && <p className="truncate text-white/35" style={{ fontSize: 11 }}>{evt.content}</p>}
                        {evt.file_path && (
                          <span className="text-pink-400/50" style={{ ...mono, fontSize: 10 }}>{evt.file_operation}: {evt.file_path} ({evt.file_size} bytes)</span>
                        )}
                        {(evt.input_tokens || evt.output_tokens) && (
                          <div className="mt-1 flex gap-3">
                            {evt.input_tokens ? <span className="text-white/20" style={{ fontSize: 10 }}>in: {evt.input_tokens}</span> : null}
                            {evt.output_tokens ? <span className="text-white/20" style={{ fontSize: 10 }}>out: {evt.output_tokens}</span> : null}
                          </div>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </div>

        {/* Build */}
        <div className="mb-6 rounded-xl border border-white/8 bg-white/[0.03]">
          <button onClick={() => setShowBuild(!showBuild)} className="flex w-full items-center justify-between p-5">
            <h3 className="flex items-center gap-2 text-white" style={{ fontSize: 14 }}>
              <Terminal className="h-4 w-4 text-white/40" />
              Build Verification
              {r.build.success ? <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400" /> : <XCircle className="h-3.5 w-3.5 text-red-400" />}
            </h3>
            {showBuild ? <ChevronDown className="h-4 w-4 text-white/20" /> : <ChevronRight className="h-4 w-4 text-white/20" />}
          </button>
          {showBuild && (
            <div className="border-t border-white/8 p-5" style={{ fontSize: 12 }}>
              <div className="mb-3 flex gap-4">
                <span className="text-white/30">Command: <code className="text-white/50" style={mono}>{r.build.command}</code></span>
                <span className="text-white/30">Exit: <code className={r.build.exit_code === 0 ? "text-emerald-400" : "text-red-400"} style={mono}>{r.build.exit_code}</code></span>
              </div>
              {r.build.stdout && (
                <pre className="mb-2 rounded-lg bg-black/30 p-3 text-emerald-400/60" style={{ ...mono, fontSize: 11 }}>{r.build.stdout}</pre>
              )}
              {r.build.stderr && (
                <pre className="rounded-lg bg-black/30 p-3 text-red-400/60" style={{ ...mono, fontSize: 11 }}>{r.build.stderr}</pre>
              )}
            </div>
          )}
        </div>

        {/* Generated Files */}
        <div className="mb-6 rounded-xl border border-white/8 bg-white/[0.03]">
          <button onClick={() => setShowFiles(!showFiles)} className="flex w-full items-center justify-between p-5">
            <h3 className="flex items-center gap-2 text-white" style={{ fontSize: 14 }}>
              <FileCode2 className="h-4 w-4 text-white/40" />
              Generated Files ({r.reviewed_files.length})
            </h3>
            {showFiles ? <ChevronDown className="h-4 w-4 text-white/20" /> : <ChevronRight className="h-4 w-4 text-white/20" />}
          </button>
          {showFiles && (
            <div className="border-t border-white/8 p-5 space-y-3">
              {r.reviewed_files.map((f, i) => (
                <div key={i} className="rounded-lg border border-white/5 bg-black/20">
                  <div className="flex items-center justify-between border-b border-white/5 px-4 py-2">
                    <span className="text-emerald-400/70" style={{ ...mono, fontSize: 12 }}>{f.path}</span>
                    <CopyButton text={f.content} />
                  </div>
                  <pre className="overflow-x-auto p-4 text-white/40" style={{ ...mono, fontSize: 11 }}>{f.content}</pre>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
