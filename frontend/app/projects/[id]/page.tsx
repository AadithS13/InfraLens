import { api, Change } from "@/lib/api";
import Link from "next/link";

const STATUS_COLORS: Record<string, { text: string; bg: string; border: string }> = {
  Ongoing:          { text: "#15803d", bg: "rgba(21,128,61,0.08)",   border: "rgba(21,128,61,0.2)" },
  New:              { text: "#1d4ed8", bg: "rgba(29,78,216,0.08)",   border: "rgba(29,78,216,0.2)" },
  Lapsed:           { text: "#b91c1c", bg: "rgba(185,28,28,0.08)",   border: "rgba(185,28,28,0.2)" },
  Completed:        { text: "#6d28d9", bg: "rgba(109,40,217,0.08)",  border: "rgba(109,40,217,0.2)" },
  "Under Approval": { text: "#b45309", bg: "rgba(180,83,9,0.08)",    border: "rgba(180,83,9,0.2)" },
};

const FIELD_LABELS: Record<string, string> = {
  project_status: "Status",
  project_current_status: "Current Status",
  proposed_completion_date: "Completion Date",
  project_name: "Project Name",
  total_units: "Total Units",
  total_sold_units: "Sold Units",
  rera_registration_no: "RERA Number",
};

function InfoRow({ label, value }: { label: string; value?: string | number | null }) {
  if (!value && value !== 0) return null;
  return (
    <div
      className="flex justify-between items-start py-2.5 gap-4"
      style={{ borderBottom: "1px solid rgba(59,107,204,0.06)" }}
    >
      <span className="text-xs text-[#9baabf] font-mono shrink-0">{label}</span>
      <span className="text-sm text-[#1e3a5f] font-medium text-right">{String(value)}</span>
    </div>
  );
}

function ChangeRow({ change, isLast }: { change: Change; isLast: boolean }) {
  return (
    <div
      className="flex gap-4 items-start py-4"
      style={{ borderBottom: isLast ? "none" : "1px solid rgba(59,107,204,0.06)" }}
    >
      <div className="flex flex-col items-center shrink-0 pt-1.5">
        <div
          className="w-2 h-2 rounded-full"
          style={{ background: "#22c55e", boxShadow: "0 0 6px rgba(34,197,94,0.5)" }}
        />
        {!isLast && (
          <div className="w-px flex-1 mt-1.5 min-h-6" style={{ background: "rgba(59,107,204,0.1)" }} />
        )}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2.5 mb-2 flex-wrap">
          <span
            className="text-[11px] font-bold px-2 py-0.5 rounded"
            style={{ background: "rgba(245,158,11,0.1)", color: "#92400e", border: "1px solid rgba(245,158,11,0.2)" }}
          >
            {FIELD_LABELS[change.field_name] ?? change.field_name}
          </span>
          <span className="text-[#9baabf] text-xs font-mono">
            {new Date(change.detected_at).toLocaleString()}
          </span>
        </div>
        <div className="flex items-center gap-2 flex-wrap">
          <span
            className="text-xs font-mono px-2.5 py-1 rounded-lg line-through"
            style={{ background: "rgba(185,28,28,0.07)", border: "1px solid rgba(185,28,28,0.18)", color: "#b91c1c" }}
          >
            {change.old_value || "—"}
          </span>
          <svg width="13" height="13" viewBox="0 0 13 13" fill="none" className="shrink-0">
            <path d="M2.5 6.5H10.5M7.5 3.5L10.5 6.5L7.5 9.5" stroke="#9baabf" strokeWidth="1.3" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
          <span
            className="text-xs font-mono px-2.5 py-1 rounded-lg"
            style={{ background: "rgba(21,128,61,0.08)", border: "1px solid rgba(21,128,61,0.2)", color: "#15803d" }}
          >
            {change.new_value || "—"}
          </span>
        </div>
      </div>
    </div>
  );
}

