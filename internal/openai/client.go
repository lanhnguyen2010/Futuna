package openai

import "strings"

const roleSystem = `Bạn là một chuyên gia phân tích tài chính. Mọi phản hồi phải ở dạng JSON thuần bằng tiếng Việt. Luôn sử dụng công cụ web_search để tìm thông tin trước khi trả lời.`

const userPromptTemplate = `Phân tích các mã cổ phiếu sau: {{TICKERS}}.
Cho mỗi mã, cung cấp:
- short_term: {"recommendation", "confidence", "reason"}
- long_term: {"recommendation", "confidence", "reason"}
- strategies: ít nhất 5 mục [{"name", "stance", "note"}]
Sau khi hoàn thành từng mã, đưa ra overall_recommendation: {"recommendation", "confidence", "justification"}.
Cuối cùng, liệt kê sources: danh sách URL đã tham khảo.`

func buildUserPrompt(tickers []string) string {
	return strings.ReplaceAll(userPromptTemplate, "{{TICKERS}}", strings.Join(tickers, ", "))
}

// BuildMessages tạo danh sách thông điệp cho API OpenAI.
func BuildMessages(tickers []string) []Message {
	return []Message{
		{Role: "system", Content: roleSystem},
		{Role: "user", Content: buildUserPrompt(tickers)},
	}
}

// Message đại diện cho một thông điệp gửi tới API OpenAI.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
