package openai

import (
	"context"
	"encoding/json"

	oa "github.com/sashabaranov/go-openai/v2"
)

// Client wraps the OpenAI API client.
type Client struct {
	api   *oa.Client
	model string
}

// New creates a new Client.
func New(apiKey, baseURL, model string) *Client {
	cfg := oa.DefaultConfig(apiKey)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	return &Client{api: oa.NewClientWithConfig(cfg), model: model}
}

// AnalyzeTickers requests analysis for all tickers using web_search.
func (c *Client) AnalyzeTickers(ctx context.Context) (string, string, string, error) {
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

	req := oa.ResponsesRequest{
		Model:       c.model,
		Temperature: 0,
		Input: []oa.ResponseMessage{
			{
				Role:    oa.ChatMessageRoleSystem,
				Content: "You are a disciplined equity analyst for Vietnam (HOSE). You may use the web_search tool.\nRules:\n- No domain restrictions; search the open web.\n- Before any time-sensitive claim (news, flows, prices, events), use web_search.\n- Prefer data from the last 14 days; use older only for reports.\n- Output JSON ONLY. No prose. Return text in vietnamese",
			},
			{
				Role:    oa.ChatMessageRoleUser,
				Content: "Generate today’s 08:00 (GMT+7) HOSE analysis for all HOSE tickers. For each ticker, provide:\n1. Short-Term View\n   - Recommendation: (ACCUMULATE | HOLD | AVOID)\n   - Confidence (0-100)\n   - Reason\n2. Long-Term View\n   - Recommendation: (ACCUMULATE | HOLD | AVOID)\n   - Confidence (0-100)\n   - Reason\n3. Strategy-Based Analysis (at least 5 technical/fundamental strategies)\n   For each strategy:\n   - Strategy Name (e.g., Moving Average Crossover, RSI, Fibonacci, Momentum, Bollinger Bands, Fundamental PE Valuation, etc.)\n   - Stance: (FAVORABLE | NEUTRAL | UNFAVORABLE)\n   - Note: Explain why this stance is taken (indicator values, patterns, or signals observed).\n4. Final Overall Recommendation\n   - Aggregate the above strategy stances into a single recommendation (ACCUMULATE | HOLD | AVOID).\n   - Provide confidence score (0-100).\n   - Justify by balancing short-term vs long-term outlook and strategy signals.\n5. Sources\n   - Include relevant URLs used in the analysis.",
			},
		},
		Tools: []oa.ToolDefinition{{Type: oa.ToolTypeWebSearch}},
		ResponseFormat: &oa.ResponseFormat{
			Type: oa.ResponseFormatTypeJSONSchema,
			JSONSchema: &oa.JSONSchemaDefinition{
				Name:   "analysis_response",
				Schema: schema,
				Strict: true,
			},
		},
	}
	reqJSON, _ := json.Marshal(req)
	resp, err := c.api.CreateResponse(ctx, req)
	if err != nil {
		return string(reqJSON), "", "", err
	}
	respJSON, _ := json.Marshal(resp)
	if len(resp.Output) == 0 {
		return string(reqJSON), string(respJSON), "", nil
	}
	// The last item of the output contains the assistant message with the JSON.
	last := resp.Output[len(resp.Output)-1]
	if len(last.Content) == 0 {
		return string(reqJSON), string(respJSON), "", nil
	}
	return string(reqJSON), string(respJSON), last.Content[0].Text, nil
}
