package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"ai-interation/internal/port"
)

type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewClientFromEnv() (*Client, error) {
	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY is empty")
	}

	model := strings.TrimSpace(os.Getenv("OPENAI_MODEL"))
	if model == "" {
		model = "gpt-4.1-mini"
	}

	return &Client{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

type responseRequest struct {
	Model string `json:"model"`
	Input any    `json:"input"`
	Text  any    `json:"text,omitempty"`
}

type responseAPIResponse struct {
	OutputText string `json:"output_text"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

func (c *Client) AnalyzeMeal(ctx context.Context, imageBytes []byte) (string, error) {
	dataURL := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(imageBytes)

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"items": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{"type": "string"},
					},
					"required": []string{"name"},
					"additionalProperties": false,
				},
			},
			"total_nutrition": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"calories":          map[string]any{"type": "number"},
					"protein":           map[string]any{"type": "number"},
					"fat":               map[string]any{"type": "number"},
					"carbohydrates":     map[string]any{"type": "number"},
					"vegetables_amount": map[string]any{"type": "string"},
				},
				"required": []string{"calories", "protein", "fat", "carbohydrates", "vegetables_amount"},
				"additionalProperties": false,
			},
			"error": map[string]any{"type": "null"},
		},
		"required": []string{"items", "total_nutrition", "error"},
		"additionalProperties": false,
	}

	body := responseRequest{
		Model: c.model,
		Input: []any{
			map[string]any{
				"role": "user",
				"content": []any{
					map[string]any{
						"type": "input_text",
						"text": "この画像の食事を解析し、献立名と栄養成分をJSONで返してください。",
					},
					map[string]any{
						"type":      "input_image",
						"image_url":  dataURL,
						"detail":     "high",
					},
				},
			},
		},
		Text: map[string]any{
			"format": map[string]any{
				"type": "json_schema",
				"name": "meal_analysis",
				"schema": schema,
				"strict": true,
			},
		},
	}

	return c.callResponsesAPI(ctx, body)
}

func (c *Client) GenerateRecommendation(ctx context.Context, prompt string) (string, error) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"recommendations": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"menu_name": map[string]any{"type": "string"},
						"reason":    map[string]any{"type": "string"},
					},
					"required": []string{"menu_name", "reason"},
					"additionalProperties": false,
				},
			},
			"error": map[string]any{"type": "null"},
		},
		"required": []string{"recommendations", "error"},
		"additionalProperties": false,
	}

	body := responseRequest{
		Model: c.model,
		Input: prompt,
		Text: map[string]any{
			"format": map[string]any{
				"type": "json_schema",
				"name": "recommendation",
				"schema": schema,
				"strict": true,
			},
		},
	}

	return c.callResponsesAPI(ctx, body)
}

func (c *Client) callResponsesAPI(ctx context.Context, payload any) (string, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/responses", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(resp.Body)
		return "", fmt.Errorf("openai api status: %s, body: %s", resp.Status, buf.String())
	}

	var out responseAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}

	fmt.Printf(
		"[OpenAI usage] input=%d output=%d total=%d\n",
		out.Usage.InputTokens,
		out.Usage.OutputTokens,
		out.Usage.TotalTokens,
	)

	return out.OutputText, nil
}

var _ port.AIClient = (*Client)(nil)