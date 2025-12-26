package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-api-key")
	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	if client.apiKey != "test-api-key" {
		t.Errorf("Expected apiKey 'test-api-key', got '%s'", client.apiKey)
	}
	if client.baseURL != "https://api.openai.com/v1" {
		t.Errorf("Expected default baseURL, got '%s'", client.baseURL)
	}
}

func TestCreateEmbedding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" {
			t.Errorf("Expected path /embeddings, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("Expected Authorization header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Expected Content-Type application/json")
		}

		response := embeddingResponse{
			Object: "list",
			Data: []struct {
				Object    string    `json:"object"`
				Index     int       `json:"index"`
				Embedding []float64 `json:"embedding"`
			}{
				{
					Object:    "embedding",
					Index:     0,
					Embedding: []float64{0.1, 0.2, 0.3},
				},
			},
			Model: "text-embedding-ada-002",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-key",
		baseURL:    server.URL,
		httpClient: http.DefaultClient,
	}

	embedding, err := client.CreateEmbedding(context.Background(), "test text", "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(embedding) != 3 {
		t.Errorf("Expected 3 embedding values, got %d", len(embedding))
	}
	if embedding[0] != 0.1 {
		t.Errorf("Expected first value 0.1, got %f", embedding[0])
	}
}

func TestCreateEmbeddingDefaultModel(t *testing.T) {
	var capturedModel string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req embeddingRequest
		json.NewDecoder(r.Body).Decode(&req)
		capturedModel = req.Model

		response := embeddingResponse{
			Data: []struct {
				Object    string    `json:"object"`
				Index     int       `json:"index"`
				Embedding []float64 `json:"embedding"`
			}{
				{Embedding: []float64{0.1}},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-key",
		baseURL:    server.URL,
		httpClient: http.DefaultClient,
	}

	_, err := client.CreateEmbedding(context.Background(), "test", "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if capturedModel != "text-embedding-ada-002" {
		t.Errorf("Expected default model 'text-embedding-ada-002', got '%s'", capturedModel)
	}
}

func TestCreateEmbeddingAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		response := apiError{
			Error: struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    string `json:"code"`
			}{
				Message: "Invalid API key",
				Type:    "invalid_request_error",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "invalid-key",
		baseURL:    server.URL,
		httpClient: http.DefaultClient,
	}

	_, err := client.CreateEmbedding(context.Background(), "test", "")
	if err == nil {
		t.Fatal("Expected error for API error response")
	}
}

func TestCreateEmbeddingNoData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := embeddingResponse{
			Data: []struct {
				Object    string    `json:"object"`
				Index     int       `json:"index"`
				Embedding []float64 `json:"embedding"`
			}{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-key",
		baseURL:    server.URL,
		httpClient: http.DefaultClient,
	}

	_, err := client.CreateEmbedding(context.Background(), "test", "")
	if err == nil {
		t.Fatal("Expected error for empty embedding response")
	}
}

func TestCreateEmbeddings(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := embeddingResponse{
			Data: []struct {
				Object    string    `json:"object"`
				Index     int       `json:"index"`
				Embedding []float64 `json:"embedding"`
			}{
				{Embedding: []float64{float64(callCount) * 0.1}},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-key",
		baseURL:    server.URL,
		httpClient: http.DefaultClient,
	}

	texts := []string{"text1", "text2", "text3"}
	embeddings, err := client.CreateEmbeddings(context.Background(), texts, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(embeddings) != 3 {
		t.Errorf("Expected 3 embeddings, got %d", len(embeddings))
	}
	if callCount != 3 {
		t.Errorf("Expected 3 API calls, got %d", callCount)
	}
}

func TestCreateChatCompletion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("Expected path /chat/completions, got %s", r.URL.Path)
		}

		response := chatCompletionResponse{
			ID:      "chatcmpl-123",
			Object:  "chat.completion",
			Created: 1677652288,
			Model:   "gpt-3.5-turbo",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "Hello! How can I help you?",
					},
					FinishReason: "stop",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-key",
		baseURL:    server.URL,
		httpClient: http.DefaultClient,
	}

	messages := []ChatMessage{
		{Role: "user", Content: "Hello"},
	}

	result, err := client.CreateChatCompletion(context.Background(), messages, "", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != "Hello! How can I help you?" {
		t.Errorf("Expected response content, got '%s'", result)
	}
}

func TestCreateChatCompletionDefaultModel(t *testing.T) {
	var capturedModel string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatCompletionRequest
		json.NewDecoder(r.Body).Decode(&req)
		capturedModel = req.Model

		response := chatCompletionResponse{
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{Content: "response"}},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-key",
		baseURL:    server.URL,
		httpClient: http.DefaultClient,
	}

	messages := []ChatMessage{{Role: "user", Content: "test"}}
	_, err := client.CreateChatCompletion(context.Background(), messages, "", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if capturedModel != "gpt-3.5-turbo" {
		t.Errorf("Expected default model 'gpt-3.5-turbo', got '%s'", capturedModel)
	}
}

func TestCreateChatCompletionWithOptions(t *testing.T) {
	var capturedTemp float64
	var capturedMaxTokens int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatCompletionRequest
		json.NewDecoder(r.Body).Decode(&req)
		capturedTemp = req.Temperature
		capturedMaxTokens = req.MaxTokens

		response := chatCompletionResponse{
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{Content: "response"}},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-key",
		baseURL:    server.URL,
		httpClient: http.DefaultClient,
	}

	messages := []ChatMessage{{Role: "user", Content: "test"}}
	opts := &CompletionOptions{
		Temperature: 0.7,
		MaxTokens:   100,
	}
	_, err := client.CreateChatCompletion(context.Background(), messages, "gpt-4", opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if capturedTemp != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", capturedTemp)
	}
	if capturedMaxTokens != 100 {
		t.Errorf("Expected maxTokens 100, got %d", capturedMaxTokens)
	}
}

func TestCreateChatCompletionNoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := chatCompletionResponse{
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-key",
		baseURL:    server.URL,
		httpClient: http.DefaultClient,
	}

	messages := []ChatMessage{{Role: "user", Content: "test"}}
	_, err := client.CreateChatCompletion(context.Background(), messages, "", nil)
	if err == nil {
		t.Fatal("Expected error for empty choices")
	}
}

func TestChatMessageStruct(t *testing.T) {
	msg := ChatMessage{
		Role:    "user",
		Content: "Hello, world!",
	}

	if msg.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", msg.Role)
	}
	if msg.Content != "Hello, world!" {
		t.Errorf("Expected content 'Hello, world!', got '%s'", msg.Content)
	}
}

func TestCompletionOptionsStruct(t *testing.T) {
	opts := CompletionOptions{
		Temperature: 0.5,
		MaxTokens:   500,
	}

	if opts.Temperature != 0.5 {
		t.Errorf("Expected temperature 0.5, got %f", opts.Temperature)
	}
	if opts.MaxTokens != 500 {
		t.Errorf("Expected maxTokens 500, got %d", opts.MaxTokens)
	}
}
