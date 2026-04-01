import { Code2, Bot, ClipboardCheck, GitMerge, BarChart3, FileCode2 } from "lucide-react";
import { motion } from "motion/react";

const pipeline = [
  {
    icon: Code2,
    title: "Prompt Authoring",
    desc: "Define natural language prompts categorized by difficulty (basic, intermediate, advanced) and scenario type (authentication, CRUD, pagination, streaming, error handling).",
    details: ["Categorized by Azure service", "Multiple difficulty tiers", "Covers data & management plane"],
  },
  {
    icon: Bot,
    title: "AI Code Generation",
    desc: "AI agents (like GitHub Copilot) generate Azure SDK code from each prompt. Multiple model/config combinations run in parallel for comparison.",
    details: ["Tracks generation duration", "Records files generated & tool calls", "Captures step-by-step agent reasoning"],
  },
  {
    icon: FileCode2,
    title: "Build Verification",
    desc: "Generated code is compiled/built to verify syntactic correctness. Build success, failure details, and build duration are recorded.",
    details: ["Compile-time validation", "Build duration metrics", "Error log capture"],
  },
  {
    icon: ClipboardCheck,
    title: "Multi-Reviewer Scoring",
    desc: "Multiple AI reviewers independently evaluate the generated code. Each reviewer scores on criteria like correctness, completeness, and best practices.",
    details: ["Independent assessments", "Per-criteria numeric scoring", "Strengths & issues identified"],
  },
  {
    icon: GitMerge,
    title: "Score Consolidation",
    desc: "A consolidator AI merges all reviewer assessments into a final unified score, resolving disagreements and producing a consensus evaluation.",
    details: ["Weighted consensus", "Conflict resolution", "Final pass/fail determination"],
  },
  {
    icon: BarChart3,
    title: "Analysis & Reporting",
    desc: "Aggregated results show pass rates by prompt, config, and model. AI-generated analysis surfaces patterns about where code generation excels or falls short.",
    details: ["Cross-config comparisons", "Duration trend analysis", "Tool usage statistics"],
  },
];

export function HowItWorksPage() {
  return (
    <div className="bg-[#0a0a0f] px-6 py-20" style={{ fontFamily: "'Inter', sans-serif" }}>
      <div className="mx-auto max-w-4xl">
        <div className="mb-16 text-center">
          <h1 className="mb-4 text-white" style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: "clamp(1.75rem, 4vw, 2.5rem)" }}>
            How hyoka Works
          </h1>
          <p className="mx-auto max-w-xl text-white/50" style={{ fontSize: 16, lineHeight: 1.7 }}>
            A six-stage pipeline that takes natural language prompts and produces comprehensive quality evaluations.
          </p>
        </div>

        <div className="relative">
          {/* Vertical line */}
          <div className="absolute left-6 top-0 hidden h-full w-px bg-gradient-to-b from-emerald-500/40 via-emerald-500/20 to-transparent md:block" />

          <div className="space-y-8">
            {pipeline.map((step, i) => (
              <motion.div
                key={step.title}
                initial={{ opacity: 0, x: -20 }}
                whileInView={{ opacity: 1, x: 0 }}
                viewport={{ once: true }}
                transition={{ delay: i * 0.1 }}
                className="relative flex gap-6"
              >
                {/* Dot */}
                <div className="relative z-10 hidden flex-shrink-0 md:block">
                  <div className="flex h-12 w-12 items-center justify-center rounded-xl border border-emerald-500/30 bg-emerald-500/10">
                    <step.icon className="h-5 w-5 text-emerald-400" />
                  </div>
                </div>

                <div className="flex-1 rounded-2xl border border-white/8 bg-white/[0.03] p-6">
                  <div className="mb-1 text-emerald-500/50" style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 12 }}>
                    Stage {i + 1}
                  </div>
                  <h3 className="mb-2 text-white" style={{ fontSize: 18 }}>{step.title}</h3>
                  <p className="mb-4 text-white/45" style={{ fontSize: 14, lineHeight: 1.7 }}>{step.desc}</p>
                  <div className="flex flex-wrap gap-2">
                    {step.details.map((d) => (
                      <span
                        key={d}
                        className="rounded-md border border-white/8 bg-white/[0.04] px-2.5 py-1 text-white/40"
                        style={{ fontSize: 12 }}
                      >
                        {d}
                      </span>
                    ))}
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
