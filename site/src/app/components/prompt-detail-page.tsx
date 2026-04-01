import { useParams, Link } from "react-router";
import { generateHistoryReport, mockRuns, computePromptCorrelations, type CorrelationStat } from "../data/mock-data";
import { useMemo } from "react";
import { ArrowLeft, CheckCircle2, XCircle, Clock, BarChart3, TrendingUp, Cpu, Wrench, ArrowUpRight, ArrowDownRight, Minus } from "lucide-react";
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, LineChart, Line } from "recharts";

const mono = { fontFamily: "'JetBrains Mono', monospace" };

function DeltaIndicator({ rate, baseline }: { rate: number; baseline: number }) {
  const delta = rate - baseline;
  if (Math.abs(delta) < 1) return <Minus className="h-3 w-3 text-white/20" />;
  if (delta > 0) return (
    <span className="inline-flex items-center gap-0.5 text-emerald-400" style={{ ...mono, fontSize: 11 }}>
      <ArrowUpRight className="h-3 w-3" />+{delta.toFixed(1)}%
    </span>
  );
  return (
    <span className="inline-flex items-center gap-0.5 text-red-400" style={{ ...mono, fontSize: 11 }}>
      <ArrowDownRight className="h-3 w-3" />{delta.toFixed(1)}%
    </span>
  );
}

