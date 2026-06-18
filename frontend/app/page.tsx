import { api } from "@/lib/api";
import AnimatedCounter from "@/components/AnimatedCounter";

const STATUS_COLORS: Record<string, { text: string; bg: string; bar: string }> = {
  Ongoing:         { text: "#15803d", bg: "rgba(21,128,61,0.1)",   bar: "#22c55e" },
  New:             { text: "#1d4ed8", bg: "rgba(29,78,216,0.1)",   bar: "#3b82f6" },
  Lapsed:          { text: "#b91c1c", bg: "rgba(185,28,28,0.1)",   bar: "#ef4444" },
  Completed:       { text: "#6d28d9", bg: "rgba(109,40,217,0.1)",  bar: "#a855f7" },
  "Under Approval":{ text: "#b45309", bg: "rgba(180,83,9,0.1)",    bar: "#f59e0b" },
};

export default async function DashboardPage() {
  const [statusRes, buildersRes, districtRes, crawlsRes, totalRes] = await Promise.all([
    api.analytics.statusDist(),
    api.analytics.topBuilders(8),
    api.analytics.byDistrict(8),
    api.crawls.list(),
    api.projects.list({ limit: 1 }),
  ]);

  const statuses = statusRes.data;
  const builders = buildersRes.data;
  const districts = districtRes.data;
  const crawls = crawlsRes.data.slice(0, 5);
  const totalProjects = totalRes.meta.total;
  const maxCount = Math.max(...statuses.map((s) => s.count), 1);

  return (
    <div className="max-w-6xl mx-auto space-y-8">

      {/* ── Header ────────────────────────────────────────────────────── */}
      <div className="animate-fade-up">
        <div className="flex items-center gap-2 mb-2">
          <span
            className="text-[10px] font-semibold px-2.5 py-0.5 rounded-full uppercase tracking-widest"
            style={{ background: "rgba(37,99,235,0.1)", color: "#1d4ed8", border: "1px solid rgba(37,99,235,0.2)" }}
          >
            Live Data
          </span>
          <span className="text-xs text-[#9baabf] font-mono">Maharashtra Real Estate Regulatory Authority</span>
        </div>
        <h1 className="text-3xl font-bold text-[#0f1e3d] tracking-tight">Intelligence Dashboard</h1>
        <p className="text-[#4b5e7f] text-sm mt-1">
          Real-time construction data across {totalProjects.toLocaleString()} registered projects
        </p>
      </div>

      {/* ── Stat cards ────────────────────────────────────────────────── */}
      <div className="grid grid-cols-3 gap-5">
        {[
          {
            label: "Total Projects",
            value: totalProjects,
            sub: "registered with MahaRERA",
            accent: "#2563eb",
            accentBg: "rgba(37,99,235,0.08)",
            icon: (
              <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
                <path d="M3 17V8.5L10 3.5L17 8.5V17H12.5V12.5H7.5V17H3Z" stroke="#2563eb" strokeWidth="1.4" fill="rgba(37,99,235,0.12)"/>
              </svg>
            ),
            delay: "",
          },
          {
            label: "Top Builders",
            value: builders.length,
            sub: "by project volume",
            accent: "#d97706",
            accentBg: "rgba(217,119,6,0.08)",
            icon: (
              <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
                <circle cx="10" cy="7" r="3.5" stroke="#d97706" strokeWidth="1.4"/>
                <path d="M3 17c0-3.314 3.134-6 7-6s7 2.686 7 6" stroke="#d97706" strokeWidth="1.4" strokeLinecap="round"/>
              </svg>
            ),
            delay: "delay-100",
          },
          {
            label: "Districts",
            value: districts.length,
            sub: "across Maharashtra",
            accent: "#7c3aed",
            accentBg: "rgba(124,58,237,0.08)",
            icon: (
              <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
                <circle cx="10" cy="9" r="4" stroke="#7c3aed" strokeWidth="1.4"/>
                <circle cx="10" cy="9" r="1.5" fill="#7c3aed"/>
                <path d="M10 13c-4 2-5 4-5 6h10c0-2-1-4-5-6Z" stroke="#7c3aed" strokeWidth="1.3" fill="rgba(124,58,237,0.1)"/>
              </svg>
            ),
            delay: "delay-200",
          },
        ].map(({ label, value, sub, accent, accentBg, icon, delay }) => (
          <div key={label} className={`glass-card animate-fade-up ${delay} p-6 relative overflow-hidden`}>
            <div
              className="absolute -top-8 -right-8 w-28 h-28 rounded-full opacity-60"
              style={{ background: `radial-gradient(circle, ${accentBg} 0%, transparent 70%)` }}
            />
            <div
              className="w-10 h-10 rounded-xl flex items-center justify-center mb-4"
              style={{ background: accentBg, border: `1px solid ${accent}22` }}
            >
              {icon}
            </div>
            <p className="text-4xl font-bold mb-1" style={{ color: "#0f1e3d" }}>
              <AnimatedCounter value={value} />
            </p>
            <p className="text-xs font-bold uppercase tracking-wider" style={{ color: "#4b5e7f" }}>{label}</p>
            <p className="text-xs mt-0.5" style={{ color: "#9baabf" }}>{sub}</p>
          </div>
        ))}
      </div>

      {/* ── Status + Districts ─────────────────────────────────────────── */}
      <div className="grid grid-cols-2 gap-6 animate-fade-up delay-200">

        {/* Status distribution */}
        <div className="glass-card p-6">
          <div className="flex items-center justify-between mb-5">
            <h2 className="text-sm font-bold text-[#0f1e3d]">Status Distribution</h2>
            <span className="text-[10px] font-mono text-[#9baabf] uppercase tracking-wider">
              {statuses.reduce((a, s) => a + s.count, 0).toLocaleString()} total
            </span>
          </div>
          <div className="space-y-4">
            {statuses.map((s) => {
              const c = STATUS_COLORS[s.status] ?? { text: "#64748b", bg: "rgba(100,116,139,0.1)", bar: "#94a3b8" };
              return (
                <div key={s.status}>
                  <div className="flex justify-between items-center mb-1.5">
                    <div className="flex items-center gap-2">
                      <span className="w-2 h-2 rounded-full" style={{ background: c.bar }} />
                      <span className="text-xs font-semibold" style={{ color: c.text }}>{s.status}</span>
                    </div>
                    <span className="text-xs font-mono text-[#9baabf]">{s.count.toLocaleString()}</span>
                  </div>
                  <div className="h-2 rounded-full" style={{ background: "rgba(59,107,204,0.08)" }}>
                    <div
                      className="h-full rounded-full"
                      style={{
                        width: `${(s.count / maxCount) * 100}%`,
                        background: `linear-gradient(90deg, ${c.bar}, ${c.bar}90)`,
                        boxShadow: `0 0 6px ${c.bar}50`,
                      }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Top districts */}
        <div className="glass-card p-6">
          <div className="flex items-center justify-between mb-5">
            <h2 className="text-sm font-bold text-[#0f1e3d]">Top Districts</h2>
            <span className="text-[10px] font-mono text-[#9baabf] uppercase tracking-wider">by count</span>
          </div>
          <div className="space-y-2.5">
            {districts.map((d, i) => (
              <div key={d.district} className="flex items-center gap-3">
                <span
                  className="w-5 h-5 rounded text-[10px] font-bold flex items-center justify-center shrink-0"
                  style={
                    i < 3
                      ? { background: "rgba(245,158,11,0.12)", color: "#b45309", border: "1px solid rgba(245,158,11,0.25)" }
                      : { background: "rgba(59,107,204,0.06)", color: "#6b82a8", border: "1px solid rgba(59,107,204,0.12)" }
                  }
                >
                  {i + 1}
                </span>
                <span className="flex-1 text-sm text-[#1e3a5f] font-medium truncate">{d.district || "—"}</span>
                <div className="flex items-center gap-2">
                  <div
                    className="h-1.5 rounded-full"
                    style={{
                      width: `${(d.count / districts[0].count) * 64}px`,
                      background: i < 3
                        ? "linear-gradient(90deg,#f59e0b,#fbbf24)"
                        : "rgba(59,107,204,0.2)",
                    }}
                  />
                  <span className="text-xs font-mono text-[#9baabf] w-12 text-right">{d.count.toLocaleString()}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* ── Top builders ──────────────────────────────────────────────── */}
      <div className="glass-card animate-fade-up delay-300">
        <div className="px-6 py-4" style={{ borderBottom: "1px solid rgba(59,107,204,0.08)" }}>
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-bold text-[#0f1e3d]">Top Builders by Project Count</h2>
            <span className="text-[10px] font-mono text-[#9baabf] uppercase tracking-wider">active developers</span>
          </div>
        </div>
        <table className="w-full text-sm">
          <thead>
            <tr style={{ borderBottom: "1px solid rgba(59,107,204,0.06)" }}>
              <th className="text-left px-6 py-3 text-[10px] font-semibold text-[#9baabf] uppercase tracking-wider">#</th>
              <th className="text-left px-4 py-3 text-[10px] font-semibold text-[#9baabf] uppercase tracking-wider">Builder</th>
              <th className="text-right px-4 py-3 text-[10px] font-semibold text-[#9baabf] uppercase tracking-wider">Projects</th>
              <th className="text-right px-6 py-3 text-[10px] font-semibold text-[#9baabf] uppercase tracking-wider">Units</th>
            </tr>
          </thead>
          <tbody>
            {builders.map((b, i) => (
              <tr
                key={b.promoter_name}
                className="row-hover transition-colors"
                style={{
                  borderBottom: i < builders.length - 1 ? "1px solid rgba(59,107,204,0.05)" : "none",
                }}
              >
                <td className="px-6 py-3.5">
                  <span
                    className="w-6 h-6 rounded text-[11px] font-bold flex items-center justify-center"
                    style={
                      i < 3
                        ? { background: "rgba(245,158,11,0.12)", color: "#b45309" }
                        : { background: "rgba(59,107,204,0.06)", color: "#9baabf" }
                    }
                  >
                    {i + 1}
                  </span>
                </td>
                <td className="px-4 py-3.5 font-medium text-[#1e3a5f]">{b.promoter_name}</td>
                <td className="px-4 py-3.5 text-right font-mono font-bold text-[#15803d]">{b.project_count}</td>
                <td className="px-6 py-3.5 text-right font-mono text-xs text-[#9baabf]">{b.total_units.toLocaleString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* ── Crawl runs ────────────────────────────────────────────────── */}
      {crawls.length > 0 && (
        <div className="animate-fade-up delay-400">
          <h2 className="text-xs font-bold text-[#9baabf] uppercase tracking-wider mb-3">Data Pipeline · Recent Runs</h2>
          <div className="space-y-2">
            {crawls.map((c) => {
              const ok = c.status === "completed";
              const warn = c.status.startsWith("completed_with");
              const [color, bg, border] = ok
                ? ["#15803d", "rgba(21,128,61,0.08)", "rgba(21,128,61,0.2)"]
                : warn
                ? ["#b45309", "rgba(180,83,9,0.08)", "rgba(180,83,9,0.2)"]
                : ["#b91c1c", "rgba(185,28,28,0.08)", "rgba(185,28,28,0.2)"];
              return (
                <div key={c.id} className="glass-card px-5 py-3.5 flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-2 h-2 rounded-full" style={{ background: color }} />
                    <span
                      className="text-[11px] font-semibold px-2.5 py-0.5 rounded-full"
                      style={{ color, background: bg, border: `1px solid ${border}` }}
                    >
                      {c.status}
                    </span>
                    <span className="text-xs font-mono text-[#9baabf]">
                      {new Date(c.started_at).toLocaleString()}
                    </span>
                  </div>
                  <div className="flex items-center gap-5 text-xs font-mono">
                    <span className="text-[#4b5e7f]">
                      <span className="font-semibold text-[#0f1e3d]">{c.processed.toLocaleString()}</span> processed
                    </span>
                    {c.failed > 0 && (
                      <span className="text-[#b91c1c] font-semibold">{c.failed} failed</span>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}
