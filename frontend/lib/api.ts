// Browser: relative (proxy rewrites kick in). Server components: direct to Go API.
const BASE =
  typeof window === "undefined"
    ? "http://localhost:8080/api/v1"
    : "/api/v1";

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`, { cache: "no-store" });
  if (!res.ok) throw new Error(`${path} → ${res.status}`);
  return res.json();
}

// ── Types ────────────────────────────────────────────────────────────────────

export type Project = {
  id: number;
  maha_id: number;
  rera_registration_no: string;
  project_name: string;
  project_type: string;
  project_status: string;
  project_current_status: string;
  rera_registration_date: string;
  proposed_completion_date: string;
  total_units: number;
  total_sold_units: number;
  promoter_name: string;
  city: string;
  district: string;
  state: string;
  pincode: string;
  relevance?: number;
};

export type ProjectDetail = Project & {
  promoter_pan?: string;
  promoter_gstin?: string;
  promoter_type?: string;
  contacts?: Contact[];
};

export type Contact = {
  id: number;
  role: string;
  name: string;
  phone: string;
  email: string;
};

export type Change = {
  field_name: string;
  old_value: string;
  new_value: string;
  detected_at: string;
};

export type Meta = { page: number; limit: number; total: number };

export type SearchResult = { data: Project[]; meta: Meta };

export type NLResult = {
  query: string;
  query_type: "projects" | "builders";
  interpreted: Record<string, string>;
  data: Project[] | Builder[];
  meta?: Meta;
};

export type Builder = {
  promoter_name: string;
  project_count: number;
  total_units: number;
};

export type StatusItem = { status: string; count: number };
export type BuilderStat = { promoter_name: string; project_count: number; total_units: number };
export type DistrictStat = { district: string; count: number };

export type CrawlRun = {
  id: number;
  started_at: string;
  finished_at: string;
  status: string;
  start_id: number;
  end_id: number;
  processed: number;
  failed: number;
};

export type Suggestion = { text: string; type: "project" | "promoter"; score: number };

// ── API calls ────────────────────────────────────────────────────────────────

export const api = {
  projects: {
    list(params: Record<string, string | number> = {}): Promise<SearchResult> {
      const q = new URLSearchParams(
        Object.fromEntries(Object.entries(params).map(([k, v]) => [k, String(v)]))
      ).toString();
      return get(`/projects${q ? `?${q}` : ""}`);
    },
    get(id: number): Promise<{ data: ProjectDetail }> {
      return get(`/projects/${id}`);
    },
    changes(id: number): Promise<{ data: Change[] }> {
      return get(`/projects/${id}/changes`);
    },
  },

  analytics: {
    statusDist(): Promise<{ data: StatusItem[] }> {
      return get("/analytics/status-distribution");
    },
    topBuilders(limit = 10): Promise<{ data: BuilderStat[] }> {
      return get(`/analytics/top-builders?limit=${limit}`);
    },
    byDistrict(limit = 10): Promise<{ data: DistrictStat[] }> {
      return get(`/analytics/by-district?limit=${limit}`);
    },
  },

  search: {
    nl(q: string): Promise<NLResult> {
      return get(`/search/nl?q=${encodeURIComponent(q)}`);
    },
    suggestions(q: string): Promise<{ data: Suggestion[] }> {
      return get(`/search/suggestions?q=${encodeURIComponent(q)}`);
    },
  },

  crawls: {
    list(): Promise<{ data: CrawlRun[] }> {
      return get("/crawls");
    },
  },
};
