import { Terminal } from "lucide-react";

export function Footer() {
  return (
    <footer className="border-t border-white/10 bg-[#0a0a0f]">
      <div className="mx-auto max-w-7xl px-6 py-12">
        <div className="flex flex-col items-center justify-between gap-6 md:flex-row">
          <div className="flex items-center gap-2.5">
            <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-emerald-500/20">
              <Terminal className="h-3.5 w-3.5 text-emerald-400" />
            </div>
            <span className="text-white/80" style={{ fontFamily: "'JetBrains Mono', monospace", fontWeight: 600 }}>
              hyoka
            </span>
          </div>
          <p className="text-white/40" style={{ fontSize: 13 }}>
            Evaluate AI code generation quality for Azure SDKs.
          </p>
          <p className="text-white/30" style={{ fontSize: 12 }}>
            © 2026 hyoka · MIT License
          </p>
        </div>
      </div>
    </footer>
  );
}
