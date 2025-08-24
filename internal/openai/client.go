package openai

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

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

func (c *Client) AnalyzeTickers(ctx context.Context, tickers []string) (string, string, string, error) {
	if len(tickers) > 5 {
		tickers = tickers[:5]
	}
	tickersList := strings.Join(tickers, ", ")
	log.Println("analysing: " + tickersList)

	// --- your JSON Schema unchanged ---

       systemPrompt := `You are a disciplined Vietnam equity analyst. Use the web_search tool for time-sensitive data.
       Output: JSON matching this exact structure with no extra text.
       {
               "as_of": "RFC3339 timestamp",
               "tickers": [
                 {
                       "ticker": "<mã>",
                       "short_term": {"recommendation": "ACCUMULATE|HOLD|AVOID", "confidence": <0-100>, "reason": "<lí do>"},
                       "long_term": {"recommendation": "ACCUMULATE|HOLD|AVOID", "confidence": <0-100>, "reason": "<lí do>"},
                       "strategies": [{"name": "<chiến lược>", "stance": "FAVORABLE|NEUTRAL|UNFAVORABLE", "note": "<ghi chú>"}],
                       "overall": {"recommendation": "ACCUMULATE|HOLD|AVOID", "confidence": <0-100>, "reason": "<lí do>"}
                 }
               ],
               "sources": ["<nguồn>"]
         }
       Rules:
       - Keep rationales concise (≤5 sentences).
       - Always include 1–3 valid URLs per ticker.
       - Prioritize T-1 trading data; if near token budget, keep T-1 + recommendations and shorten others.`
	vn30AnalysisPrompt := `Task: Generate today’s 08:00 (GMT+7) VN30 stock analysis for: ` + tickersList + `
	Return JSON with fields:
	- as_of
	- tickers: array (1 object per ticker, with short_term, long_term, strategies[], overall, sources[])
	Constraints:
	- Rationales must be concise (≤5 sentences).
	- Always fill "sources" with valid URLs.
	- If token budget exceeded, Prioritize  T-1 trading info + recommendations.
        `

	// Build the minimal typed params (only what we must type)
	params := responses.ResponseNewParams{
		Model: c.model, // string, not openai.String(...)
	}

	// Keep a copy of the request (with our extra JSON injected below)
	reqJSON, _ := json.Marshal(params)

	// Shared request options for the API call
	opts := []option.RequestOption{
		// Enable web search tool (try "web_search" first; some regions still use "..._preview")
		option.WithJSONSet("tools", []map[string]any{
			{"type": "web_search"},
		}),
		// Input messages
		option.WithJSONSet("input", []map[string]any{
			{
				"type": "message", "role": "system",
				"content": []map[string]any{
					{"type": "input_text", "text": systemPrompt},
				},
			},
			{
				"type": "message", "role": "user",
				"content": []map[string]any{
					{"type": "input_text", "text": vn30AnalysisPrompt},
				},
			},
		}),
	}

	var resp *responses.Response
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		resp, err = c.api.Responses.New(ctx, params, opts...)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "rate limit") {
				log.Println("rate limit reached, sleeping before retrying")
				time.Sleep(10 * time.Second)
				continue
			}
			return string(reqJSON), "", "", err
		}
		break
	}
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
