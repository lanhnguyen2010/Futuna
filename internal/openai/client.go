package openai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
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

// BatchResult holds the request, response and extracted output text for a batch item.
type BatchResult struct {
	Request  string
	Response string
	Output   string
}

func New(apiKey, baseURL, model string) *Client {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	return &Client{api: openai.NewClient(opts...), model: model}
}

// buildRequest constructs the request body for analysing the provided tickers and
// returns the body along with its JSON encoding.
func (c *Client) buildRequest(tickers []string) (map[string]any, string) {
	if len(tickers) > 2 {
		tickers = tickers[:2]
	}
	tickersList := strings.Join(tickers, ", ")
	log.Println("analysing: " + tickersList)

	systemPrompt := `Bạn là chuyên gia phân tích chứng khoán Việt Nam. Nhiệm vụ: tạo báo cáo hiện tại {YYYY_MM_DD} (GMT+7) cho rổ VN30 theo cấu trúc 0→7.
Bạn được phép dùng công cụ web-search.
Quy tắc bắt buộc:

Chỉ tiếng Việt. Không chèn HTML.

Phạm vi dữ liệu:

Giá/khối lượng/khối ngoại cập nhật đến hết phiên T-1 (ngày {T_MINUS_1_YYYY_MM_DD}).

Tin tức & sự kiện trong 30 ngày gần nhất tính đến hiện tại {YYYY_MM_DD} (GMT+7).

BCTC 4 quý gần nhất và 3 năm gần nhất (ưu tiên báo cáo kiểm toán/soát xét).

Phái sinh liên quan: VN30F1M (basis, OI, chênh lệch với cơ sở) — ghi rõ ngày/khung thời gian.

Định giá: P/E (ttm/fwd), P/B, EV/EBITDA, PEG, dividend yield; so sánh với trung vị ngành nếu có nguồn đáng tin. Nếu không có, nêu rõ “(chưa sẵn công khai)”.

Kỹ thuật: nếu có chuỗi giá đủ dài thì tự tính ATR(14), biến động 30 ngày, MA/RSI/Bollinger;.

Xếp hạng:

Recommendation ∈ {ACCUMULATE, HOLD, AVOID}

Confidence ∈ [0,100] và kèm Reason rõ ràng.

Định dạng đầu ra: theo mỗi mã (ví dụ: VCB, VNM, …) và đúng mục 0→7 dưới đây. Dùng gạch đầu dòng, ghi đơn vị & kỳ tham chiếu.

Không đưa code, không JSON trong phần kết quả; chỉ là văn bản định dạng như yêu cầu.

Khi dùng web-search:

Ưu tiên: HSX/HOSE, HNX, SSC, SSI, Vietstock, CafeF, DN IR/Investors Relations, sở giao dịch phái sinh, TradingView/Investing (cho giá/lịch sử), CTCK uy tín (nguồn báo cáo), cố gắng tìm dữ liệu giao dịch cần thiết.

Giá/khối lượng T-1 (Investing):
site:investing.com {tên công ty tiếng Anh} historical data → mở trang “Historical Data” để lấy OHLC/Volume.
Investing.com

Khối ngoại theo ngày (Vietstock/CafeF):
site:finance.vietstock.vn "{TICKER}" "giao dịch nhà đầu tư nước ngoài" hoặc
site:cafef.vn "{TICKER}" "khối ngoại" "{YYYY-MM-DD}"
VietstockFinance
cafef

Room ngoại:
site:cafef.vn "room ngoại" "{TICKER}"
cafef

Phái sinh VN30F1M (OI/basis):
site:hnx.vn VN30F1M (tổng quan) + site:finance.vietstock.vn VN30F1M (chi tiết) + site:tradingview.com VN30 futures I1 (OI tham khảo).
HNX
VietstockFinance
TradingView

Định giá/ratios:
site:marketscreener.com {company} "Valuation ratios"
MarketScreener India

BCTC/CBTT:
site:congbothongtin.ssc.gov.vn {mã chứng khoán} "Báo cáo tài chính"
congbothongtin.ssc.gov.vn`

	vn30AnalysisPrompt := `
        Hãy **tạo bản phân tích hiện tại (GMT+7)** cho rổ VN30 trên HOSE**, danh sách mã:` + tickersList + `. Với **mỗi mã**, hãy thực hiện **điều tra dữ liệu** và xuất kết quả **chỉ bằng tiếng Việt** theo cấu trúc dưới đây. **Dữ liệu sử dụng phải cập nhật đến hết phiên T-1**, còn tin tức & sự kiện trong **30 ngày gần nhất**; báo cáo tài chính **4 quý gần nhất và 3 năm gần nhất**.

        ### 0) Tóm tắt dữ liệu đầu vào (cho từng mã)
        - **Thời điểm dữ liệu**: ghi rõ timestamp nguồn (giờ Việt Nam).
        - **Giá & thị trường (T-1)**: O/H/L/C, % thay đổi, khối lượng, **GTGD**, **VWAP**, **ATR(14)**, **biến động 30 ngày**, **kháng cự/hỗ trợ** gần nhất, **gap** nếu có.
        - **Dòng tiền nước ngoài (T-1 và 5/20 phiên)**: **khối lượng & giá trị** **mua ròng/bán ròng**, **tỷ lệ so với GTGD**, **room ngoại** còn lại (%).
        - **Phái sinh liên quan** (nếu có): basis, OI, tín hiệu chênh lệch với cơ sở.
        - **Định giá hiện tại**: **P/E (ttm/fwd)**, **P/B**, **EV/EBITDA**, **PEG**, **dividend yield**, so sánh với **trung vị ngành**.
        - **Cơ bản tài chính** (từ BCTC kiểm toán/soát xét): doanh thu, LNST, **biên gộp**, **biên EBIT/NP**, **ROE/ROA**, **nợ vay ròng/EBITDA**, **hệ số thanh toán**, **chu kỳ tiền mặt**, **CAPEX**, **OCF/FCF** (YoY & QoQ), **tăng trưởng 3 năm CAGR**.
        - **Sự kiện/catalyst** gần đây: KQKD, chia cổ tức/ESOP, M&A, phát hành, thay đổi quản trị, **tin đồn đã kiểm chứng**.
        - **Rủi ro**: pháp lý, hàng tồn kho, khách hàng/nhà cung cấp lớn, rủi ro tỷ giá/lãi suất.

        ### 1) Short-Term View (1–4 tuần)
        - **Recommendation**: (ACCUMULATE | HOLD | AVOID)
        - **Confidence (0–100)**
        - **Reason**: nêu rõ luận điểm dựa trên **tín hiệu kỹ thuật**, **dòng tiền ngoại**, **biến động thị trường**, **tin tức ngắn hạn**.

        ### 2) Long-Term View (6–24 tháng)
        - **Recommendation**: (ACCUMULATE | HOLD | AVOID)
        - **Confidence (0–100)**
        - **Reason**: cân nhắc **nền tảng cơ bản**, **chu kỳ ngành**, **định giá**, **dòng tiền tự do**, **đòn bẩy**, **chính sách cổ tức**.

        ### 3) Strategy-Based Analysis (ít nhất 5 chiến lược kỹ thuật/cơ bản)
        - **Strategy Name** (ví dụ: Moving Average Crossover, RSI, Fibonacci, Momentum, Bollinger Bands, Ichimoku, Volume Profile, Fundamental PE/EVEBITDA Valuation, FCF Yield, Growth-Quality-Valuation, v.v.)
        - **Stance**: (FAVORABLE | NEUTRAL | UNFAVORABLE)
        - **Note**: giải thích **chỉ báo/giá trị/thiết lập** đang quan sát.

        ### 4) Phân tích Dòng tiền & Báo cáo tài chính
        - Doanh thu & LNST: YoY/QoQ, động lực chính.
        - Biên lợi nhuận: gộp/EBIT/NP và xu hướng.
        - Dòng tiền: **OCF**, **CAPEX**, **FCF**.
        - Đòn bẩy & thanh khoản: **Nợ ròng/EBITDA**, **ICR**, **Current ratio/Quick ratio**.
        - Định giá: P/E, EV/EBITDA, P/B so với ngành & lịch sử.
        - Chính sách cổ tức.
        - Nhận định ảnh hưởng tới khuyến nghị.

        ### 5) Phân tích Giao dịch Nước ngoài & Thị trường
        - Mua/bán ròng: T-1, 5 phiên, 20 phiên.
        - Room ngoại còn lại.
        - Tín hiệu thị trường: VWAP, thanh khoản, ATR, độ rộng ngành.

        ### 6) Final Overall Recommendation
        - **Khuyến nghị tổng hợp**: (ACCUMULATE | HOLD | AVOID)
        - **Confidence (0–100)**
        - **Justification**: cân bằng short-term vs long-term và các chiến lược.

        **Yêu cầu định dạng đầu ra:**
        - Viết theo từng mã VN30.
        - Chỉ tiếng Việt. Không chèn HTML.
        `

	schema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"as_of": map[string]any{
				"type":   "string",
				"format": "date-time",
			},
			"tickers": map[string]any{
				"type":     "array",
				"minItems": 1,
				"maxItems": 2,
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"ticker": map[string]any{
							"type":    "string",
							"pattern": "^[A-Z]{3,5}$",
						},
						"short_term": map[string]any{
							"type":                 "object",
							"additionalProperties": false,
							"properties": map[string]any{
								"recommendation": map[string]any{"type": "string", "enum": []string{"ACCUMULATE", "HOLD", "AVOID"}},
								"confidence":     map[string]any{"type": "integer", "minimum": 0, "maximum": 100},
								"reason":         map[string]any{"type": "string", "minLength": 1},
							},
							"required": []string{"recommendation", "confidence", "reason"},
						},
						"long_term": map[string]any{
							"type":                 "object",
							"additionalProperties": false,
							"properties": map[string]any{
								"recommendation": map[string]any{"type": "string", "enum": []string{"ACCUMULATE", "HOLD", "AVOID"}},
								"confidence":     map[string]any{"type": "integer", "minimum": 0, "maximum": 100},
								"reason":         map[string]any{"type": "string", "minLength": 1},
							},
							"required": []string{"recommendation", "confidence", "reason"},
						},
						"strategies": map[string]any{
							"type":     "array",
							"minItems": 5,
							"items": map[string]any{
								"type":                 "object",
								"additionalProperties": false,
								"properties": map[string]any{
									"name":   map[string]any{"type": "string", "minLength": 1},
									"stance": map[string]any{"type": "string", "enum": []string{"FAVORABLE", "NEUTRAL", "UNFAVORABLE"}},
									"note":   map[string]any{"type": "string", "minLength": 1},
								},
								"required": []string{"name", "stance", "note"},
							},
						},
						"overall": map[string]any{
							"type":                 "object",
							"additionalProperties": false,
							"properties": map[string]any{
								"recommendation": map[string]any{"type": "string", "enum": []string{"ACCUMULATE", "HOLD", "AVOID"}},
								"confidence":     map[string]any{"type": "integer", "minimum": 0, "maximum": 100},
								"reason":         map[string]any{"type": "string", "minLength": 1},
							},
							"required": []string{"recommendation", "confidence", "reason"},
						},
					},
					"required": []string{"ticker", "short_term", "long_term", "strategies", "overall"},
				},
			},
			"sources": map[string]any{
				"type":     "array",
				"minItems": 1,
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []string{"as_of", "tickers", "sources"},
	}

	tools := []map[string]any{{"type": "web_search"}}
	cache := map[string]any{"type": "ephemeral"}
	textFormat := map[string]any{
		"type":   "json_schema",
		"schema": schema,
		"name":   "analysis_response",
	}
	input := []map[string]any{
		{
			"type": "message", "role": "system",
			"content": []map[string]any{{"type": "input_text", "text": systemPrompt}},
		},
		{
			"type": "message", "role": "user",
			"content": []map[string]any{{"type": "input_text", "text": vn30AnalysisPrompt}},
		},
	}

	body := map[string]any{
		"model":         c.model,
		"tools":         tools,
		"cache_control": cache,
		"text":          map[string]any{"format": textFormat},
		"input":         input,
	}

	reqJSON, _ := json.Marshal(body)
	return body, string(reqJSON)
}

