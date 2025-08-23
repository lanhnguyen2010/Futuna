package openai

import (
	"context"
	"encoding/json"
	"log"
	"strings"

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
				"maxItems": 5,
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

	// Build the minimal typed params (only what we must type)
	params := responses.ResponseNewParams{
		Model: c.model, // string, not openai.String(...)
	}

	// Keep a copy of the request (with our extra JSON injected below)
	reqJSON, _ := json.Marshal(params)

	// Call with JSON overrides:
	// - tools: enable web_search (preview)
	// - structured outputs via JSON schema
	// - input: system + user messages
	resp, err := c.api.Responses.New(
		ctx,
		params,
		// Enable web search tool (try "web_search" first; some regions still use "..._preview")
		option.WithJSONSet("tools", []map[string]any{
			{"type": "web_search"},
		}),
		option.WithJSONSet("cache_control", map[string]any{"type": "ephemeral"}),
		// Structured outputs under text.format (JSON Schema)
		option.WithJSONSet("text.format", map[string]any{
			"type":   "json_schema",
			"schema": schema,
			"name":   "analysis_response",
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
