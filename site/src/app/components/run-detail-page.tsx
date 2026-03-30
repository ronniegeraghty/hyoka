import { useParams, Link } from "react-router";
import { mockRuns, computeSummaryStats } from "../data/mock-data";
import { CheckCircle2, XCircle, Clock, FileCode2, Cpu, ArrowLeft, TrendingUp, Wrench } from "lucide-react";
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from "recharts";
import { useState } from "react";

const mono = { fontFamily: "'JetBrains Mono', monospace" };

function ScoreBadge({ score, max = 100 }: { score: number; max?: number }) {
  const pct = (score / max) * 100;
  const color = pct >= 80 ? "text-emerald-400" : pct >= 60 ? "text-amber-400" : "text-red-400";
  return <span className={color} style={{ ...mono, fontSize: 13 }}>{score}/{max}</span>;
}

export function RunDetailPage() {
  const { runId } = useParams();
  const run = mockRuns.find(r => r.run_id === runId);
  const [filterStatus, setFilterStatus] = useState<"all" | "pass" | "fail">("all");
  const [filterService, setFilterService] = useState<string>("all");
  const [filterLang, setFilterLang] = useState<string>("all");

  if (!run) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[#0a0a0f]">
        <div className="text-center">
          <p className="mb-4 text-white/50">Run not found: {runId}</p>
          <Link to="/runs" className="text-emerald-400">← Back to runs</Link>
        </div>
      </div>
    );
  }

  const stats = computeSummaryStats(run.results);
  const rate = ((run.passed / run.total_evaluations) * 100).toFixed(1);

  const services = [...new Set(run.results.map(r => r.prompt_metadata.service))];
  const langs = [...new Set(run.results.map(r => r.prompt_metadata.language))];

  const filtered = run.results.filter(r => {
    if (filterStatus === "pass" && !r.success) return false;
    if (filterStatus === "fail" && r.success) return false;
    if (filterService !== "all" && r.prompt_metadata.service !== filterService) return false;
    if (filterLang !== "all" && r.prompt_metadata.language !== filterLang) return false;
    return true;
  });

  const configChartData = stats.config_pass_rates.map(c => ({ name: c.config.replace("baseline-", "").replace("copilot-", "cp-"), rate: c.rate, passed: c.passed, failed: c.failed }));

  return (
    <div className="min-h-screen bg-[#0a0a0f] px-4 py-8 sm:px-6" style={{ fontFamily: "'Inter', sans-serif" }}>
      <div className="mx-auto max-w-7xl">
        {/* Header */}
        <Link to="/runs" className="mb-6 inline-flex items-center gap-1.5 text-white/40 no-underline transition hover:text-emerald-400" style={{ fontSize: 13 }}>
          <ArrowLeft className="h-3.5 w-3.5" /> All Runs
        </Link>

        <div className="mb-8 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h1 className="mb-1 text-white" style={{ ...mono, fontSize: "clamp(1.25rem, 3vw, 1.75rem)" }}>
              Run {run.run_id}
            </h1>
            <p className="text-white/40" style={{ fontSize: 13 }}>
              {new Date(run.timestamp).toLocaleString()} · {run.total_evaluations} evaluations
            </p>
          </div>
          <div className="flex items-center gap-2 rounded-lg border border-emerald-500/20 bg-emerald-500/10 px-4 py-2">
            <span className="text-emerald-400/60" style={{ fontSize: 12 }}>Pass Rate</span>
            <span className="text-emerald-400" style={{ ...mono, fontSize: 20 }}>{rate}%</span>
          </div>
        </div>

        {/* Summary cards */}
        <div className="mb-8 grid grid-cols-2 gap-3 md:grid-cols-5">
          {[
            { label: "Passed", value: run.passed, icon: CheckCircle2, color: "text-emerald-400" },
            { label: "Failed", value: run.failed, icon: XCircle, color: "text-red-400" },
            { label: "Avg Gen", value: `${run.avg_generation_duration_seconds}s`, icon: Cpu, color: "text-blue-400" },
            { label: "Avg Build", value: `${run.avg_build_duration_seconds}s`, icon: FileCode2, color: "text-amber-400" },
            { label: "Avg Review", value: `${run.avg_review_duration_seconds}s`, icon: Clock, color: "text-purple-400" },
          ].map(s => (
            <div key={s.label} className="rounded-xl border border-white/8 bg-white/[0.03] p-4">
              <div className="mb-2 flex items-center gap-1.5">
                <s.icon className={`h-3.5 w-3.5 ${s.color}`} />
                <span className="text-white/35" style={{ fontSize: 11 }}>{s.label}</span>
              </div>
              <span className="text-white" style={{ ...mono, fontSize: 20 }}>{s.value}</span>
            </div>
          ))}
        </div>

        {/* Config pass rates chart + tool stats */}
        <div className="mb-8 grid gap-6 lg:grid-cols-2">
          <div className="rounded-xl border border-white/8 bg-white/[0.03] p-6">
            <h3 className="mb-4 text-white" style={{ fontSize: 15 }}>Pass Rate by Config</h3>
            <ResponsiveContainer width="100%" height={220}>
              <BarChart data={configChartData}>
                <XAxis dataKey="name" tick={{ fill: "rgba(255,255,255,0.35)", fontSize: 10 }} axisLine={false} tickLine={false} />
                <YAxis domain={[0, 100]} tick={{ fill: "rgba(255,255,255,0.35)", fontSize: 10 }} axisLine={false} tickLine={false} />
                <Tooltip contentStyle={{ background: "#1a1a2e", border: "1px solid rgba(255,255,255,0.1)", borderRadius: 8, color: "#fff", fontSize: 12 }} />
                <Bar dataKey="rate" fill="#10b981" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>

          <div className="rounded-xl border border-white/8 bg-white/[0.03] p-6">
            <h3 className="mb-4 flex items-center gap-2 text-white" style={{ fontSize: 15 }}>
              <Wrench className="h-4 w-4 text-white/40" /> Tool Usage Stats
            </h3>
            <div className="space-y-2.5 overflow-y-auto" style={{ maxHeight: 220 }}>
              {stats.tool_stats.map(t => (
                <div key={t.name} className="flex items-center gap-3">
                  <span className="w-28 shrink-0 truncate text-white/50" style={{ ...mono, fontSize: 11 }}>{t.name}</span>
                  <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-white/10">
                    <div className="h-full rounded-full bg-emerald-500/60" style={{ width: `${t.success_rate}%` }} />
                  </div>
                  <span className="w-16 text-right text-white/40" style={{ fontSize: 11 }}>{t.count}× ({t.success_rate}%)</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* AI Analysis */}
        <div className="mb-8 rounded-xl border border-emerald-500/15 bg-emerald-500/[0.03] p-5">
          <div className="mb-2 flex items-center gap-2">
            <TrendingUp className="h-4 w-4 text-emerald-400" />
            <span className="text-emerald-400" style={{ fontSize: 14 }}>AI Analysis</span>
          </div>
          <p className="text-white/50" style={{ fontSize: 13, lineHeight: 1.7 }}>{run.analysis}</p>
        </div>

        {/* Filters */}
        <div className="mb-4 flex flex-wrap gap-2">
          <select
            value={filterStatus}
            onChange={e => setFilterStatus(e.target.value as any)}
            className="rounded-lg border border-white/10 bg-white/5 px-3 py-1.5 text-white/70"
            style={{ fontSize: 12 }}
          >
            <option value="all">All Status</option>
            <option value="pass">Passed</option>
            <option value="fail">Failed</option>
          </select>
          <select
            value={filterService}
            onChange={e => setFilterService(e.target.value)}
            className="rounded-lg border border-white/10 bg-white/5 px-3 py-1.5 text-white/70"
            style={{ fontSize: 12 }}
          >
            <option value="all">All Services</option>
            {services.map(s => <option key={s} value={s}>{s}</option>)}
          </select>
          <select
            value={filterLang}
            onChange={e => setFilterLang(e.target.value)}
            className="rounded-lg border border-white/10 bg-white/5 px-3 py-1.5 text-white/70"
            style={{ fontSize: 12 }}
          >
            <option value="all">All Languages</option>
            {langs.map(l => <option key={l} value={l}>{l}</option>)}
          </select>
          <span className="self-center text-white/30" style={{ fontSize: 12 }}>{filtered.length} results</span>
        </div>

        {/* Results table */}
        <div className="overflow-x-auto rounded-xl border border-white/8 bg-white/[0.03]">
          <table className="w-full" style={{ fontSize: 13 }}>
            <thead>
              <tr className="border-b border-white/8">
                {["Status", "Prompt", "Config", "Service", "Lang", "Difficulty", "Score", "Duration", "Files", ""].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-white/30" style={{ fontWeight: 500, fontSize: 11 }}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {filtered.map((r, i) => (
                <tr key={`${r.prompt_id}-${r.config_name}-${i}`} className="border-b border-white/5 transition hover:bg-white/[0.02]">
                  <td className="px-4 py-3">
                    {r.success ? <CheckCircle2 className="h-4 w-4 text-emerald-400" /> : <XCircle className="h-4 w-4 text-red-400" />}
                  </td>
                  <td className="max-w-[200px] truncate px-4 py-3">
                    <Link to={`/prompts/${r.prompt_id}`} className="text-emerald-400/80 no-underline hover:text-emerald-400" style={{ ...mono, fontSize: 12 }}>
                      {r.prompt_id}
                    </Link>
                  </td>
                  <td className="px-4 py-3 text-white/50" style={{ ...mono, fontSize: 12 }}>{r.config_name}</td>
                  <td className="px-4 py-3">
                    <span className="rounded-md bg-white/5 px-2 py-0.5 text-white/50" style={{ fontSize: 11 }}>{r.prompt_metadata.service}</span>
                  </td>
                  <td className="px-4 py-3 text-white/50" style={{ fontSize: 12 }}>{r.prompt_metadata.language}</td>
                  <td className="px-4 py-3">
                    <span className={`rounded-md px-2 py-0.5 ${
                      r.prompt_metadata.difficulty === "basic" ? "bg-emerald-500/10 text-emerald-400/70" :
                      r.prompt_metadata.difficulty === "intermediate" ? "bg-amber-500/10 text-amber-400/70" :
                      "bg-red-500/10 text-red-400/70"
                    }`} style={{ fontSize: 11 }}>
                      {r.prompt_metadata.difficulty}
                    </span>
                  </td>
                  <td className="px-4 py-3"><ScoreBadge score={r.review.overall_score} max={r.review.max_score} /></td>
                  <td className="px-4 py-3 text-white/40" style={{ ...mono, fontSize: 12 }}>{r.duration_seconds.toFixed(1)}s</td>
                  <td className="px-4 py-3 text-white/40" style={{ fontSize: 12 }}>{r.generated_files.length}</td>
                  <td className="px-4 py-3">
                    <Link
                      to={`/eval/${encodeURIComponent(r.prompt_id)}/${encodeURIComponent(r.config_name)}`}
                      className="text-white/30 no-underline transition hover:text-emerald-400"
                      style={{ fontSize: 12 }}
                    >
                      View →
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