// AnalyzeTickers sends a single analysis request using the Responses API.
func (c *Client) AnalyzeTickers(ctx context.Context, tickers []string) (string, string, string, error) {
	body, reqJSON := c.buildRequest(tickers)

	params := responses.ResponseNewParams{Model: c.model}
	resp, err := c.api.Responses.New(
		ctx,
		params,
		option.WithJSONSet("tools", body["tools"]),
		option.WithJSONSet("cache_control", body["cache_control"]),
		option.WithJSONSet("text.format", body["text"].(map[string]any)["format"]),
		option.WithJSONSet("input", body["input"]),
	)
	if err != nil {
		return reqJSON, "", "", err
	}

	respJSON, _ := json.Marshal(resp)
	out := resp.OutputText()
	if out == "" && len(resp.Output) > 0 && len(resp.Output[0].Content) > 0 {
		out = resp.Output[0].Content[0].Text
	}
	return reqJSON, string(respJSON), out, nil
}

// AnalyzeTickersBatch submits multiple analysis requests using the Batch API.
func (c *Client) AnalyzeTickersBatch(ctx context.Context, batches [][]string) ([]BatchResult, error) {
	tmp, err := os.CreateTemp("", "batch-*.jsonl")
	if err != nil {
		return nil, err
	}
	enc := json.NewEncoder(tmp)
	reqs := make([]string, len(batches))
	for i, tickers := range batches {
		body, reqJSON := c.buildRequest(tickers)
		line := map[string]any{
			"custom_id": fmt.Sprintf("req_%d", i),
			"method":    "POST",
			"url":       "/v1/responses",
			"body":      body,
		}
		if err := enc.Encode(line); err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
			return nil, err
		}
		reqs[i] = reqJSON
	}
	tmp.Close()

	f, err := os.Open(tmp.Name())
	if err != nil {
		os.Remove(tmp.Name())
		return nil, err
	}
	defer os.Remove(tmp.Name())

	fileObj, err := c.api.Files.New(ctx, openai.FileNewParams{File: f, Purpose: openai.FilePurposeBatch})
	f.Close()
	if err != nil {
		return nil, err
	}

	batch, err := c.api.Batches.New(ctx, openai.BatchNewParams{
		InputFileID:      fileObj.ID,
		Endpoint:         openai.BatchNewParamsEndpointV1Responses,
		CompletionWindow: openai.BatchNewParamsCompletionWindow24h,
	})
	if err != nil {
		return nil, err
	}

	for {
		b, err := c.api.Batches.Get(ctx, batch.ID)
		if err != nil {
			return nil, err
		}
		if b.Status == openai.BatchStatusCompleted {
			batch = b
			break
		}
		if b.Status == openai.BatchStatusFailed || b.Status == openai.BatchStatusCancelled {
			return nil, fmt.Errorf("batch ended with status %s", b.Status)
		}
		time.Sleep(5 * time.Second)
	}

	resp, err := c.api.Files.Content(ctx, batch.OutputFileID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	results := make([]BatchResult, len(batches))
	for scanner.Scan() {
		var line struct {
			CustomID string             `json:"custom_id"`
			Response responses.Response `json:"response"`
		}
		data := scanner.Bytes()
		if err := json.Unmarshal(data, &line); err != nil {
			return nil, err
		}
		idx, err := strconv.Atoi(strings.TrimPrefix(line.CustomID, "req_"))
		if err != nil || idx < 0 || idx >= len(results) {
			continue
		}
		out := line.Response.OutputText()
		if out == "" && len(line.Response.Output) > 0 && len(line.Response.Output[0].Content) > 0 {
			out = line.Response.Output[0].Content[0].Text
		}
		results[idx] = BatchResult{Request: reqs[idx], Response: string(data), Output: out}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
