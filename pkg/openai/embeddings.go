package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type embeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

type apiError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func (c *Client) CreateEmbedding(ctx context.Context, text string, model string) ([]float64, error) {
	if model == "" {
		model = "text-embedding-ada-002"
	}

	reqBody := embeddingRequest{
		Model: model,
		Input: text,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/embeddings", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error.Message != "" {
			return nil, fmt.Errorf("OpenAI API error: %s (type: %s)", apiErr.Error.Message, apiErr.Error.Type)
		}
		return nil, fmt.Errorf("OpenAI API error: status %d", resp.StatusCode)
	}

	var embResp embeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(embResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return embResp.Data[0].Embedding, nil
}

func (c *Client) CreateEmbeddings(ctx context.Context, texts []string, model string) ([][]float64, error) {
	embeddings := make([][]float64, len(texts))
	for i, text := range texts {
		emb, err := c.CreateEmbedding(ctx, text, model)
		if err != nil {
			return nil, fmt.Errorf("failed to create embedding for text %d: %w", i, err)
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}
