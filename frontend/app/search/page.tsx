"use client";

import { useState, useRef } from "react";
import Link from "next/link";
import { api, NLResult, Project, Builder } from "@/lib/api";

const STATUS_COLORS: Record<string, { text: string; bg: string; border: string }> = {
  Ongoing:          { text: "#15803d", bg: "rgba(21,128,61,0.08)",   border: "rgba(21,128,61,0.2)" },
  New:              { text: "#1d4ed8", bg: "rgba(29,78,216,0.08)",   border: "rgba(29,78,216,0.2)" },
  Lapsed:           { text: "#b91c1c", bg: "rgba(185,28,28,0.08)",   border: "rgba(185,28,28,0.2)" },
  Completed:        { text: "#6d28d9", bg: "rgba(109,40,217,0.08)",  border: "rgba(109,40,217,0.2)" },
  "Under Approval": { text: "#b45309", bg: "rgba(180,83,9,0.08)",    border: "rgba(180,83,9,0.2)" },
};

const EXAMPLE_QUERIES = [
  "show ongoing residential projects in Thane",
  "plotted development projects that missed their deadline",
  "show stalled projects in Pune",
  "which developers have registered more than 5 projects",
  "lapsed commercial projects in Nagpur",
];

export default function SearchPage() {
  const [query, setQuery] = useState("");
  const [result, setResult] = useState<NLResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  async function runSearch(q: string) {
    if (!q.trim()) return;
    setLoading(true);
    setError(null);
    setResult(null);
    try {
      const res = await api.search.nl(q.trim());
      if (res.query_type === "projects") {
        res.data = Array.from(new Map((res.data as Project[]).map((p) => [p.id, p])).values());
      }
      setResult(res);
    } catch {
      setError("Could not reach the InfraLens API. Is the server running on :8080?");
    } finally {
      setLoading(false);
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter") runSearch(query);
  }

  function pickExample(q: string) {
    setQuery(q);
    runSearch(q);
    inputRef.current?.focus();
  }

  return (
    <div className="max-w-4xl mx-auto">

      {/* ── Hero ─────────────────────────────────────────────────────── */}
      <div className="text-center mb-10 animate-fade-up">
        <div
          className="inline-flex items-center gap-2 text-xs font-semibold px-3 py-1.5 rounded-full mb-4"
          style={{ background: "rgba(37,99,235,0.07)", border: "1px solid rgba(37,99,235,0.18)", color: "#1d4ed8" }}
        >
          <span className="w-1.5 h-1.5 rounded-full animate-pulse-dot" style={{ background: "#3b82f6" }} />
          AI-powered search
        </div>
        <h1 className="text-4xl font-bold text-[#0f1e3d] mb-3 tracking-tight">
          Ask anything about<br />
          <span style={{ color: "#2563eb" }}>Maharashtra&apos;s real estate</span>
        </h1>
        <p className="text-[#4b5e7f] text-sm">
          Natural language queries across 57,489 MahaRERA registered projects
        </p>
      </div>

      {/* ── Search bar ───────────────────────────────────────────────── */}
      <div className="animate-fade-up delay-100 mb-4">
        <div
          className="relative flex items-center rounded-2xl overflow-hidden"
          style={{
            background: "rgba(255,255,255,0.9)",
            border: "1.5px solid rgba(37,99,235,0.18)",
            boxShadow: "0 4px 24px rgba(15,30,61,0.08), 0 0 0 4px rgba(37,99,235,0.04)",
          }}
        >
          <div className="pl-5 pr-3 shrink-0">
            <svg width="17" height="17" viewBox="0 0 17 17" fill="none">
              <circle cx="7.5" cy="7.5" r="5" stroke="#9baabf" strokeWidth="1.5"/>
              <path d="M11.5 11.5L15 15" stroke="#9baabf" strokeWidth="1.5" strokeLinecap="round"/>
            </svg>
          </div>
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="show ongoing residential projects in Thane..."
            className="flex-1 py-4 pr-4 text-[15px] text-[#0f1e3d] placeholder-[#c5d0e8] focus:outline-none bg-transparent"
          />
          <button
            onClick={() => runSearch(query)}
            disabled={loading || !query.trim()}
            className="mr-2 px-5 py-2.5 rounded-xl text-sm font-bold transition-all duration-150 disabled:opacity-40 disabled:cursor-not-allowed"
            style={{
              background: loading || !query.trim() ? "rgba(37,99,235,0.08)" : "#2563eb",
              color: loading || !query.trim() ? "#2563eb" : "#fff",
              boxShadow: loading || !query.trim() ? "none" : "0 2px 8px rgba(37,99,235,0.3)",
            }}
          >
            {loading ? (
              <span className="flex items-center gap-2">
                <span className="w-3.5 h-3.5 border-2 border-current border-t-transparent rounded-full animate-spin" />
                Thinking…
              </span>
            ) : (
              "Search"
            )}
          </button>
        </div>
      </div>

      {/* ── Example chips ────────────────────────────────────────────── */}
      {!result && !loading && (
        <div className="flex flex-wrap gap-2 mb-10 animate-fade-up delay-200">
          {EXAMPLE_QUERIES.map((q) => (
            <button
              key={q}
              onClick={() => pickExample(q)}
              className="text-xs font-medium px-3 py-1.5 rounded-lg transition-all duration-150"
              style={{
                background: "rgba(255,255,255,0.7)",
                border: "1px solid rgba(59,107,204,0.16)",
                color: "#4b5e7f",
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLButtonElement).style.borderColor = "rgba(37,99,235,0.35)";
                (e.currentTarget as HTMLButtonElement).style.color = "#1d4ed8";
                (e.currentTarget as HTMLButtonElement).style.background = "rgba(37,99,235,0.06)";
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLButtonElement).style.borderColor = "rgba(59,107,204,0.16)";
                (e.currentTarget as HTMLButtonElement).style.color = "#4b5e7f";
                (e.currentTarget as HTMLButtonElement).style.background = "rgba(255,255,255,0.7)";
              }}
            >
              {q}
            </button>
          ))}
        </div>
      )}

      {/* ── Loading ──────────────────────────────────────────────────── */}
      {loading && (
        <div className="flex flex-col items-center gap-4 py-16 text-center">
          <div
            className="w-12 h-12 rounded-2xl flex items-center justify-center"
            style={{ background: "rgba(37,99,235,0.08)", border: "1px solid rgba(37,99,235,0.18)" }}
          >
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none" className="animate-spin">
              <circle cx="10" cy="10" r="8" stroke="rgba(37,99,235,0.15)" strokeWidth="2"/>
              <path d="M10 2a8 8 0 0 1 8 8" stroke="#2563eb" strokeWidth="2" strokeLinecap="round"/>
            </svg>
          </div>
          <div>
            <p className="text-[#1e3a5f] text-sm font-semibold">Natural language → structured project data</p>
            <p className="text-[#9baabf] text-xs mt-1 font-mono">extracting filters · searching 57k+ projects</p>
          </div>
        </div>
      )}

      {/* ── Error ────────────────────────────────────────────────────── */}
      {error && (
        <div
          className="rounded-xl px-5 py-4 flex items-start gap-3"
          style={{ background: "rgba(185,28,28,0.06)", border: "1px solid rgba(185,28,28,0.2)" }}
        >
          <span className="text-lg shrink-0" style={{ color: "#b91c1c" }}>⚠</span>
          <p className="text-sm" style={{ color: "#991b1b" }}>{error}</p>
        </div>
      )}

      {/* ── Results ──────────────────────────────────────────────────── */}
      {result && (
        <div className="animate-fade-up space-y-5">

          {/* Interpreted filters */}
          <div
            className="rounded-xl px-5 py-4"
            style={{ background: "rgba(255,255,255,0.8)", border: "1px solid rgba(59,107,204,0.14)" }}
          >
            <div className="flex items-start gap-3 flex-wrap">
              <div className="flex items-center gap-1.5 shrink-0 pt-0.5">
                <svg width="13" height="13" viewBox="0 0 13 13" fill="none">
                  <circle cx="6.5" cy="6.5" r="5.5" stroke="#2563eb" strokeWidth="1.3"/>
                  <path d="M4.5 6.5L6 8L8.5 5" stroke="#2563eb" strokeWidth="1.3" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
                <span className="text-[11px] font-semibold text-[#9baabf]">interpreted as</span>
              </div>
              {Object.entries(result.interpreted).length === 0 ? (
                <span className="text-[#9baabf] text-xs font-mono">no filters extracted</span>
              ) : (
                Object.entries(result.interpreted).map(([k, v]) => (
                  <span
                    key={k}
                    className="text-[11px] font-mono px-2.5 py-1 rounded-lg"
                    style={{ background: "rgba(37,99,235,0.07)", border: "1px solid rgba(37,99,235,0.15)" }}
                  >
                    <span className="text-[#9baabf]">{k}=</span>
                    <span className="font-semibold text-[#1d4ed8]">{String(v)}</span>
                  </span>
                ))
              )}
              <span
                className="ml-auto text-[10px] font-semibold px-2.5 py-1 rounded-full shrink-0"
                style={{ background: "rgba(37,99,235,0.07)", color: "#1d4ed8", border: "1px solid rgba(37,99,235,0.15)" }}
              >
                AI Search
              </span>
            </div>
          </div>

          {/* Builder results */}
          {result.query_type === "builders" && (
            <div>
              <p className="text-[10px] font-bold text-[#9baabf] uppercase tracking-wider mb-3">
                Builders · {(result.data as Builder[]).length} results
              </p>
              <div className="glass-card overflow-hidden">
                <table className="w-full text-sm">
                  <thead>
                    <tr style={{ borderBottom: "1px solid rgba(59,107,204,0.08)" }}>
                      <th className="text-left px-5 py-3 text-[10px] font-semibold text-[#9baabf] uppercase tracking-wider">Builder</th>
                      <th className="text-right px-5 py-3 text-[10px] font-semibold text-[#9baabf] uppercase tracking-wider">Projects</th>
                      <th className="text-right px-5 py-3 text-[10px] font-semibold text-[#9baabf] uppercase tracking-wider">Units</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(result.data as Builder[]).map((b, i) => (
                      <tr
                        key={b.promoter_name}
                        style={{ borderBottom: i < (result.data as Builder[]).length - 1 ? "1px solid rgba(59,107,204,0.06)" : "none" }}
                      >
                        <td className="px-5 py-3.5 font-medium text-[#1e3a5f]">{b.promoter_name}</td>
                        <td className="px-5 py-3.5 text-right font-mono font-bold text-[#15803d]">{b.project_count}</td>
                        <td className="px-5 py-3.5 text-right font-mono text-xs text-[#9baabf]">{b.total_units.toLocaleString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Project results */}
          {result.query_type === "projects" && (
            <div>
              <p className="text-[10px] font-bold text-[#9baabf] uppercase tracking-wider mb-3">
                Projects · {result.meta?.total.toLocaleString() ?? (result.data as Project[]).length} results
              </p>
              {(result.data as Project[]).length === 0 ? (
                <div
                  className="rounded-xl px-8 py-12 text-center"
                  style={{ border: "1px dashed rgba(59,107,204,0.2)" }}
                >
                  <p className="text-[#9baabf] text-sm">No projects match this query.</p>
                </div>
              ) : (
                <div className="space-y-2">
                  {(result.data as Project[]).map((p) => {
                    const c = STATUS_COLORS[p.project_status] ?? { text: "#64748b", bg: "rgba(100,116,139,0.08)", border: "rgba(100,116,139,0.2)" };
                    return (
                      <Link
                        key={p.id}
                        href={`/projects/${p.id}`}
                        className="glass-card block px-5 py-4 transition-all duration-150"
                      >
                        <div className="flex items-start justify-between gap-4">
                          <div className="min-w-0">
                            <h3 className="font-semibold text-sm text-[#0f1e3d] truncate">{p.project_name}</h3>
                            <p className="text-xs text-[#9baabf] mt-0.5 truncate">{p.promoter_name}</p>
                          </div>
                          <span
                            className="text-[11px] font-semibold px-2.5 py-0.5 rounded-full shrink-0"
                            style={{ color: c.text, background: c.bg, border: `1px solid ${c.border}` }}
                          >
                            {p.project_status}
                          </span>
                        </div>
                        <div className="flex gap-3 mt-2">
                          {p.district && <span className="text-[11px] text-[#9baabf]">📍 {p.district}</span>}
                          {p.project_type && <span className="text-[11px] text-[#c5d0e8]">{p.project_type}</span>}
                          {p.rera_registration_no && (
                            <span className="text-[11px] font-mono text-[#c5d0e8] ml-auto">{p.rera_registration_no}</span>
                          )}
                        </div>
                      </Link>
                    );
                  })}
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {/* ── Empty state ──────────────────────────────────────────────── */}
      {!result && !loading && !error && (
        <div
          className="rounded-2xl px-8 py-16 text-center animate-fade-up delay-300"
          style={{ border: "1px dashed rgba(59,107,204,0.2)" }}
        >
          <div
            className="w-12 h-12 rounded-2xl flex items-center justify-center mx-auto mb-4"
            style={{ background: "rgba(37,99,235,0.07)", border: "1px solid rgba(37,99,235,0.14)" }}
          >
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
              <circle cx="9" cy="9" r="6" stroke="#2563eb" strokeWidth="1.4"/>
              <path d="M14 14L18 18" stroke="#2563eb" strokeWidth="1.4" strokeLinecap="round"/>
              <path d="M7 9H11M9 7V11" stroke="#2563eb" strokeWidth="1.2" strokeLinecap="round"/>
            </svg>
          </div>
          <p className="text-[#4b5e7f] text-sm font-semibold">Type a question to search</p>
          <p className="text-[#9baabf] text-xs mt-1 font-mono">natural language → structured project data</p>
        </div>
      )}
    </div>
  );
}