function CorrelationTable({ title, icon: Icon, data, baseline, showDuration }: {
  title: string;
  icon: React.ComponentType<{ className?: string }>;
  data: CorrelationStat[];
  baseline: number;
  showDuration?: boolean;
}) {
  if (data.length === 0) return null;
  const best = data[0];
  const worst = data[data.length - 1];

  return (
    <div className="rounded-xl border border-white/8 bg-white/[0.03] p-5">
      <div className="mb-1 flex items-center gap-2">
        <Icon className="h-4 w-4 text-white/40" />
        <h3 className="text-white" style={{ fontSize: 14 }}>{title}</h3>
      </div>
      <p className="mb-4 text-white/25" style={{ fontSize: 11 }}>
        How this prompt's pass rate changes based on which {title.toLowerCase().replace("pass rate by ", "")} is in the eval config.
        Overall baseline: <span style={mono}>{baseline}%</span>
      </p>

      {/* Best / Worst callout */}
      {data.length > 1 && (
        <div className="mb-4 flex flex-wrap gap-3">
          <div className="rounded-lg border border-emerald-500/15 bg-emerald-500/[0.05] px-3 py-2">
            <span className="text-white/30" style={{ fontSize: 10 }}>Best</span>
            <div className="flex items-center gap-2">
              <span className="text-emerald-400" style={{ ...mono, fontSize: 13 }}>{best.name}</span>
              <span className="text-emerald-400" style={{ ...mono, fontSize: 13 }}>{best.rate}%</span>
              <DeltaIndicator rate={best.rate} baseline={baseline} />
            </div>
          </div>
          {worst.rate < baseline && (
            <div className="rounded-lg border border-red-500/15 bg-red-500/[0.05] px-3 py-2">
              <span className="text-white/30" style={{ fontSize: 10 }}>Worst</span>
              <div className="flex items-center gap-2">
                <span className="text-red-400" style={{ ...mono, fontSize: 13 }}>{worst.name}</span>
                <span className="text-red-400" style={{ ...mono, fontSize: 13 }}>{worst.rate}%</span>
                <DeltaIndicator rate={worst.rate} baseline={baseline} />
              </div>
            </div>
          )}
        </div>
      )}

      <div className="overflow-x-auto">
        <table className="w-full" style={{ fontSize: 12 }}>
          <thead>
            <tr className="border-b border-white/8">
              <th className="px-3 py-2 text-left text-white/30" style={{ fontWeight: 500, fontSize: 10 }}>Name</th>
              <th className="px-3 py-2 text-left text-white/30" style={{ fontWeight: 500, fontSize: 10 }}>Evals</th>
              <th className="px-3 py-2 text-left text-white/30" style={{ fontWeight: 500, fontSize: 10 }}>Pass Rate</th>
              <th className="px-3 py-2 text-left text-white/30" style={{ fontWeight: 500, fontSize: 10 }}>vs Baseline</th>
              <th className="px-3 py-2 text-left text-white/30" style={{ fontWeight: 500, fontSize: 10 }}>Avg Score</th>
              {showDuration && <th className="px-3 py-2 text-left text-white/30" style={{ fontWeight: 500, fontSize: 10 }}>Avg Duration</th>}
            </tr>
          </thead>
          <tbody>
            {data.map(d => {
              const rateColor = d.rate >= 80 ? "text-emerald-400" : d.rate >= 60 ? "text-amber-400" : "text-red-400";
              return (
                <tr key={d.name} className="border-b border-white/5 transition hover:bg-white/[0.02]">
                  <td className="px-3 py-2.5 text-white/70" style={{ ...mono, fontSize: 11 }}>{d.name}</td>
                  <td className="px-3 py-2.5 text-white/40" style={{ ...mono, fontSize: 11 }}>{d.total}</td>
                  <td className="px-3 py-2.5">
                    <div className="flex items-center gap-2">
                      <div className="h-1.5 w-14 overflow-hidden rounded-full bg-white/10">
                        <div className={`h-full rounded-full ${d.rate >= 80 ? "bg-emerald-500" : d.rate >= 60 ? "bg-amber-500" : "bg-red-500"}`} style={{ width: `${d.rate}%` }} />
                      </div>
                      <span className={rateColor} style={{ ...mono, fontSize: 11 }}>{d.rate}%</span>
                    </div>
                  </td>
                  <td className="px-3 py-2.5">
                    <DeltaIndicator rate={d.rate} baseline={baseline} />
                  </td>
                  <td className="px-3 py-2.5">
                    <span className={d.avgScore >= 80 ? "text-emerald-400/70" : d.avgScore >= 60 ? "text-amber-400/70" : "text-red-400/70"} style={{ ...mono, fontSize: 11 }}>
                      {d.avgScore}
                    </span>
                  </td>
                  {showDuration && <td className="px-3 py-2.5 text-white/40" style={{ ...mono, fontSize: 11 }}>{d.avgDuration}s</td>}
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export function PromptDetailPage() {
  const { promptId } = useParams();
  const decodedId = decodeURIComponent(promptId || "");
  const history = useMemo(() => generateHistoryReport(decodedId), [decodedId]);
  const correlations = useMemo(() => computePromptCorrelations(decodedId), [decodedId]);

  // Find metadata from any run
  const metadata = useMemo(() => {
    for (const run of mockRuns) {
      const found = run.results.find(r => r.prompt_id === decodedId);
      if (found) return found.prompt_metadata;
    }
    return null;
  }, [decodedId]);

  const configChartData = history.configs.map(c => ({
    name: c.config.replace("baseline-", "").replace("copilot-", "cp-"),
    rate: c.pass_rate,
    runs: c.runs,
    avgDuration: c.avg_duration,
  }));

  // Entries over time (by run_id)
  const timelineData = history.entries.reduce<Record<string, { run: string; passed: number; failed: number; avgScore: number; count: number }>>((acc, e) => {
    if (!acc[e.run_id]) acc[e.run_id] = { run: e.run_id.slice(0, 8), passed: 0, failed: 0, avgScore: 0, count: 0 };
    acc[e.run_id].count++;
    acc[e.run_id].avgScore += e.score;
    if (e.success) acc[e.run_id].passed++;
    else acc[e.run_id].failed++;
    return acc;
  }, {});
  const timelineChartData = Object.values(timelineData).map(d => ({
    ...d,
    avgScore: Math.round(d.avgScore / d.count),
    rate: Math.round((d.passed / d.count) * 100),
  }));

  const diffColor = metadata?.difficulty === "basic" ? "bg-emerald-500/10 text-emerald-400" :
    metadata?.difficulty === "intermediate" ? "bg-amber-500/10 text-amber-400" : "bg-red-500/10 text-red-400";

  return (
    <div className="min-h-screen bg-[#0a0a0f] px-4 py-8 sm:px-6" style={{ fontFamily: "'Inter', sans-serif" }}>
      <div className="mx-auto max-w-6xl">
        <Link to="/prompts" className="mb-6 inline-flex items-center gap-1.5 text-white/40 no-underline transition hover:text-emerald-400" style={{ fontSize: 13 }}>
          <ArrowLeft className="h-3.5 w-3.5" /> All Prompts
        </Link>

        {/* Header */}
        <div className="mb-8">
          <h1 className="mb-3 text-white" style={{ ...mono, fontSize: "clamp(1.2rem, 2.5vw, 1.6rem)" }}>
            {decodedId}
          </h1>
          {metadata && (
            <div className="flex flex-wrap gap-2">
              <span className="rounded-md bg-white/5 px-2.5 py-1 text-white/50" style={{ fontSize: 12 }}>{metadata.service}</span>
              <span className="rounded-md bg-white/5 px-2.5 py-1 text-white/50" style={{ fontSize: 12 }}>{metadata.language}</span>
              <span className={`rounded-md px-2.5 py-1 ${diffColor}`} style={{ fontSize: 12 }}>{metadata.difficulty}</span>
              <span className="rounded-md bg-white/5 px-2.5 py-1 text-white/40" style={{ fontSize: 12 }}>{metadata.plane}</span>
              <span className="rounded-md bg-white/5 px-2.5 py-1 text-white/40" style={{ fontSize: 12 }}>{metadata.category}</span>
              <span className="rounded-md bg-blue-500/10 px-2.5 py-1 text-blue-400/70" style={{ ...mono, fontSize: 11 }}>{metadata.sdk_package}</span>
            </div>
          )}
        </div>

        {/* Summary cards */}
        <div className="mb-8 grid grid-cols-2 gap-3 md:grid-cols-4">
          {[
            { label: "Total Runs", value: history.total_runs, icon: BarChart3, color: "text-blue-400" },
            { label: "Pass Rate", value: `${history.pass_rate}%`, icon: CheckCircle2, color: history.pass_rate >= 80 ? "text-emerald-400" : history.pass_rate >= 60 ? "text-amber-400" : "text-red-400" },
            { label: "Passed", value: history.passed, icon: CheckCircle2, color: "text-emerald-400" },
            { label: "Avg Duration", value: `${history.avg_duration_seconds}s`, icon: Clock, color: "text-purple-400" },
          ].map(s => (
            <div key={s.label} className="rounded-xl border border-white/8 bg-white/[0.03] p-4">
              <div className="mb-2 flex items-center gap-1.5">
                <s.icon className={`h-3.5 w-3.5 ${s.color}`} />
                <span className="text-white/35" style={{ fontSize: 11 }}>{s.label}</span>
              </div>
              <span className={`${s.color}`} style={{ ...mono, fontSize: 22 }}>{s.value}</span>
            </div>
          ))}
        </div>

        {/* Charts */}
        <div className="mb-8 grid gap-6 lg:grid-cols-2">
          <div className="rounded-xl border border-white/8 bg-white/[0.03] p-6">
            <h3 className="mb-4 text-white" style={{ fontSize: 15 }}>Pass Rate by Config</h3>
            <ResponsiveContainer width="100%" height={230}>
              <BarChart data={configChartData}>
                <XAxis dataKey="name" tick={{ fill: "rgba(255,255,255,0.35)", fontSize: 10 }} axisLine={false} tickLine={false} />
                <YAxis domain={[0, 100]} tick={{ fill: "rgba(255,255,255,0.35)", fontSize: 10 }} axisLine={false} tickLine={false} />
                <Tooltip contentStyle={{ background: "#1a1a2e", border: "1px solid rgba(255,255,255,0.1)", borderRadius: 8, color: "#fff", fontSize: 12 }} />
                <Bar dataKey="rate" fill="#10b981" radius={[4, 4, 0, 0]} name="Pass Rate %" />
              </BarChart>
            </ResponsiveContainer>
          </div>

          <div className="rounded-xl border border-white/8 bg-white/[0.03] p-6">
            <h3 className="mb-4 text-white" style={{ fontSize: 15 }}>Score Trend Across Runs</h3>
            <ResponsiveContainer width="100%" height={230}>
              <LineChart data={timelineChartData}>
                <XAxis dataKey="run" tick={{ fill: "rgba(255,255,255,0.35)", fontSize: 10 }} axisLine={false} tickLine={false} />
                <YAxis domain={[0, 100]} tick={{ fill: "rgba(255,255,255,0.35)", fontSize: 10 }} axisLine={false} tickLine={false} />
                <Tooltip contentStyle={{ background: "#1a1a2e", border: "1px solid rgba(255,255,255,0.1)", borderRadius: 8, color: "#fff", fontSize: 12 }} />
                <Line type="monotone" dataKey="avgScore" stroke="#10b981" strokeWidth={2} dot={{ fill: "#10b981", r: 3 }} name="Avg Score" />
                <Line type="monotone" dataKey="rate" stroke="#8b5cf6" strokeWidth={2} dot={{ fill: "#8b5cf6", r: 3 }} name="Pass Rate %" />
              </LineChart>
            </ResponsiveContainer>
            <div className="mt-3 flex justify-center gap-5">
              {[{ label: "Avg Score", color: "#10b981" }, { label: "Pass Rate", color: "#8b5cf6" }].map(l => (
                <div key={l.label} className="flex items-center gap-1.5">
                  <div className="h-2 w-2 rounded-full" style={{ background: l.color }} />
                  <span className="text-white/40" style={{ fontSize: 11 }}>{l.label}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* ── Correlation Insights ── */}
        <div className="mb-8">
          <div className="mb-5 flex items-center gap-3">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-purple-500/15">
              <TrendingUp className="h-4 w-4 text-purple-400" />
            </div>
            <div>
              <h2 className="text-white" style={{ fontSize: 17 }}>Pass Rate Correlations</h2>
              <p className="text-white/30" style={{ fontSize: 12 }}>
                How this prompt performs when specific models or tools are present in the eval config.
              </p>
            </div>
          </div>

          <div className="grid gap-6 lg:grid-cols-2">
            <CorrelationTable
              title="Pass Rate by Model"
              icon={Cpu}
              data={correlations.byModel}
              baseline={correlations.overallRate}
              showDuration
            />
            <CorrelationTable
              title="Pass Rate by Tool Used"
              icon={Wrench}
              data={correlations.byTool}
              baseline={correlations.overallRate}
            />
          </div>
        </div>

        {/* Config breakdown table */}
        <div className="mb-8 rounded-xl border border-white/8 bg-white/[0.03] p-6">
          <h3 className="mb-4 text-white" style={{ fontSize: 15 }}>Config Breakdown</h3>
          <div className="overflow-x-auto">
            <table className="w-full" style={{ fontSize: 13 }}>
              <thead>
                <tr className="border-b border-white/8">
                  {["Config", "Runs", "Passed", "Pass Rate", "Avg Duration"].map(h => (
                    <th key={h} className="px-4 py-2.5 text-left text-white/30" style={{ fontWeight: 500, fontSize: 11 }}>{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {history.configs.map(c => (
                  <tr key={c.config} className="border-b border-white/5">
                    <td className="px-4 py-3 text-emerald-400/80" style={{ ...mono, fontSize: 12 }}>{c.config}</td>
                    <td className="px-4 py-3 text-white/50" style={{ ...mono, fontSize: 12 }}>{c.runs}</td>
                    <td className="px-4 py-3 text-white/50" style={{ ...mono, fontSize: 12 }}>{c.passed}</td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <div className="h-1.5 w-16 overflow-hidden rounded-full bg-white/10">
                          <div className={`h-full rounded-full ${c.pass_rate >= 80 ? "bg-emerald-500" : c.pass_rate >= 60 ? "bg-amber-500" : "bg-red-500"}`} style={{ width: `${c.pass_rate}%` }} />
                        </div>
                        <span className={c.pass_rate >= 80 ? "text-emerald-400" : c.pass_rate >= 60 ? "text-amber-400" : "text-red-400"} style={{ ...mono, fontSize: 12 }}>{c.pass_rate}%</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-white/40" style={{ ...mono, fontSize: 12 }}>{c.avg_duration}s</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* History entries */}
        <div className="rounded-xl border border-white/8 bg-white/[0.03] p-6">
          <h3 className="mb-4 text-white" style={{ fontSize: 15 }}>All Entries</h3>
          <div className="overflow-x-auto">
            <table className="w-full" style={{ fontSize: 13 }}>
              <thead>
                <tr className="border-b border-white/8">
                  {["Status", "Run", "Config", "Score", "Duration", "Files", ""].map(h => (
                    <th key={h} className="px-4 py-2.5 text-left text-white/30" style={{ fontWeight: 500, fontSize: 11 }}>{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {history.entries.map((e, i) => (
                  <tr key={`${e.run_id}-${e.config_name}-${i}`} className="border-b border-white/5 transition hover:bg-white/[0.02]">
                    <td className="px-4 py-3">
                      {e.success ? <CheckCircle2 className="h-4 w-4 text-emerald-400" /> : <XCircle className="h-4 w-4 text-red-400" />}
                    </td>
                    <td className="px-4 py-3">
                      <Link to={`/runs/${e.run_id}`} className="text-blue-400/70 no-underline hover:text-blue-400" style={{ ...mono, fontSize: 11 }}>
                        {e.run_id}
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-white/50" style={{ ...mono, fontSize: 12 }}>{e.config_name}</td>
                    <td className="px-4 py-3">
                      <span className={e.score >= 80 ? "text-emerald-400" : e.score >= 60 ? "text-amber-400" : "text-red-400"} style={{ ...mono, fontSize: 13 }}>
                        {e.score}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-white/40" style={{ ...mono, fontSize: 12 }}>{e.duration.toFixed(1)}s</td>
                    <td className="px-4 py-3 text-white/40" style={{ fontSize: 12 }}>{e.file_count}</td>
                    <td className="px-4 py-3">
                      <Link
                        to={`/eval/${encodeURIComponent(decodedId)}/${encodeURIComponent(e.config_name)}`}
                        className="text-white/30 no-underline transition hover:text-emerald-400"
                        style={{ fontSize: 12 }}
                      >
                        Detail →
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}