import { useState, useMemo } from "react";
import { Link } from "react-router";
import { getAllPrompts } from "../data/mock-data";
import { Search, Filter, ChevronRight } from "lucide-react";
import { motion } from "motion/react";

const mono = { fontFamily: "'JetBrains Mono', monospace" };

export function PromptsPage() {
  const allPrompts = useMemo(() => getAllPrompts(), []);
  const [search, setSearch] = useState("");
  const [filterService, setFilterService] = useState("all");
  const [filterLang, setFilterLang] = useState("all");
  const [filterDifficulty, setFilterDifficulty] = useState("all");
  const [filterPlane, setFilterPlane] = useState("all");

  const services = [...new Set(allPrompts.map(p => p.metadata.service))].sort();
  const langs = [...new Set(allPrompts.map(p => p.metadata.language))].sort();

  const filtered = allPrompts.filter(p => {
    if (search && !p.prompt_id.toLowerCase().includes(search.toLowerCase())) return false;
    if (filterService !== "all" && p.metadata.service !== filterService) return false;
    if (filterLang !== "all" && p.metadata.language !== filterLang) return false;
    if (filterDifficulty !== "all" && p.metadata.difficulty !== filterDifficulty) return false;
    if (filterPlane !== "all" && p.metadata.plane !== filterPlane) return false;
    return true;
  });

  return (
    <div className="min-h-screen bg-[#0a0a0f] px-4 py-8 sm:px-6" style={{ fontFamily: "'Inter', sans-serif" }}>
      <div className="mx-auto max-w-6xl">
        <div className="mb-8">
          <h1 className="mb-2 text-white" style={{ ...mono, fontSize: "clamp(1.5rem, 3vw, 2rem)" }}>
            Prompt Explorer
          </h1>
          <p className="text-white/40" style={{ fontSize: 14 }}>
            Browse and filter all evaluation prompts. Click any prompt to see its history across runs.
          </p>
        </div>

        {/* Search & Filters */}
        <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-white/30" />
            <input
              type="text"
              value={search}
              onChange={e => setSearch(e.target.value)}
              placeholder="Search prompts..."
              className="w-full rounded-lg border border-white/10 bg-white/5 py-2 pl-10 pr-4 text-white placeholder-white/30 outline-none focus:border-emerald-500/30"
              style={{ fontSize: 13 }}
            />
          </div>
          <div className="flex flex-wrap gap-2">
            <select value={filterService} onChange={e => setFilterService(e.target.value)}
              className="rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-white/70" style={{ fontSize: 12 }}>
              <option value="all">All Services</option>
              {services.map(s => <option key={s} value={s}>{s}</option>)}
            </select>
            <select value={filterLang} onChange={e => setFilterLang(e.target.value)}
              className="rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-white/70" style={{ fontSize: 12 }}>
              <option value="all">All Languages</option>
              {langs.map(l => <option key={l} value={l}>{l}</option>)}
            </select>
            <select value={filterDifficulty} onChange={e => setFilterDifficulty(e.target.value)}
              className="rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-white/70" style={{ fontSize: 12 }}>
              <option value="all">All Difficulty</option>
              <option value="basic">Basic</option>
              <option value="intermediate">Intermediate</option>
              <option value="advanced">Advanced</option>
            </select>
            <select value={filterPlane} onChange={e => setFilterPlane(e.target.value)}
              className="rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-white/70" style={{ fontSize: 12 }}>
              <option value="all">All Planes</option>
              <option value="data-plane">Data Plane</option>
              <option value="management-plane">Management Plane</option>
            </select>
          </div>
        </div>

        <p className="mb-4 text-white/30" style={{ fontSize: 12 }}>{filtered.length} prompts found</p>

        {/* Prompt grid */}
        <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
          {filtered.map((p, i) => {
            const rateColor = p.passRate >= 80 ? "text-emerald-400" : p.passRate >= 60 ? "text-amber-400" : "text-red-400";
            const rateBg = p.passRate >= 80 ? "bg-emerald-500/10" : p.passRate >= 60 ? "bg-amber-500/10" : "bg-red-500/10";
            const diffColor = p.metadata.difficulty === "basic" ? "bg-emerald-500/10 text-emerald-400/70" :
              p.metadata.difficulty === "intermediate" ? "bg-amber-500/10 text-amber-400/70" : "bg-red-500/10 text-red-400/70";

            return (
              <motion.div
                key={p.prompt_id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: Math.min(i * 0.02, 0.3) }}
              >
                <Link
                  to={`/prompts/${encodeURIComponent(p.prompt_id)}`}
                  className="group block rounded-xl border border-white/8 bg-white/[0.03] p-4 no-underline transition hover:border-emerald-500/20 hover:bg-white/[0.05]"
                >
                  <div className="mb-3 flex items-start justify-between">
                    <span className="text-emerald-400/80" style={{ ...mono, fontSize: 12 }}>
                      {p.prompt_id}
                    </span>
                    <ChevronRight className="h-3.5 w-3.5 text-white/15 transition group-hover:text-emerald-400" />
                  </div>

                  <div className="mb-3 flex flex-wrap gap-1.5">
                    <span className="rounded-md bg-white/5 px-2 py-0.5 text-white/50" style={{ fontSize: 10 }}>{p.metadata.service}</span>
                    <span className="rounded-md bg-white/5 px-2 py-0.5 text-white/50" style={{ fontSize: 10 }}>{p.metadata.language}</span>
                    <span className={`rounded-md px-2 py-0.5 ${diffColor}`} style={{ fontSize: 10 }}>{p.metadata.difficulty}</span>
                    <span className="rounded-md bg-white/5 px-2 py-0.5 text-white/40" style={{ fontSize: 10 }}>{p.metadata.plane}</span>
                  </div>

                  <div className="flex items-center justify-between">
                    <span className="text-white/30" style={{ fontSize: 11 }}>{p.evalCount} evals</span>
                    <div className="flex items-center gap-2">
                      <div className="h-1.5 w-14 overflow-hidden rounded-full bg-white/10">
                        <div className={`h-full rounded-full ${p.passRate >= 80 ? "bg-emerald-500" : p.passRate >= 60 ? "bg-amber-500" : "bg-red-500"}`} style={{ width: `${p.passRate}%` }} />
                      </div>
                      <span className={rateColor} style={{ ...mono, fontSize: 12 }}>{p.passRate}%</span>
                    </div>
                  </div>

                  <div className="mt-2 flex flex-wrap gap-1">
                    {p.metadata.tags.slice(0, 3).map(t => (
                      <span key={t} className="rounded bg-white/[0.04] px-1.5 py-0.5 text-white/25" style={{ fontSize: 9 }}>{t}</span>
                    ))}
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
