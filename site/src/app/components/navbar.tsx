import { Link, useLocation } from "react-router";
import { Terminal, Menu, X } from "lucide-react";
import { useState } from "react";

const navLinks = [
  { to: "/", label: "Home" },
  { to: "/how-it-works", label: "How It Works" },
  { to: "/runs", label: "Runs" },
  { to: "/prompts", label: "Prompts" },
  { to: "/dashboard", label: "Dashboard" },
  { to: "/docs", label: "Docs" },
];

export function Navbar() {
  const location = useLocation();
  const [open, setOpen] = useState(false);

  return (
    <nav className="sticky top-0 z-50 border-b border-white/10 bg-[#0a0a0f]/80 backdrop-blur-xl">
      <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4">
        <Link to="/" className="flex items-center gap-2.5 no-underline">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-500/20">
            <Terminal className="h-4 w-4 text-emerald-400" />
          </div>
          <span className="text-white" style={{ fontFamily: "'JetBrains Mono', monospace", fontWeight: 700 }}>
            hyoka
          </span>
        </Link>

        <div className="hidden items-center gap-1 md:flex">
          {navLinks.map((l) => (
            <Link
              key={l.to}
              to={l.to}
              className={`rounded-lg px-3.5 py-2 no-underline transition-colors ${
                (l.to === "/" ? location.pathname === "/" : location.pathname.startsWith(l.to))
                  ? "bg-white/10 text-white"
                  : "text-white/60 hover:bg-white/5 hover:text-white"
              }`}
              style={{ fontSize: 14 }}
            >
              {l.label}
            </Link>
          ))}
        </div>

        <a
          href="https://github.com"
          target="_blank"
          rel="noopener noreferrer"
          className="hidden rounded-lg border border-emerald-500/30 bg-emerald-500/10 px-4 py-2 text-emerald-400 no-underline transition-colors hover:bg-emerald-500/20 md:block"
          style={{ fontSize: 14 }}
        >
          Get Started
        </a>

        <button className="text-white/60 md:hidden" onClick={() => setOpen(!open)}>
          {open ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </div>

      {open && (
        <div className="border-t border-white/10 px-6 py-4 md:hidden">
          {navLinks.map((l) => (
            <Link
              key={l.to}
              to={l.to}
              onClick={() => setOpen(false)}
              className={`block rounded-lg px-3 py-2.5 no-underline ${
                location.pathname === l.to ? "text-white" : "text-white/60"
              }`}
            >
              {l.label}
            </Link>
          ))}
        </div>
      )}
    </nav>
  );
}