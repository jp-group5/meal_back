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
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (c *Client) AnalyzeMeal(ctx context.Context, imageBytes []byte) (string, error) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"contents": map[string]any{
				"type": "array",
				"items": map[string]any{"type": "string"},
			},
			"total_nutrition": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"calories": map[string]any{"type": "number"},
					"protein":  map[string]any{"type": "number"},
					"fat":      map[string]any{"type": "number"},
					"carbs":    map[string]any{"type": "number"},
				},
				"required": []string{"calories", "protein", "fat", "carbs"},
				"additionalProperties": false,
			},
			"error": map[string]any{"type": "null"},
		},
		"required": []string{"contents", "total_nutrition", "error"},
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
						"text": "食事画像を解析し、料理名の一覧と合計栄養成分をJSONだけで返してください。",
					},
					map[string]any{
						"type":      "input_image",
						"image_url": "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(imageBytes),
						"detail":    "high",
					},
				},
			},
		},
		"text": jsonSchemaFormat("meal_analysis", schema),
	}

	return c.call(ctx, body)
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
		"text": jsonSchemaFormat("recommendation", schema),
	}

	return c.call(ctx, body)
}

func jsonSchemaFormat(name string, schema map[string]any) map[string]any {
	return map[string]any{
		"format": map[string]any{
			"type":   "json_schema",
			"name":   name,
			"schema": schema,
			"strict": true,
		},
	}
}

type responseBody struct {
	OutputText string `json:"output_text"`
	Output     []struct {
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

func (c *Client) call(ctx context.Context, payload any) (string, error) {
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

	var out responseBody
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}

	fmt.Printf(
		"[OpenAI usage] input=%d output=%d total=%d\n",
		out.Usage.InputTokens,
		out.Usage.OutputTokens,
		out.Usage.TotalTokens,
	)

	text := extractText(out)
	if strings.TrimSpace(text) == "" {
		return "", errors.New("openai returned empty output text")
	}

	return text, nil
}

func extractText(out responseBody) string {
	if strings.TrimSpace(out.OutputText) != "" {
		return out.OutputText
	}

	var parts []string
	for _, item := range out.Output {
		for _, content := range item.Content {
			if strings.TrimSpace(content.Text) != "" {
				parts = append(parts, content.Text)
			}
		}
	}

	return strings.Join(parts, "")
}