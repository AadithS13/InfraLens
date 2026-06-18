"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const nav = [
  {
    href: "/",
    label: "Dashboard",
    icon: (
      <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
        <rect x="1.5" y="1.5" width="5.5" height="5.5" rx="1.2" stroke="currentColor" strokeWidth="1.3"/>
        <rect x="9" y="1.5" width="5.5" height="5.5" rx="1.2" stroke="currentColor" strokeWidth="1.3"/>
        <rect x="1.5" y="9" width="5.5" height="5.5" rx="1.2" stroke="currentColor" strokeWidth="1.3"/>
        <rect x="9" y="9" width="5.5" height="5.5" rx="1.2" stroke="currentColor" strokeWidth="1.3"/>
      </svg>
    ),
  },
  {
    href: "/search",
    label: "NL Search",
    icon: (
      <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
        <circle cx="7" cy="7" r="4.5" stroke="currentColor" strokeWidth="1.3"/>
        <path d="M10.5 10.5L14 14" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round"/>
        <path d="M5.5 7H8.5M7 5.5V8.5" stroke="currentColor" strokeWidth="1.1" strokeLinecap="round"/>
      </svg>
    ),
  },
  {
    href: "/projects",
    label: "Projects",
    icon: (
      <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
        <rect x="1.5" y="3.5" width="13" height="10" rx="1.5" stroke="currentColor" strokeWidth="1.3"/>
        <path d="M1.5 6.5H14.5" stroke="currentColor" strokeWidth="1.1"/>
        <path d="M5.5 6.5V13.5" stroke="currentColor" strokeWidth="1.1"/>
      </svg>
    ),
  },
];

export default function Sidebar() {
  const path = usePathname();

  return (
    <aside
      className="fixed top-0 left-0 h-screen w-60 flex flex-col z-40"
      style={{
        background: "rgba(255,255,255,0.8)",
        borderRight: "1px solid rgba(59,107,204,0.14)",
        backdropFilter: "blur(16px)",
        WebkitBackdropFilter: "blur(16px)",
      }}
    >
      {/* Logo */}
      <div className="px-5 pt-6 pb-5" style={{ borderBottom: "1px solid rgba(59,107,204,0.1)" }}>
        <div className="flex items-center gap-3">
          <div
            className="w-9 h-9 rounded-xl flex items-center justify-center shrink-0"
            style={{
              background: "linear-gradient(135deg, #1d4ed8 0%, #2563eb 100%)",
              boxShadow: "0 2px 8px rgba(37,99,235,0.3)",
            }}
          >
            {/* Blueprint / hard-hat icon */}
            <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
              <path d="M3 12C3 12 3.5 7.5 9 6.5C14.5 7.5 15 12 15 12H3Z" fill="rgba(255,255,255,0.25)" stroke="white" strokeWidth="1.3"/>
              <path d="M2 12H16" stroke="white" strokeWidth="1.3" strokeLinecap="round"/>
              <path d="M9 6.5V4.5" stroke="white" strokeWidth="1.3" strokeLinecap="round"/>
              <path d="M6 7.2V5.2" stroke="white" strokeWidth="1.2" strokeLinecap="round"/>
              <path d="M12 7.2V5.2" stroke="white" strokeWidth="1.2" strokeLinecap="round"/>
            </svg>
          </div>
          <div>
            <p className="text-sm font-bold text-[#0f1e3d] tracking-tight leading-none">InfraLens</p>
            <p className="text-[10px] mt-0.5 font-medium leading-none" style={{ color: "#d97706" }}>
              MahaRERA · AI
            </p>
          </div>
        </div>
      </div>

      {/* Nav */}
      <nav className="flex-1 px-3 py-4 space-y-0.5">
        {nav.map(({ href, label, icon }) => {
          const active = href === "/" ? path === "/" : path.startsWith(href);
          return (
            <Link
              key={href}
              href={href}
              className="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-150 group"
              style={
                active
                  ? {
                      background: "rgba(37,99,235,0.08)",
                      border: "1px solid rgba(37,99,235,0.16)",
                      color: "#1d4ed8",
                      fontWeight: 600,
                    }
                  : {
                      background: "transparent",
                      border: "1px solid transparent",
                      color: "#4b5e7f",
                    }
              }
            >
              <span
                className="transition-colors shrink-0"
                style={{ color: active ? "#2563eb" : "#9baabf" }}
              >
                {icon}
              </span>
              <span>{label}</span>
              {active && (
                <span
                  className="ml-auto w-1.5 h-1.5 rounded-full"
                  style={{ background: "#f59e0b" }}
                />
              )}
            </Link>
          );
        })}
      </nav>

      {/* AI status badge */}
      <div className="mx-3 mb-3">
        <div
          className="px-3 py-2.5 rounded-xl flex items-center gap-2.5"
          style={{
            background: "rgba(245,158,11,0.07)",
            border: "1px solid rgba(245,158,11,0.2)",
          }}
        >
          <div className="relative shrink-0">
            <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse-dot" />
          </div>
          <div>
            <p className="text-[11px] font-semibold text-[#92400e]">Claude Haiku</p>
            <p className="text-[10px] text-[#b45309]">NL search · live</p>
          </div>
        </div>
      </div>

      {/* Footer */}
      <div className="px-5 py-4" style={{ borderTop: "1px solid rgba(59,107,204,0.08)" }}>
        <p className="text-[10px] font-mono" style={{ color: "#9baabf" }}>
          v7.2 · 57,489 projects
        </p>
      </div>
    </aside>
  );
}
