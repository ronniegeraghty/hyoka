import { useState, useEffect } from "react";
import { Book, ChevronRight, Loader2 } from "lucide-react";
import { fetchDocs, fetchDoc } from "../data/api";
import type { DocEntry } from "../data/types";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

const mono = { fontFamily: "'JetBrains Mono', monospace" };

export function DocsPage() {
  const [docs, setDocs] = useState<DocEntry[]>([]);
  const [active, setActive] = useState<string>("");
  const [content, setContent] = useState<string>("");
  const [activeTitle, setActiveTitle] = useState<string>("");
  const [loadingList, setLoadingList] = useState(true);
  const [loadingContent, setLoadingContent] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchDocs()
      .then(list => {
        setDocs(list);
        if (list.length > 0) {
          setActive(list[0].slug);
          setActiveTitle(list[0].title);
        }
      })
      .catch(e => setError(e.message))
      .finally(() => setLoadingList(false));
  }, []);

  useEffect(() => {
    if (!active) return;
    setLoadingContent(true);
    fetchDoc(active)
      .then(doc => {
        setContent(doc.content);
        setActiveTitle(doc.title);
      })
      .catch(e => setContent(`Error loading document: ${e.message}`))
      .finally(() => setLoadingContent(false));
  }, [active]);

  if (loadingList) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[#0a0a0f]">
        <Loader2 className="h-6 w-6 animate-spin text-emerald-400" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[#0a0a0f]">
        <div className="text-center">
          <p className="mb-2 text-red-400">Failed to load documentation</p>
          <p className="text-white/40" style={{ fontSize: 13 }}>{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#0a0a0f]" style={{ fontFamily: "'Inter', sans-serif" }}>
      <div className="mx-auto flex max-w-7xl flex-col md:flex-row">
        {/* Sidebar */}
        <aside className="border-b border-white/10 p-6 md:w-64 md:border-b-0 md:border-r md:py-10">
          <h2 className="mb-4 text-white/30" style={{ fontSize: 11, letterSpacing: "0.1em", textTransform: "uppercase" }}>
            Documentation
          </h2>
          <nav className="flex gap-1 overflow-x-auto md:flex-col">
            {docs.map((d) => (
              <button
                key={d.slug}
                onClick={() => setActive(d.slug)}
                className={`flex items-center gap-2.5 whitespace-nowrap rounded-lg px-3 py-2.5 text-left transition ${
                  active === d.slug ? "bg-emerald-500/10 text-emerald-400" : "text-white/50 hover:bg-white/5 hover:text-white/70"
                }`}
                style={{ fontSize: 14 }}
              >
                <Book className="h-4 w-4 flex-shrink-0" />
                {d.title}
                {active === d.slug && <ChevronRight className="ml-auto hidden h-3 w-3 md:block" />}
              </button>
            ))}
          </nav>
        </aside>

        {/* Content */}
        <main className="flex-1 p-6 md:p-10">
          <div className="max-w-3xl">
            <div className="mb-6 flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-500/10">
                <Book className="h-5 w-5 text-emerald-400" />
              </div>
              <h1 className="text-white" style={{ ...mono, fontSize: 24 }}>
                {activeTitle}
              </h1>
            </div>

            {loadingContent ? (
              <div className="flex justify-center py-12">
                <Loader2 className="h-5 w-5 animate-spin text-emerald-400" />
              </div>
            ) : (
              <div className="prose-custom">
                <ReactMarkdown
                  remarkPlugins={[remarkGfm]}
                  components={{
                    h1: ({ children }) => (
                      <h1 className="mb-4 mt-8 text-white" style={{ ...mono, fontSize: 24 }}>{children}</h1>
                    ),
                    h2: ({ children }) => (
                      <h2 className="mb-3 mt-8 text-white" style={{ fontSize: 18 }}>{children}</h2>
                    ),
                    h3: ({ children }) => (
                      <h3 className="mb-2 mt-6 text-white/90" style={{ fontSize: 16 }}>{children}</h3>
                    ),
                    p: ({ children }) => (
                      <p className="mb-3 text-white/60" style={{ fontSize: 14, lineHeight: 1.8 }}>{children}</p>
                    ),
                    ul: ({ children }) => (
                      <ul className="mb-3 ml-4 list-disc space-y-1">{children}</ul>
                    ),
                    ol: ({ children }) => (
                      <ol className="mb-3 ml-4 list-decimal space-y-1">{children}</ol>
                    ),
                    li: ({ children }) => (
                      <li className="text-white/60" style={{ fontSize: 14, lineHeight: 1.7 }}>{children}</li>
                    ),
                    strong: ({ children }) => (
                      <strong className="text-white/80">{children}</strong>
                    ),
                    a: ({ href, children }) => (
                      <a href={href} className="text-emerald-400 underline decoration-emerald-400/30 transition hover:decoration-emerald-400" target="_blank" rel="noopener noreferrer">{children}</a>
                    ),
                    code: ({ className, children }) => {
                      const isBlock = className?.includes("language-");
                      if (isBlock) {
                        return (
                          <code className="text-emerald-300/70" style={{ ...mono, fontSize: 13 }}>
                            {children}
                          </code>
                        );
                      }
                      return (
                        <code className="rounded bg-white/10 px-1.5 py-0.5 text-emerald-300/80" style={{ ...mono, fontSize: 12 }}>
                          {children}
                        </code>
                      );
                    },
                    pre: ({ children }) => (
                      <pre className="my-4 overflow-x-auto rounded-xl border border-white/8 bg-white/[0.04] p-4">
                        {children}
                      </pre>
                    ),
                    table: ({ children }) => (
                      <div className="my-4 overflow-x-auto">
                        <table className="w-full border-collapse" style={{ fontSize: 13 }}>
                          {children}
                        </table>
                      </div>
                    ),
                    thead: ({ children }) => (
                      <thead className="border-b border-white/10">{children}</thead>
                    ),
                    th: ({ children }) => (
                      <th className="px-3 py-2 text-left text-white/50" style={{ fontSize: 12 }}>{children}</th>
                    ),
                    td: ({ children }) => (
                      <td className="border-t border-white/5 px-3 py-2 text-white/60" style={{ fontSize: 13 }}>{children}</td>
                    ),
                    blockquote: ({ children }) => (
                      <blockquote className="my-3 border-l-2 border-emerald-500/30 pl-4 text-white/50 italic">
                        {children}
                      </blockquote>
                    ),
                    hr: () => <hr className="my-6 border-white/10" />,
                  }}
                >
                  {content}
                </ReactMarkdown>
              </div>
            )}
          </div>
        </main>
      </div>
    </div>
  );
}
