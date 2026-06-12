package nlsearch

// V7.2 LLM-Powered Natural Language Search
//
// LLMParser replaces keyword pattern-matching with a Claude Haiku API call.
// The model is given one tool — extract_search_filters — and is prompted to
// always call it. The tool input JSON is decoded directly into ParsedQuery,
// so the handler, SQL, and response shape stay 100% identical to V7.1.
//
// Cost: ~₹0.05 per query (claude-haiku-4-5 at $1/$5 per 1M in/out tokens).
// The calling code in main.go wraps this in a FallbackParser so any API
// error transparently degrades to the rule-based parser.

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// llmFilterInput is the JSON schema the model fills via tool use.
type llmFilterInput struct {
	QueryType   string `json:"query_type"`   // "projects" | "builders"
	Status      string `json:"status"`       // "Ongoing" | "New" | ... | ""
	Type        string `json:"type"`         // "Residential" | ... | ""
	District    string `json:"district"`     // e.g. "Pune", ""
	City        string `json:"city"`         // fallback location field
	Delayed     bool   `json:"delayed"`      // true → proposed > original completion
	MinProjects int    `json:"min_projects"` // for builder queries
}

// LLMParser calls Claude Haiku to extract structured filters from free-form text.
type LLMParser struct {
	client *anthropic.Client
}

// NewLLMParser creates a V7.2 LLM parser using the provided Anthropic API key.
func NewLLMParser(apiKey string) *LLMParser {
	c := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &LLMParser{client: &c}
}

// nlSystemPrompt explains the MahaRERA schema to the model.
const nlSystemPrompt = `You are a search query parser for InfraLens, a MahaRERA real-estate project intelligence platform.

Your only job is to call the extract_search_filters tool with the structured filters that match the user's query.

Field rules:
- query_type: set "builders" when the query asks about builders/promoters/developers; otherwise "projects"
- status: one of "Ongoing", "New", "Lapsed", "Completed", "Under Approval" — empty string if not mentioned. Use "Lapsed" for queries like "stalled", "abandoned", "expired", "dead", "inactive". Use "Ongoing" for "active", "in progress", "current".
- type: one of "Residential", "Commercial", "Plotted Development", "Retail", "Mixed Development" — empty string if not mentioned
- district: city or district name (title-cased, e.g. "Pune", "Thane") if a location is mentioned — empty string otherwise
- delayed: true when the query implies a project has overrun its deadline — e.g. "delayed", "overdue", "past due", "missed deadline", "missed their deadline", "behind schedule", "past completion", "running late", "past original date"
- min_projects: numeric threshold for builder queries (0 if not mentioned)

Always call the tool. Never respond with plain text.`

// nlSearchTool is the single tool the model is prompted to always call.
var nlSearchTool = anthropic.ToolParam{
	Name:        "extract_search_filters",
	Description: anthropic.String("Extract structured MahaRERA project search filters from a natural language query."),
	InputSchema: anthropic.ToolInputSchemaParam{
		Properties: map[string]any{
			"query_type": map[string]any{
				"type":        "string",
				"enum":        []string{"projects", "builders"},
				"description": "Whether to search for projects or builders/promoters",
			},
			"status": map[string]any{
				"type":        "string",
				"enum":        []string{"Ongoing", "New", "Lapsed", "Completed", "Under Approval", ""},
				"description": "Project registration status. Use 'Lapsed' for: stalled, abandoned, expired, dead, inactive, cancelled. Use 'Ongoing' for: active, in progress, current, running, live.",
			},
			"type": map[string]any{
				"type":        "string",
				"enum":        []string{"Residential", "Commercial", "Plotted Development", "Retail", "Mixed Development", ""},
				"description": "Construction project type",
			},
			"district": map[string]any{
				"type":        "string",
				"description": "District or city name (title-cased). Empty string if not mentioned.",
			},
			"city": map[string]any{
				"type":        "string",
				"description": "Alternative city field. Use district instead when possible.",
			},
			"delayed": map[string]any{
				"type":        "boolean",
				"description": "True if the query implies projects overran their deadline. Set true for: 'delayed', 'overdue', 'past due', 'missed deadline', 'missed their deadline', 'behind schedule', 'running late', 'past completion', 'past original date', 'past their deadline'.",
			},
			"min_projects": map[string]any{
				"type":        "integer",
				"description": "Minimum project count for builder queries. 0 if not specified.",
			},
		},
	},
}

// Parse calls Claude Haiku with the natural language query and decodes the
// tool-use response into a ParsedQuery compatible with the rest of the stack.
func (p *LLMParser) Parse(ctx context.Context, q string) (ParsedQuery, error) {
	resp, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 512,
		System: []anthropic.TextBlockParam{
			{Text: nlSystemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(q)),
		},
		Tools: []anthropic.ToolUnionParam{
			{OfTool: &nlSearchTool},
		},
	})
	if err != nil {
		return ParsedQuery{}, fmt.Errorf("llm parser: api call: %w", err)
	}

	// Find the tool-use block in the response and decode it.
	for _, block := range resp.Content {
		if tu, ok := block.AsAny().(anthropic.ToolUseBlock); ok {
			var in llmFilterInput
			if err := json.Unmarshal([]byte(tu.JSON.Input.Raw()), &in); err != nil {
				return ParsedQuery{}, fmt.Errorf("llm parser: decode tool input: %w", err)
			}
			return buildParsedQuery(q, in), nil
		}
	}

	return ParsedQuery{}, fmt.Errorf("llm parser: no tool_use block in response (stop_reason=%s)", resp.StopReason)
}

// buildParsedQuery converts the LLM's structured output into a ParsedQuery
// that is wire-compatible with what the RuleParser produces.
func buildParsedQuery(raw string, in llmFilterInput) ParsedQuery {
	pq := ParsedQuery{
		Raw:         raw,
		QueryType:   in.QueryType,
		Interpreted: map[string]string{},
	}
	if pq.QueryType == "" {
		pq.QueryType = "projects"
	}

	if pq.QueryType == "builders" {
		pq.MinProjects = in.MinProjects
		if in.MinProjects > 0 {
			pq.Interpreted["min_projects"] = strconv.Itoa(in.MinProjects)
		}
		return pq
	}

	// Project-level filters
	if in.Status != "" {
		s := in.Status
		pq.Filter.Status = &s
		pq.Interpreted["status"] = in.Status
	}
	if in.Type != "" {
		t := in.Type
		pq.Filter.Type = &t
		pq.Interpreted["type"] = in.Type
	}

	// Prefer district; fall back to city if the model used that field instead
	district := in.District
	if district == "" {
		district = in.City
	}
	if district != "" {
		pq.Filter.District = &district
		pq.Interpreted["district"] = district
	}

	if in.Delayed {
		b := true
		pq.Filter.Delayed = &b
		pq.Interpreted["filter"] = "proposed_completion_date > original_completion_date"
	}

	return pq
}

// Ensure LLMParser satisfies the Parser interface at compile time.
var _ Parser = (*LLMParser)(nil)
