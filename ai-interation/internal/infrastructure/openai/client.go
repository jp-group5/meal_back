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
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

type responseAPIResponse struct {
	OutputText string `json:"output_text"`
	Output     []struct {
		Type    string `json:"type"`
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

func (c *Client) AnalyzeMeal(ctx context.Context, imageBytes []byte) (string, error) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"date": map[string]any{"type": "string"},
			"content": map[string]any{
				"type": "array",
				"items": map[string]any{"type": "string"},
			},
			"calories": map[string]any{"type": "number"},
			"protein":  map[string]any{"type": "number"},
			"fat":      map[string]any{"type": "number"},
			"carbs":    map[string]any{"type": "number"},
		},
		"required": []string{"date", "content", "calories", "protein", "fat", "carbs"},
		"additionalProperties": false,
	}

	body := map[string]any{
		"model": c.model,
		"input": []any{
			map[string]any{
				"role": "user",
				"content": []any{
					map[string]any{
						"type": "input_text",
						"text": "食事画像を解析し、献立名と栄養成分をJSONだけで返してください。",
					},
					map[string]any{
						"type":      "input_image",
						"image_url": "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(imageBytes),
						"detail":    "high",
					},
				},
			},
		},
		"text": map[string]any{
			"format": map[string]any{
				"type":   "json_schema",
				"name":   "meal_analysis",
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
						"calories":  map[string]any{"type": "number"},
						"protein":   map[string]any{"type": "number"},
						"fat":       map[string]any{"type": "number"},
						"carbs":     map[string]any{"type": "number"},
						"reason":    map[string]any{"type": "string"},
					},
					"required": []string{"menu_name", "calories", "protein", "fat", "carbs", "reason"},
					"additionalProperties": false,
				},
			},
			"error": map[string]any{"type": "null"},
		},
		"required": []string{"recommendations", "error"},
		"additionalProperties": false,
	}

	body := map[string]any{
		"model": c.model,
		"input": []any{
			map[string]any{
				"role": "user",
				"content": []any{
					map[string]any{
						"type": "input_text",
						"text": prompt,
					},
				},
			},
		},
		"text": map[string]any{
			"format": map[string]any{
				"type":   "json_schema",
				"name":   "recommendation",
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

	raw := extractOutputText(out)
	if strings.TrimSpace(raw) == "" {
		return "", errors.New("openai returned empty output text")
	}

	fmt.Printf("[OpenAI usage] input=%d output=%d total=%d\n", out.Usage.InputTokens, out.Usage.OutputTokens, out.Usage.TotalTokens)

	return raw, nil
}

func extractOutputText(out responseAPIResponse) string {
	if strings.TrimSpace(out.OutputText) != "" {
		return out.OutputText
	}

	var parts []string
	for _, item := range out.Output {
		for _, c := range item.Content {
			if strings.TrimSpace(c.Text) != "" {
				parts = append(parts, c.Text)
			}
		}
	}
	return strings.Join(parts, "")
}

var _ port.AIClient = (*Client)(nil)