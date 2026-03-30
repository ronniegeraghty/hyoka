import { useState } from "react";
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, LineChart, Line, RadarChart, Radar, PolarGrid, PolarAngleAxis, PolarRadiusAxis } from "recharts";
import { CheckCircle2, XCircle, Clock, FileCode2, Cpu, TrendingUp, Filter } from "lucide-react";

const passRateByService = [
  { name: "Storage", rate: 87 },
  { name: "Key Vault", rate: 79 },
  { name: "Cosmos DB", rate: 72 },
  { name: "Event Hubs", rate: 68 },
  { name: "Service Bus", rate: 74 },
  { name: "Identity", rate: 91 },
];

const passRateByLang = [
  { name: "Python", rate: 88 },
  { name: "JavaScript", rate: 84 },
  { name: "Go", rate: 71 },
  { name: ".NET", rate: 82 },
  { name: "Java", rate: 76 },
  { name: "TypeScript", rate: 85 },
];

const durationTrend = [
  { run: "Run 1", gen: 12.3, build: 4.1, review: 8.7 },
  { run: "Run 2", gen: 11.8, build: 3.9, review: 7.2 },
  { run: "Run 3", gen: 10.5, build: 4.3, review: 6.8 },
  { run: "Run 4", gen: 9.8, build: 3.7, review: 6.1 },
  { run: "Run 5", gen: 9.2, build: 3.5, review: 5.9 },
  { run: "Run 6", gen: 8.7, build: 3.3, review: 5.4 },
];

const radarData = [
  { criteria: "Correctness", "GPT-4o": 88, "Claude 3.5": 91, "Copilot": 82 },
  { criteria: "Completeness", "GPT-4o": 85, "Claude 3.5": 87, "Copilot": 79 },
  { criteria: "Best Practices", "GPT-4o": 82, "Claude 3.5": 89, "Copilot": 75 },
  { criteria: "Error Handling", "GPT-4o": 78, "Claude 3.5": 84, "Copilot": 71 },
  { criteria: "Security", "GPT-4o": 90, "Claude 3.5": 92, "Copilot": 85 },
  { criteria: "Documentation", "GPT-4o": 76, "Claude 3.5": 80, "Copilot": 73 },
];

const recentEvals = [
  { id: "EVL-0042", prompt: "Create a blob storage client with retry", lang: "Python", model: "GPT-4o", score: 92, pass: true, duration: "8.3s", files: 3 },
  { id: "EVL-0041", prompt: "Implement Key Vault secret rotation", lang: "Go", model: "Claude 3.5", score: 67, pass: false, duration: "14.1s", files: 5 },
  { id: "EVL-0040", prompt: "Cosmos DB paginated query with continuation", lang: ".NET", model: "GPT-4o", score: 85, pass: true, duration: "11.7s", files: 4 },
  { id: "EVL-0039", prompt: "Event Hubs consumer with checkpointing", lang: "Java", model: "Copilot", score: 58, pass: false, duration: "16.2s", files: 6 },
  { id: "EVL-0038", prompt: "Service Bus topic subscription filter", lang: "TypeScript", model: "Claude 3.5", score: 94, pass: true, duration: "7.9s", files: 2 },
  { id: "EVL-0037", prompt: "DefaultAzureCredential with fallback chain", lang: "Python", model: "GPT-4o", score: 96, pass: true, duration: "6.4s", files: 2 },
];

const stats = [
  { label: "Total Evaluations", value: "1,247", icon: FileCode2, color: "text-blue-400" },
  { label: "Overall Pass Rate", value: "78.3%", icon: CheckCircle2, color: "text-emerald-400" },
  { label: "Avg Duration", value: "9.8s", icon: Clock, color: "text-amber-400" },
  { label: "Models Tested", value: "6", icon: Cpu, color: "text-purple-400" },
];

