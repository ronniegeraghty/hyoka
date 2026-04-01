import { Link } from "react-router";
import { mockRuns } from "../data/mock-data";
import { CheckCircle2, XCircle, AlertTriangle, Clock, ChevronRight, Activity } from "lucide-react";
import { motion } from "motion/react";

function formatDuration(s: number): string {
  if (s < 60) return `${s.toFixed(1)}s`;
  const m = Math.floor(s / 60);
  const sec = (s % 60).toFixed(0);
  return `${m}m ${sec}s`;
}

function formatDate(ts: string): string {
  const d = new Date(ts);
  return d.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric", hour: "2-digit", minute: "2-digit" });
}

export function RunsPage() {
  return (
    <div className="min-h-screen bg-[#0a0a0f] px-4 py-8 sm:px-6" style={{ fontFamily: "'Inter', sans-serif" }}>
      <div className="mx-auto max-w-5xl">
        <div className="mb-8">
          <h1 className="mb-2 text-white" style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: "clamp(1.5rem, 3vw, 2rem)" }}>
            Evaluation Runs
          </h1>
          <p className="text-white/40" style={{ fontSize: 14 }}>
            Browse all evaluation runs and drill into individual results.
          </p>
        </div>

        <div className="space-y-4">
          {mockRuns.map((run, i) => {
            const rate = ((run.passed / run.total_evaluations) * 100).toFixed(1);
            return (
              <motion.div
                key={run.run_id}
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.05 }}
              >
                <Link
                  to={`/runs/${run.run_id}`}
                  className="group block rounded-xl border border-white/8 bg-white/[0.03] p-5 no-underline transition hover:border-emerald-500/20 hover:bg-white/[0.05]"
                >
                  <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                    <div className="flex-1">
                      <div className="mb-1 flex items-center gap-3">
                        <span className="text-emerald-400" style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 15 }}>
                          {run.run_id}
                        </span>
                        <span className="rounded-md bg-white/5 px-2 py-0.5 text-white/30" style={{ fontSize: 11 }}>
                          {run.total_evaluations} evals
                        </span>
                      </div>
                      <p className="text-white/40" style={{ fontSize: 13 }}>
                        {formatDate(run.timestamp)} · {run.total_prompts} prompts · {run.total_configs} configs
                      </p>
                    </div>

                    <div className="flex items-center gap-5">
                      <div className="flex items-center gap-4">
                        <div className="flex items-center gap-1.5">
                          <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400" />
                          <span className="text-emerald-400" style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 13 }}>{run.passed}</span>
                        </div>
                        <div className="flex items-center gap-1.5">
                          <XCircle className="h-3.5 w-3.5 text-red-400" />
                          <span className="text-red-400" style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 13 }}>{run.failed}</span>
                        </div>
                        {run.errors > 0 && (
                          <div className="flex items-center gap-1.5">
                            <AlertTriangle className="h-3.5 w-3.5 text-amber-400" />
                            <span className="text-amber-400" style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 13 }}>{run.errors}</span>
                          </div>
                        )}
                      </div>

                      <div className="hidden items-center gap-2 sm:flex">
                        <div className="h-2 w-24 overflow-hidden rounded-full bg-white/10">
                          <div className="h-full rounded-full bg-emerald-500" style={{ width: `${rate}%` }} />
                        </div>
                        <span className="text-white/50" style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 12 }}>
                          {rate}%
                        </span>
                      </div>

                      <div className="hidden items-center gap-1.5 text-white/30 sm:flex">
                        <Clock className="h-3.5 w-3.5" />
                        <span style={{ fontSize: 12 }}>{formatDuration(run.duration_seconds)}</span>
                      </div>

                      <ChevronRight className="h-4 w-4 text-white/20 transition group-hover:text-emerald-400" />
                    </div>
                  </div>
                </Link>
              </motion.div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
