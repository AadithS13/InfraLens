const badges = [
  {
    label: "57,489 Projects Indexed",
    color: "#1d4ed8",
    bg: "rgba(37,99,235,0.07)",
    border: "rgba(37,99,235,0.16)",
    dot: "#3b82f6",
  },
  {
    label: "7 APIs Reverse Engineered",
    color: "#b45309",
    bg: "rgba(245,158,11,0.07)",
    border: "rgba(245,158,11,0.18)",
    dot: "#f59e0b",
  },
  {
    label: "V7 · AI Search",
    color: "#6d28d9",
    bg: "rgba(109,40,217,0.07)",
    border: "rgba(109,40,217,0.16)",
    dot: "#a855f7",
  },
];

export default function StatsBar() {
  return (
    <div
      className="sticky top-0 z-20 -mx-8 px-8 py-2.5 mb-8 flex items-center gap-2 flex-wrap"
      style={{
        background: "rgba(238,242,251,0.82)",
        backdropFilter: "blur(12px)",
        WebkitBackdropFilter: "blur(12px)",
        borderBottom: "1px solid rgba(59,107,204,0.08)",
      }}
    >
      {badges.map(({ label, color, bg, border, dot }) => (
        <span
          key={label}
          className="inline-flex items-center gap-1.5 text-[11px] font-semibold px-3 py-1 rounded-full"
          style={{ color, background: bg, border: `1px solid ${border}` }}
        >
          <span className="w-1.5 h-1.5 rounded-full shrink-0" style={{ background: dot }} />
          {label}
        </span>
      ))}
    </div>
  );
}
