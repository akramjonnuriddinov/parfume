package assistant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const openAIURL = "https://api.openai.com/v1/chat/completions"

// Minimal Chat Completions types — only the fields the assistant uses.

type chatMessage struct {
	Role       string     `json:"role"`
	Content    *string    `json:"content"`
	ToolCalls  []toolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type toolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type toolDef struct {
	Type     string      `json:"type"`
	Function functionDef `json:"function"`
}

type functionDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

type completionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Tools       []toolDef     `json:"tools,omitempty"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_completion_tokens,omitempty"`
}

type completionResponse struct {
	Choices []struct {
		Message      chatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type openAIClient struct {
	apiKey string
	model  string
	http   *http.Client
}

func (c *openAIClient) complete(ctx context.Context, req completionRequest) (chatMessage, error) {
	req.Model = c.model
	body, err := json.Marshal(req)
	if err != nil {
		return chatMessage{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIURL, bytes.NewReader(body))
	if err != nil {
		return chatMessage{}, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return chatMessage{}, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return chatMessage{}, err
	}

	var out completionResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return chatMessage{}, fmt.Errorf("openai: status %d: invalid response", resp.StatusCode)
	}
	if out.Error != nil {
		return chatMessage{}, fmt.Errorf("openai: status %d: %s", resp.StatusCode, out.Error.Message)
	}
	if resp.StatusCode != http.StatusOK || len(out.Choices) == 0 {
		return chatMessage{}, fmt.Errorf("openai: status %d: empty response", resp.StatusCode)
	}
	return out.Choices[0].Message, nil
}

func newOpenAIClient(apiKey, model string) *openAIClient {
	return &openAIClient{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 60 * time.Second},
	}
}
