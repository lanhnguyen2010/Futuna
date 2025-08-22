package openai

import (
	"context"

	oa "github.com/sashabaranov/go-openai"
)

// Client wraps the OpenAI API client.
type Client struct {
	api *oa.Client
}

// New creates a new Client.
func New(apiKey string) *Client {
	return &Client{api: oa.NewClient(apiKey)}
}

// AnalyzeVN30 requests analysis for all VN30 tickers using web_search.
func (c *Client) AnalyzeVN30(ctx context.Context) (string, error) {
	req := oa.ResponsesRequest{
		Model:       oa.GPT4oMini,
		Temperature: 0,
		Input: []oa.ResponseMessage{
			{
				Role:    oa.ChatMessageRoleSystem,
				Content: "You are a disciplined equity analyst for Vietnam (HOSE). You may use the web_search tool.\nRules:\n- No domain restrictions; search the open web.\n- Before any time-sensitive claim (news, flows, prices, events), use web_search.\n- Prefer data from the last 14 days; use older only for reports.\n- Output JSON ONLY. No prose. Return text in vietnamese",
			},
			{
				Role:    oa.ChatMessageRoleUser,
				Content: "Generate today’s 08:00 (GMT+7) HOSE analysis for all VN30 tickers. Include:\n- short_term (ACCUMULATE|HOLD|AVOID + confidence 0-100 + reason)\n- long_term (ACCUMULATE|HOLD|AVOID + confidence 0-100 + reason)\n- at least 5 strategies (name, stance FAVORABLE|NEUTRAL|UNFAVORABLE, note)\n- sources (URLs)",
			},
		},
		Tools: []oa.ToolDefinition{{Type: oa.ToolTypeWebSearch}},
	}
	resp, err := c.api.CreateResponse(ctx, req)
	if err != nil {
		return "", err
	}
	if len(resp.Output) == 0 {
		return "", nil
	}
	// The last item of the output contains the assistant message with the JSON.
	last := resp.Output[len(resp.Output)-1]
	if len(last.Content) == 0 {
		return "", nil
	}
	return last.Content[0].Text, nil
}