export function DashboardPage() {
  const [activeChart, setActiveChart] = useState<"service" | "language">("service");

  return (
    <div className="min-h-screen bg-[#0a0a0f] px-4 py-8 sm:px-6" style={{ fontFamily: "'Inter', sans-serif" }}>
      <div className="mx-auto max-w-7xl">
        <div className="mb-8 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="text-white" style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: "clamp(1.5rem, 3vw, 2rem)" }}>
              Evaluation Dashboard
            </h1>
            <p className="text-white/40" style={{ fontSize: 14 }}>Last updated: March 29, 2026 · 14:32 UTC</p>
          </div>
          <button className="inline-flex items-center gap-2 self-start rounded-lg border border-white/10 bg-white/5 px-4 py-2 text-white/60 transition hover:bg-white/10" style={{ fontSize: 13 }}>
            <Filter className="h-3.5 w-3.5" /> Filters
          </button>
        </div>

        {/* Stats */}
        <div className="mb-8 grid grid-cols-2 gap-4 lg:grid-cols-4">
          {stats.map((s) => (
            <div key={s.label} className="rounded-xl border border-white/8 bg-white/[0.03] p-5">
              <div className="mb-3 flex items-center gap-2">
                <s.icon className={`h-4 w-4 ${s.color}`} />
                <span className="text-white/40" style={{ fontSize: 12 }}>{s.label}</span>
              </div>
              <span className="text-white" style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 24 }}>{s.value}</span>
            </div>
          ))}
        </div>

        {/* Charts */}
        <div className="mb-8 grid gap-6 lg:grid-cols-2">
          {/* Pass Rate */}
          <div className="rounded-xl border border-white/8 bg-white/[0.03] p-6">
            <div className="mb-4 flex items-center justify-between">
              <h3 className="text-white" style={{ fontSize: 15 }}>Pass Rate</h3>
              <div className="flex gap-1 rounded-lg bg-white/5 p-0.5">
                {(["service", "language"] as const).map((t) => (
                  <button
                    key={t}
                    onClick={() => setActiveChart(t)}
                    className={`rounded-md px-3 py-1 capitalize transition ${activeChart === t ? "bg-emerald-500/20 text-emerald-400" : "text-white/40 hover:text-white/60"}`}
                    style={{ fontSize: 12 }}
                  >
                    {t}
                  </button>
                ))}
              </div>
            </div>
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={activeChart === "service" ? passRateByService : passRateByLang}>
                <XAxis dataKey="name" tick={{ fill: "rgba(255,255,255,0.35)", fontSize: 11 }} axisLine={false} tickLine={false} />
                <YAxis domain={[0, 100]} tick={{ fill: "rgba(255,255,255,0.35)", fontSize: 11 }} axisLine={false} tickLine={false} />
                <Tooltip contentStyle={{ background: "#1a1a2e", border: "1px solid rgba(255,255,255,0.1)", borderRadius: 8, color: "#fff", fontSize: 13 }} />
                <Bar dataKey="rate" fill="#10b981" radius={[6, 6, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>

          {/* Duration Trend */}
          <div className="rounded-xl border border-white/8 bg-white/[0.03] p-6">
            <h3 className="mb-4 text-white" style={{ fontSize: 15 }}>Duration Trends (seconds)</h3>
            <ResponsiveContainer width="100%" height={250}>
              <LineChart data={durationTrend}>
                <XAxis dataKey="run" tick={{ fill: "rgba(255,255,255,0.35)", fontSize: 11 }} axisLine={false} tickLine={false} />
                <YAxis tick={{ fill: "rgba(255,255,255,0.35)", fontSize: 11 }} axisLine={false} tickLine={false} />
                <Tooltip contentStyle={{ background: "#1a1a2e", border: "1px solid rgba(255,255,255,0.1)", borderRadius: 8, color: "#fff", fontSize: 13 }} />
                <Line type="monotone" dataKey="gen" stroke="#10b981" strokeWidth={2} dot={false} name="Generation" />
                <Line type="monotone" dataKey="build" stroke="#f59e0b" strokeWidth={2} dot={false} name="Build" />
                <Line type="monotone" dataKey="review" stroke="#8b5cf6" strokeWidth={2} dot={false} name="Review" />
              </LineChart>
            </ResponsiveContainer>
            <div className="mt-3 flex justify-center gap-5">
              {[{ label: "Generation", color: "#10b981" }, { label: "Build", color: "#f59e0b" }, { label: "Review", color: "#8b5cf6" }].map((l) => (
                <div key={l.label} className="flex items-center gap-1.5">
                  <div className="h-2 w-2 rounded-full" style={{ background: l.color }} />
                  <span className="text-white/40" style={{ fontSize: 11 }}>{l.label}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Radar + Table */}
        <div className="mb-8 grid gap-6 lg:grid-cols-5">
          {/* Radar */}
          <div className="rounded-xl border border-white/8 bg-white/[0.03] p-6 lg:col-span-2">
            <h3 className="mb-4 text-white" style={{ fontSize: 15 }}>Model Comparison by Criteria</h3>
            <ResponsiveContainer width="100%" height={280}>
              <RadarChart data={radarData}>
                <PolarGrid stroke="rgba(255,255,255,0.08)" />
                <PolarAngleAxis dataKey="criteria" tick={{ fill: "rgba(255,255,255,0.4)", fontSize: 10 }} />
                <PolarRadiusAxis domain={[0, 100]} tick={false} axisLine={false} />
                <Radar name="GPT-4o" dataKey="GPT-4o" stroke="#10b981" fill="#10b981" fillOpacity={0.15} strokeWidth={2} />
                <Radar name="Claude 3.5" dataKey="Claude 3.5" stroke="#8b5cf6" fill="#8b5cf6" fillOpacity={0.1} strokeWidth={2} />
                <Radar name="Copilot" dataKey="Copilot" stroke="#f59e0b" fill="#f59e0b" fillOpacity={0.08} strokeWidth={2} />
                <Tooltip contentStyle={{ background: "#1a1a2e", border: "1px solid rgba(255,255,255,0.1)", borderRadius: 8, color: "#fff", fontSize: 13 }} />
              </RadarChart>
            </ResponsiveContainer>
            <div className="mt-2 flex justify-center gap-5">
              {[{ label: "GPT-4o", color: "#10b981" }, { label: "Claude 3.5", color: "#8b5cf6" }, { label: "Copilot", color: "#f59e0b" }].map((l) => (
                <div key={l.label} className="flex items-center gap-1.5">
                  <div className="h-2 w-2 rounded-full" style={{ background: l.color }} />
                  <span className="text-white/40" style={{ fontSize: 11 }}>{l.label}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Recent evals table */}
          <div className="rounded-xl border border-white/8 bg-white/[0.03] p-6 lg:col-span-3">
            <h3 className="mb-4 text-white" style={{ fontSize: 15 }}>Recent Evaluations</h3>
            <div className="overflow-x-auto">
              <table className="w-full" style={{ fontSize: 13 }}>
                <thead>
                  <tr className="border-b border-white/8">
                    {["ID", "Prompt", "Lang", "Model", "Score", "Status", "Time"].map((h) => (
                      <th key={h} className="px-3 py-2.5 text-left text-white/30" style={{ fontWeight: 500, fontSize: 11 }}>
                        {h}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {recentEvals.map((e) => (
                    <tr key={e.id} className="border-b border-white/5 transition hover:bg-white/[0.02]">
                      <td className="px-3 py-3 text-emerald-400/70" style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 12 }}>{e.id}</td>
                      <td className="max-w-[200px] truncate px-3 py-3 text-white/70">{e.prompt}</td>
                      <td className="px-3 py-3 text-white/50">{e.lang}</td>
                      <td className="px-3 py-3 text-white/50">{e.model}</td>
                      <td className="px-3 py-3" style={{ fontFamily: "'JetBrains Mono', monospace" }}>
                        <span className={e.score >= 80 ? "text-emerald-400" : e.score >= 60 ? "text-amber-400" : "text-red-400"}>
                          {e.score}
                        </span>
                      </td>
                      <td className="px-3 py-3">
                        {e.pass ? (
                          <CheckCircle2 className="h-4 w-4 text-emerald-400" />
                        ) : (
                          <XCircle className="h-4 w-4 text-red-400" />
                        )}
                      </td>
                      <td className="px-3 py-3 text-white/40" style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 12 }}>{e.duration}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>

        {/* AI Insights */}
        <div className="rounded-xl border border-emerald-500/15 bg-emerald-500/[0.03] p-6">
          <div className="mb-3 flex items-center gap-2">
            <TrendingUp className="h-4 w-4 text-emerald-400" />
            <h3 className="text-emerald-400" style={{ fontSize: 15 }}>AI-Generated Insights</h3>
          </div>
          <div className="space-y-2 text-white/50" style={{ fontSize: 14, lineHeight: 1.7 }}>
            <p>• <strong className="text-white/70">Identity service prompts</strong> consistently achieve the highest pass rates (91%) across all models, suggesting auth patterns are well-represented in training data.</p>
            <p>• <strong className="text-white/70">Event Hubs</strong> shows the lowest pass rate (68%), particularly for advanced streaming scenarios. Most failures relate to incorrect checkpoint store configuration.</p>
            <p>• <strong className="text-white/70">Generation duration</strong> has decreased 29% over the last 6 runs, correlating with improved prompt specificity in recent batches.</p>
            <p>• <strong className="text-white/70">Claude 3.5</strong> leads in best practices and security criteria, while GPT-4o shows stronger performance in completeness for multi-file generation tasks.</p>
          </div>
        </div>
      </div>
    </div>
  );
}