export default async function ProjectDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const [detailRes, changesRes] = await Promise.all([
    api.projects.get(Number(id)),
    api.projects.changes(Number(id)),
  ]);

  const p = detailRes.data;
  const changes = changesRes.data;
  const sc = STATUS_COLORS[p.project_status] ?? { text: "#64748b", bg: "rgba(100,116,139,0.08)", border: "rgba(100,116,139,0.2)" };

  return (
    <div className="max-w-4xl mx-auto space-y-5">

      {/* Breadcrumb */}
      <Link
        href="/projects"
        className="inline-flex items-center gap-1.5 text-xs font-semibold text-[#9baabf] hover:text-[#2563eb] transition-colors animate-fade-up"
      >
        <svg width="13" height="13" viewBox="0 0 13 13" fill="none">
          <path d="M8.5 2.5L4.5 6.5L8.5 10.5" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round"/>
        </svg>
        Back to Projects
      </Link>

      {/* ── Header card ───────────────────────────────────────────────── */}
      <div className="glass-card p-6 animate-fade-up delay-100">
        <div className="flex items-start justify-between gap-6 mb-4">
          <div className="min-w-0">
            <div className="flex items-center gap-2 mb-2 flex-wrap">
              <span
                className="text-[11px] font-bold px-2.5 py-0.5 rounded-full"
                style={{ color: sc.text, background: sc.bg, border: `1px solid ${sc.border}` }}
              >
                {p.project_status}
              </span>
              {p.project_type && (
                <span className="text-[11px] font-medium text-[#9baabf]">{p.project_type}</span>
              )}
            </div>
            <h1 className="text-2xl font-bold text-[#0f1e3d] tracking-tight">{p.project_name}</h1>
            <p className="text-[#4b5e7f] text-sm mt-1">{p.promoter_name}</p>
          </div>
          {p.total_units > 0 && (
            <div
              className="text-right shrink-0 px-5 py-4 rounded-xl"
              style={{ background: "rgba(37,99,235,0.06)", border: "1px solid rgba(37,99,235,0.12)" }}
            >
              <p className="text-3xl font-bold text-[#0f1e3d]">{p.total_units.toLocaleString()}</p>
              <p className="text-xs text-[#9baabf] mt-0.5">total units</p>
              {p.total_sold_units > 0 && (
                <p className="text-xs font-bold mt-1" style={{ color: "#15803d" }}>
                  {p.total_sold_units.toLocaleString()} sold
                </p>
              )}
            </div>
          )}
        </div>

        <div className="pt-4" style={{ borderTop: "1px solid rgba(59,107,204,0.08)" }}>
          <div className="flex items-center gap-3">
            <span className="text-[10px] font-bold text-[#9baabf] uppercase tracking-wider">RERA Reg.</span>
            <code
              className="text-xs font-mono px-2.5 py-1 rounded-lg"
              style={{ background: "rgba(245,158,11,0.08)", color: "#92400e", border: "1px solid rgba(245,158,11,0.2)" }}
            >
              {p.rera_registration_no}
            </code>
          </div>
        </div>
      </div>

      {/* ── Info grid ─────────────────────────────────────────────────── */}
      <div className="grid grid-cols-2 gap-4 animate-fade-up delay-200">
        <div className="glass-card p-5">
          <p className="text-[10px] font-bold text-[#9baabf] uppercase tracking-wider mb-3">Project Details</p>
          <InfoRow label="Type" value={p.project_type} />
          <InfoRow label="Current Status" value={p.project_current_status} />
          <InfoRow label="Registration Date" value={p.rera_registration_date ? new Date(p.rera_registration_date).toLocaleDateString() : null} />
          <InfoRow label="Completion Date" value={p.proposed_completion_date ? new Date(p.proposed_completion_date).toLocaleDateString() : null} />
          <InfoRow label="Total Units" value={p.total_units} />
          <InfoRow label="Sold Units" value={p.total_sold_units} />
          {p.promoter_type && <InfoRow label="Promoter Type" value={p.promoter_type} />}
        </div>

        <div className="glass-card p-5">
          <p className="text-[10px] font-bold text-[#9baabf] uppercase tracking-wider mb-3">Location</p>
          <InfoRow label="City" value={p.city} />
          <InfoRow label="District" value={p.district} />
          <InfoRow label="State" value={p.state} />
          <InfoRow label="Pincode" value={p.pincode} />
          {p.district && (
            <div className="mt-4 pt-4" style={{ borderTop: "1px solid rgba(59,107,204,0.06)" }}>
              <div
                className="h-20 rounded-xl flex items-center justify-center"
                style={{ background: "rgba(37,99,235,0.04)", border: "1px dashed rgba(37,99,235,0.15)" }}
              >
                <div className="text-center">
                  <p className="text-xl">📍</p>
                  <p className="text-xs text-[#9baabf] mt-1 font-mono">{p.district}, Maharashtra</p>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* ── Change history ─────────────────────────────────────────────── */}
      <div className="glass-card p-6 animate-fade-up delay-300">
        <div className="flex items-center justify-between mb-5">
          <h2 className="text-sm font-bold text-[#0f1e3d]">Change History</h2>
          <span
            className="text-[11px] font-bold px-2.5 py-0.5 rounded-full"
            style={
              changes.length > 0
                ? { background: "rgba(21,128,61,0.08)", color: "#15803d", border: "1px solid rgba(21,128,61,0.2)" }
                : { background: "rgba(59,107,204,0.06)", color: "#9baabf", border: "1px solid rgba(59,107,204,0.12)" }
            }
          >
            {changes.length} {changes.length === 1 ? "change" : "changes"}
          </span>
        </div>
        {changes.length === 0 ? (
          <div
            className="text-center py-8 rounded-xl"
            style={{ border: "1px dashed rgba(59,107,204,0.15)" }}
          >
            <p className="text-[#9baabf] text-sm">No changes recorded for this project.</p>
            <p className="text-[#c5d0e8] text-xs mt-1 font-mono">changes are detected on each crawl run</p>
          </div>
        ) : (
          <div>
            {changes.map((c, i) => (
              <ChangeRow key={i} change={c} isLast={i === changes.length - 1} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
