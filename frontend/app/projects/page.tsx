"use client";

import { useState, useCallback } from "react";
import Link from "next/link";
import { api, Project, Meta } from "@/lib/api";

const STATUS_OPTIONS = ["", "Ongoing", "New", "Lapsed", "Completed", "Under Approval"];
const TYPE_OPTIONS = ["", "Residential", "Commercial", "Plotted Development", "Retail", "Mixed Development"];

const STATUS_COLORS: Record<string, { text: string; bg: string; border: string }> = {
  Ongoing:          { text: "#15803d", bg: "rgba(21,128,61,0.08)",   border: "rgba(21,128,61,0.2)" },
  New:              { text: "#1d4ed8", bg: "rgba(29,78,216,0.08)",   border: "rgba(29,78,216,0.2)" },
  Lapsed:           { text: "#b91c1c", bg: "rgba(185,28,28,0.08)",   border: "rgba(185,28,28,0.2)" },
  Completed:        { text: "#6d28d9", bg: "rgba(109,40,217,0.08)",  border: "rgba(109,40,217,0.2)" },
  "Under Approval": { text: "#b45309", bg: "rgba(180,83,9,0.08)",    border: "rgba(180,83,9,0.2)" },
};

const inputStyle: React.CSSProperties = {
  background: "rgba(255,255,255,0.75)",
  border: "1px solid rgba(59,107,204,0.16)",
  color: "#0f1e3d",
  borderRadius: "10px",
  padding: "9px 14px",
  fontSize: "13px",
  outline: "none",
  transition: "border-color .15s",
};

