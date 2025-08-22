package openai

import (
	"context"
	"encoding/json"

	openai "github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/responses"
)

type Client struct {
	api   openai.Client // value, not pointer
	model string
}

func New(apiKey, baseURL, model string) *Client {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	return &Client{api: openai.NewClient(opts...), model: model}
}

func (c *Client) AnalyzeTickers(ctx context.Context) (string, string, string, error) {
	// --- your JSON Schema unchanged ---
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"as_of": map[string]any{"type": "string"},
			"tickers": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"ticker": map[string]any{"type": "string"},
						"short_term": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"recommendation": map[string]any{"type": "string", "enum": []string{"ACCUMULATE", "HOLD", "AVOID"}},
								"confidence":     map[string]any{"type": "integer"},
								"reason":         map[string]any{"type": "string"},
							},
							"required": []string{"recommendation", "confidence", "reason"},
						},
						"long_term": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"recommendation": map[string]any{"type": "string", "enum": []string{"ACCUMULATE", "HOLD", "AVOID"}},
								"confidence":     map[string]any{"type": "integer"},
								"reason":         map[string]any{"type": "string"},
							},
							"required": []string{"recommendation", "confidence", "reason"},
						},
						"strategies": map[string]any{
							"type":     "array",
							"minItems": 5,
							"items": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"name":   map[string]any{"type": "string"},
									"stance": map[string]any{"type": "string", "enum": []string{"FAVORABLE", "NEUTRAL", "UNFAVORABLE"}},
									"note":   map[string]any{"type": "string"},
								},
								"required": []string{"name", "stance", "note"},
							},
						},
						"overall": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"recommendation": map[string]any{"type": "string", "enum": []string{"ACCUMULATE", "HOLD", "AVOID"}},
								"confidence":     map[string]any{"type": "integer"},
								"reason":         map[string]any{"type": "string"},
							},
							"required": []string{"recommendation", "confidence", "reason"},
						},
					},
					"required": []string{"ticker", "short_term", "long_term", "strategies", "overall"},
				},
			},
			"sources": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
		},
		"required": []string{"as_of", "tickers", "sources"},
	}

	// Build the minimal typed params (only what we must type)
	params := responses.ResponseNewParams{
		Model: c.model, // string, not openai.String(...)
	}

	// Keep a copy of the request (with our extra JSON injected below)
	reqJSON, _ := json.Marshal(params)

	// Call with JSON overrides:
	// - tools: enable web_search (preview)
	// - response_format: structured outputs via our schema
	// - input: system + user messages
	resp, err := c.api.Responses.New(
		ctx,
		params,
		// Tools (prefer the dated variant if plain preview errors in your region)
		option.WithJSONSet("tools", []map[string]any{
			{"type": "web_search_preview"},
			// or: {"type": "web_search_preview_2025_03_11"},
		}),
		// Structured Outputs
		option.WithJSONSet("response_format", map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   "analysis_response",
				"schema": schema,
				"strict": true,
			},
		}),
		// Input messages
		option.WithJSONSet("input", []map[string]any{
			{
				"type": "message", "role": "system",
				"content": []map[string]any{
					{"type": "input_text", "text": "You are a disciplined equity analyst for Vietnam (HOSE). You may use the web_search tool.\nRules:\n- No domain restrictions; search the open web.\n- Before any time-sensitive claim (news, flows, prices, events), use web_search.\n- Prefer data from the last 14 days; use older only for reports.\n- Output JSON ONLY. No prose. Return text in vietnamese"},
				},
			},
			{
				"type": "message", "role": "user",
				"content": []map[string]any{
					{"type": "input_text", "text": "Generate today’s 08:00 (GMT+7) HOSE analysis for all HOSE tickers. For each ticker, provide:\n1. Short-Term View\n   - Recommendation: (ACCUMULATE | HOLD | AVOID)\n   - Confidence (0-100)\n   - Reason\n2. Long-Term View\n   - Recommendation: (ACCUMULATE | HOLD | AVOID)\n   - Confidence (0-100)\n   - Reason\n3. Strategy-Based Analysis (at least 5 technical/fundamental strategies)\n   For each strategy:\n   - Strategy Name (e.g., Moving Average Crossover, RSI, Fibonacci, Momentum, Bollinger Bands, Fundamental PE Valuation, etc.)\n   - Stance: (FAVORABLE | NEUTRAL | UNFAVORABLE)\n   - Note: Explain why this stance is taken (indicator values, patterns, or signals observed).\n4. Final Overall Recommendation\n   - Aggregate the above strategy stances into a single recommendation (ACCUMULATE | HOLD | AVOID).\n   - Provide confidence score (0-100).\n   - Justify by balancing short-term vs long-term outlook and strategy signals.\n5. Sources\n   - Include relevant URLs used in the analysis."},
				},
			},
		}),
	)
	if err != nil {
		return string(reqJSON), "", "", err
	}

	respJSON, _ := json.Marshal(resp)

	// Prefer the SDK helper if present; otherwise read last output text
	out := resp.OutputText() // concatenated text, when available
	if out == "" && len(resp.Output) > 0 && len(resp.Output[0].Content) > 0 {
		out = resp.Output[0].Content[0].Text
	}

	return string(reqJSON), string(respJSON), out, nil
}
