// Package nlsearch implements V7.1 Rule-Based Natural Language Search.
//
// Parse() converts a plain-English query into a structured ParsedQuery that
// carries a core.SearchFilter ready to pass straight to the search layer.
// No LLM or external dependency is required — intent is extracted by matching
// known keyword patterns against the lowercased input.
//
// V7.2 will replace Parse() with an LLM-backed implementation that returns
// the same ParsedQuery type, keeping the rest of the stack unchanged.
package nlsearch

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/infralens/infralens/internal/core"
)

// ParsedQuery holds the extracted intent from a natural-language query.
type ParsedQuery struct {
	Raw         string            `json:"raw"`
	QueryType   string            `json:"query_type"`  // "projects" | "builders"
	Interpreted map[string]string `json:"interpreted"` // human-readable breakdown shown in the API response
	Filter      core.SearchFilter `json:"-"`
	MinProjects int               `json:"-"` // only set when QueryType == "builders"
}

// --- pattern tables ---

var statusRules = []struct{ pattern, value string }{
	{"under approval", "Under Approval"}, // must come before "ongoing" / generic checks
	{"ongoing", "Ongoing"},
	{"new project", "New"},
	{"lapsed", "Lapsed"},
	{"completed", "Completed"},
}

var typeRules = []struct{ pattern, value string }{
	{"residential", "Residential"},
	{"commercial", "Commercial"},
	{"plotted", "Plotted Development"},
	{"retail", "Retail"},
	{"mixed", "Mixed Development"},
}

var delayedKeywords = []string{
	"delayed beyond original",
	"past original",
	"beyond original",
	"overdue",
	"delayed",
}

// locationRe matches "in Pune", "in Navi Mumbai", "at Thane", "near Nagpur".
// Capture group 1 is the raw location text (may be multiple words).
var locationRe = regexp.MustCompile(`\b(?:in|at|near)\s+([a-zA-Z][a-zA-Z ]{0,25})`)

// minProjectsRe matches "more than 10", "at least 5", "over 20", etc.
var minProjectsRe = regexp.MustCompile(`(?:more than|at least|over|greater than)\s+(\d+)`)

// Parse converts a natural-language query string into a ParsedQuery.
//
// Rules applied in order:
//  1. If the query mentions "builder / promoter / developer" it becomes a
//     builder-count query — all other rules are skipped.
//  2. Status keywords are matched against a priority-ordered list.
//  3. Project type keywords (residential, commercial, …) are matched.
//  4. The first "in / at / near <place>" phrase sets the district filter.
//  5. Delay keywords set Delayed=true → SQL: proposed > original completion.
func Parse(q string) ParsedQuery {
	lower := strings.ToLower(strings.TrimSpace(q))
	pq := ParsedQuery{
		Raw:         q,
		QueryType:   "projects",
		Interpreted: map[string]string{},
	}

	// 1. Builder / promoter queries — different result shape, exit early
	if strings.Contains(lower, "builder") ||
		strings.Contains(lower, "promoter") ||
		strings.Contains(lower, "developer") {

		pq.QueryType = "builders"
		if m := minProjectsRe.FindStringSubmatch(lower); len(m) >= 2 {
			if n, err := strconv.Atoi(m[1]); err == nil {
				pq.MinProjects = n
				pq.Interpreted["min_projects"] = strconv.Itoa(n)
			}
		}
		return pq
	}

	// 2. Status
	for _, r := range statusRules {
		if strings.Contains(lower, r.pattern) {
			s := r.value
			pq.Filter.Status = &s
			pq.Interpreted["status"] = r.value
			break
		}
	}

	// 3. Project type
	for _, r := range typeRules {
		if strings.Contains(lower, r.pattern) {
			t := r.value
			pq.Filter.Type = &t
			pq.Interpreted["type"] = r.value
			break
		}
	}

	// 4. Location
	if loc := extractLocation(lower); loc != "" {
		pq.Filter.District = &loc
		pq.Interpreted["district"] = loc
	}

	// 5. Delayed beyond original completion date
	for _, kw := range delayedKeywords {
		if strings.Contains(lower, kw) {
			b := true
			pq.Filter.Delayed = &b
			pq.Interpreted["filter"] = "proposed_completion_date > original_completion_date"
			break
		}
	}

	return pq
}

// extractLocation pulls the place name from the first "in/at/near <place>"
// phrase and returns it title-cased (e.g. "pune" → "Pune").
func extractLocation(s string) string {
	m := locationRe.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	words := strings.Fields(strings.TrimSpace(m[1]))
	if len(words) == 0 {
		return ""
	}
	// Take only the first word to avoid eating trailing verbs ("in Pune that are…")
	word := words[0]
	return strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
}
