import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import Sidebar from "@/components/Sidebar";
import Background from "@/components/Background";
import StatsBar from "@/components/StatsBar";

const inter = Inter({ subsets: ["latin"], variable: "--font-inter" });

export const metadata: Metadata = {
  title: "InfraLens — MahaRERA Intelligence",
  description: "AI-powered construction intelligence platform for MahaRERA project data",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={inter.variable}>
      <body className="min-h-screen" style={{ background: "#eef2fb", color: "#0f1e3d" }}>
        <Background />
        <Sidebar />
        <main className="relative z-10 ml-60 min-h-screen">
          <StatsBar />
          <div className="p-8">{children}</div>
        </main>
      </body>
    </html>
  );
}
