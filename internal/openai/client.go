package openai

import (
	"context"
	"encoding/json"

	oa "github.com/sashabaranov/go-openai"
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
				Content: "Generate today’s 08:00 (GMT+7) HOSE analysis for all HOSE tickers. Include:\n- short_term (ACCUMULATE|HOLD|AVOID + confidence 0-100 + reason)\n- long_term (ACCUMULATE|HOLD|AVOID + confidence 0-100 + reason)\n- at least 5 strategies (name, stance FAVORABLE|NEUTRAL|UNFAVORABLE, note)\n- sources (URLs)",
			},
		},
		Tools: []oa.ToolDefinition{{Type: oa.ToolTypeWebSearch}},
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