export default function ProjectsPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [meta, setMeta] = useState<Meta | null>(null);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);

  const [q, setQ] = useState("");
  const [status, setStatus] = useState("");
  const [type, setType] = useState("");
  const [district, setDistrict] = useState("");
  const [page, setPage] = useState(1);

  const search = useCallback(
    async (p = 1) => {
      setLoading(true);
      try {
        const params: Record<string, string | number> = { limit: 20, page: p };
        if (q) params.q = q;
        if (status) params.status = status;
        if (type) params.type = type;
        if (district) params.district = district;
        const res = await api.projects.list(params);
        const unique = Array.from(new Map(res.data.map((item) => [item.id, item])).values());
        setProjects(unique);
        setMeta(res.meta);
        setPage(p);
        setSearched(true);
      } finally {
        setLoading(false);
      }
    },
    [q, status, type, district]
  );

  function handleKey(e: React.KeyboardEvent) {
    if (e.key === "Enter") search(1);
  }

  const totalPages = meta ? Math.ceil(meta.total / meta.limit) : 0;

  return (
    <div className="max-w-6xl mx-auto space-y-6">

      {/* ── Header ───────────────────────────────────────────────────── */}
      <div className="animate-fade-up">
        <h1 className="text-3xl font-bold text-[#0f1e3d] tracking-tight">Projects</h1>
        <p className="text-[#4b5e7f] text-sm mt-1">Search and filter 57,489 MahaRERA registered projects</p>
      </div>

      {/* ── Filter bar ───────────────────────────────────────────────── */}
      <div className="glass-card p-5 animate-fade-up delay-100">
        <div className="flex gap-3 flex-wrap items-end">
          <div className="flex-1 min-w-48">
            <label className="text-[10px] font-bold text-[#9baabf] uppercase tracking-wider block mb-1.5">Name / Promoter</label>
            <input
              value={q}
              onChange={(e) => setQ(e.target.value)}
              onKeyDown={handleKey}
              placeholder="Search projects..."
              style={inputStyle}
              className="w-full"
            />
          </div>
          <div>
            <label className="text-[10px] font-bold text-[#9baabf] uppercase tracking-wider block mb-1.5">Status</label>
            <select
              value={status}
              onChange={(e) => setStatus(e.target.value)}
              style={inputStyle}
              className="cursor-pointer appearance-none pr-8"
            >
              {STATUS_OPTIONS.map((s) => (
                <option key={s} value={s}>{s || "All statuses"}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="text-[10px] font-bold text-[#9baabf] uppercase tracking-wider block mb-1.5">Type</label>
            <select
              value={type}
              onChange={(e) => setType(e.target.value)}
              style={inputStyle}
              className="cursor-pointer appearance-none pr-8"
            >
              {TYPE_OPTIONS.map((t) => (
                <option key={t} value={t}>{t || "All types"}</option>
              ))}
            </select>
          </div>
          <div className="w-36">
            <label className="text-[10px] font-bold text-[#9baabf] uppercase tracking-wider block mb-1.5">District</label>
            <input
              value={district}
              onChange={(e) => setDistrict(e.target.value)}
              onKeyDown={handleKey}
              placeholder="e.g. Pune"
              style={inputStyle}
              className="w-full"
            />
          </div>
          <button
            onClick={() => search(1)}
            disabled={loading}
            className="px-6 py-2.5 rounded-xl text-sm font-bold transition-all duration-150 disabled:opacity-50"
            style={{
              background: loading ? "rgba(37,99,235,0.08)" : "#2563eb",
              color: loading ? "#2563eb" : "#fff",
              boxShadow: loading ? "none" : "0 2px 8px rgba(37,99,235,0.25)",
            }}
          >
            {loading ? (
              <span className="flex items-center gap-2">
                <span className="w-3.5 h-3.5 border-2 border-current border-t-transparent rounded-full animate-spin" />
                Searching
              </span>
            ) : "Search"}
          </button>
        </div>
      </div>

      {/* ── Empty state ───────────────────────────────────────────────── */}
      {!searched && !loading && (
        <div
          className="rounded-2xl px-8 py-16 text-center animate-fade-up delay-200"
          style={{ border: "1px dashed rgba(59,107,204,0.2)" }}
        >
          <p className="text-[#9baabf] text-sm">Enter filters above and click Search</p>
        </div>
      )}

      {/* ── Results ───────────────────────────────────────────────────── */}
      {searched && !loading && (
        <div className="animate-fade-up">
          {meta && (
            <p className="text-xs font-mono text-[#9baabf] mb-3">
              <span className="font-bold text-[#0f1e3d]">{meta.total.toLocaleString()}</span> projects found
              {" · "}page {meta.page} of {totalPages}
            </p>
          )}

          <div className="glass-card overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr style={{ borderBottom: "1px solid rgba(59,107,204,0.08)" }}>
                  <th className="text-left px-5 py-3.5 text-[10px] font-bold text-[#9baabf] uppercase tracking-wider">Project</th>
                  <th className="text-left px-4 py-3.5 text-[10px] font-bold text-[#9baabf] uppercase tracking-wider">Promoter</th>
                  <th className="text-left px-4 py-3.5 text-[10px] font-bold text-[#9baabf] uppercase tracking-wider">District</th>
                  <th className="text-left px-4 py-3.5 text-[10px] font-bold text-[#9baabf] uppercase tracking-wider">Type</th>
                  <th className="text-left px-5 py-3.5 text-[10px] font-bold text-[#9baabf] uppercase tracking-wider">Status</th>
                </tr>
              </thead>
              <tbody>
                {projects.map((p, i) => {
                  const c = STATUS_COLORS[p.project_status] ?? { text: "#64748b", bg: "rgba(100,116,139,0.08)", border: "rgba(100,116,139,0.2)" };
                  return (
                    <tr
                      key={p.id}
                      style={{ borderBottom: i < projects.length - 1 ? "1px solid rgba(59,107,204,0.05)" : "none" }}
                      onMouseEnter={(e) => (e.currentTarget.style.background = "rgba(37,99,235,0.025)")}
                      onMouseLeave={(e) => (e.currentTarget.style.background = "")}
                    >
                      <td className="px-5 py-3.5">
                        <Link
                          href={`/projects/${p.id}`}
                          className="font-semibold text-[#0f1e3d] hover:text-[#2563eb] transition-colors block truncate max-w-[240px]"
                        >
                          {p.project_name}
                        </Link>
                        <p className="text-[10px] font-mono text-[#c5d0e8] mt-0.5 truncate">{p.rera_registration_no}</p>
                      </td>
                      <td className="px-4 py-3.5 text-xs text-[#4b5e7f] truncate max-w-[180px]">{p.promoter_name}</td>
                      <td className="px-4 py-3.5 text-xs text-[#4b5e7f]">{p.district || "—"}</td>
                      <td className="px-4 py-3.5 text-xs text-[#9baabf]">{p.project_type}</td>
                      <td className="px-5 py-3.5">
                        <span
                          className="text-[11px] font-semibold px-2.5 py-0.5 rounded-full"
                          style={{ color: c.text, background: c.bg, border: `1px solid ${c.border}` }}
                        >
                          {p.project_status}
                        </span>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          {meta && meta.total > meta.limit && (
            <div className="flex items-center justify-between mt-4">
              <button
                onClick={() => search(page - 1)}
                disabled={page <= 1}
                className="text-xs font-semibold text-[#4b5e7f] hover:text-[#2563eb] disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
              >
                ← Previous
              </button>
              <div className="flex items-center gap-1">
                {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                  const pg = Math.max(1, Math.min(page - 2, totalPages - 4)) + i;
                  return (
                    <button
                      key={pg}
                      onClick={() => search(pg)}
                      className="w-7 h-7 rounded-lg text-xs font-semibold transition-all"
                      style={
                        pg === page
                          ? { background: "#2563eb", color: "#fff", boxShadow: "0 1px 4px rgba(37,99,235,0.3)" }
                          : { background: "rgba(255,255,255,0.5)", color: "#4b5e7f", border: "1px solid rgba(59,107,204,0.14)" }
                      }
                    >
                      {pg}
                    </button>
                  );
                })}
              </div>
              <button
                onClick={() => search(page + 1)}
                disabled={page >= totalPages}
                className="text-xs font-semibold text-[#4b5e7f] hover:text-[#2563eb] disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
              >
                Next →
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
