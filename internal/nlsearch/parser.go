// Package nlsearch implements natural language search parsing for InfraLens.
//
// V7.1 — RuleParser: zero-cost, instant, pattern-matching.
// V7.2 — LLMParser:  Claude API (Haiku), richer intent extraction (~₹0.05/query).
//
// Both implement the Parser interface and return the same ParsedQuery type,
// so the handler, SQL, and response shape are identical regardless of which
// parser is active. main.go selects the parser based on ANTHROPIC_API_KEY.
package nlsearch

import (
	"context"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/infralens/infralens/internal/core"
)

// Parser converts a natural-language query into a structured ParsedQuery.
// V7.1: NewRuleParser() — no external calls, always free.
// V7.2: NewLLMParser()  — Claude Haiku API, ~₹0.05 per query.
type Parser interface {
	Parse(ctx context.Context, q string) (ParsedQuery, error)
}

// ParsedQuery holds the extracted intent from a natural-language query.
// Filter is passed straight to the search layer; Interpreted is returned
// in the API response so callers can show "Searched for: status=Ongoing, district=Pune".
type ParsedQuery struct {
	Raw         string            `json:"raw"`
	QueryType   string            `json:"query_type"`  // "projects" | "builders"
	Interpreted map[string]string `json:"interpreted"` // human-readable breakdown
	Filter      core.SearchFilter `json:"-"`
	MinProjects int               `json:"-"` // only set when QueryType == "builders"
}

// ── Rule-Based Parser (V7.1) ─────────────────────────────────────────────────

// RuleParser implements Parser using keyword/regex rules — no LLM required.
type RuleParser struct{}

// NewRuleParser returns a V7.1 rule-based parser.
func NewRuleParser() Parser { return &RuleParser{} }

var statusRules = []struct{ pattern, value string }{
	{"under approval", "Under Approval"}, // must come before "ongoing"
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

var locationRe = regexp.MustCompile(`\b(?:in|at|near)\s+([a-zA-Z][a-zA-Z ]{0,25})`)
var minProjectsRe = regexp.MustCompile(`(?:more than|at least|over|greater than)\s+(\d+)`)

// Parse implements Parser using keyword rules.
// Rules applied in order:
//  1. Builder/promoter/developer → builder-count query (exits early).
//  2. Status keywords in priority order.
//  3. Project type keywords (residential, commercial, …).
//  4. First "in/at/near <place>" phrase → district filter.
//  5. Delay keywords → Delayed=true (proposed > original completion date).
func (r *RuleParser) Parse(_ context.Context, q string) (ParsedQuery, error) {
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
		return pq, nil
	}

	// 2. Status
	for _, rule := range statusRules {
		if strings.Contains(lower, rule.pattern) {
			s := rule.value
			pq.Filter.Status = &s
			pq.Interpreted["status"] = rule.value
			break
		}
	}

	// 3. Project type
	for _, rule := range typeRules {
		if strings.Contains(lower, rule.pattern) {
			t := rule.value
			pq.Filter.Type = &t
			pq.Interpreted["type"] = rule.value
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

	return pq, nil
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

// ── Fallback Parser ──────────────────────────────────────────────────────────

// NewFallbackParser wraps primary with fallback: on any error from primary
// it logs and transparently delegates to fallback. Used in main.go to wrap
// LLMParser with RuleParser so quota/network errors are never user-visible.
func NewFallbackParser(primary, fallback Parser) Parser {
	return &fallbackParser{primary: primary, fallback: fallback}
}

type fallbackParser struct {
	primary  Parser
	fallback Parser
}

func (f *fallbackParser) Parse(ctx context.Context, q string) (ParsedQuery, error) {
	pq, err := f.primary.Parse(ctx, q)
	if err != nil {
		log.Printf("[NL] LLM parser error (%v) — falling back to V7.1 rule-based", err)
		return f.fallback.Parse(ctx, q)
	}
	return pq, nil
}
