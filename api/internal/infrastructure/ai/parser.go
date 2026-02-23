package ai

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/generative-ai-go/genai"
)

func extractText(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 {
		return ""
	}
	var parts []string
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			parts = append(parts, string(text))
		}
	}
	return strings.Join(parts, "")
}

var jsonFenceRegex = regexp.MustCompile("(?s)```(?:json)?\\s*(.+?)```")

func parseJSONFromGemini[T any](resp *genai.GenerateContentResponse) (*T, error) {
	text := extractText(resp)

	if matches := jsonFenceRegex.FindStringSubmatch(text); len(matches) > 1 {
		text = matches[1]
	}

	text = strings.TrimSpace(text)

	var result T
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("parsing gemini json response: %w\nraw: %s", err, text)
	}

	return &result, nil
}
