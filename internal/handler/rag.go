package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/elprogramadorgt/lucidRAG/internal/domain"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
)

// RAGHandler handles RAG-related requests
type RAGHandler struct {
	ragSvc domain.RAGService
	logger *logger.Logger
}

// NewRAGHandler creates a new RAG handler
func NewRAGHandler(ragSvc domain.RAGService, log *logger.Logger) *RAGHandler {
	return &RAGHandler{
		ragSvc: ragSvc,
		logger: log,
	}
}

// Query handles RAG query requests
func (h *RAGHandler) Query(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var query domain.RAGQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		h.logger.Error("Failed to decode query: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Set defaults if not provided
	if query.TopK == 0 {
		query.TopK = 5
	}
	if query.Threshold == 0 {
		query.Threshold = 0.7
	}

	response, err := h.ragSvc.Query(r.Context(), query)
	if err != nil {
		h.logger.Error("Failed to process query: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AddDocument handles document upload requests
func (h *RAGHandler) AddDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var doc domain.Document
	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		h.logger.Error("Failed to decode document: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := h.ragSvc.AddDocument(r.Context(), &doc); err != nil {
		h.logger.Error("Failed to add document: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id":      doc.ID,
		"message": "Document added successfully",
	})
}

// GetDocument handles document retrieval requests
func (h *RAGHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	docID := r.URL.Query().Get("id")
	if docID == "" {
		http.Error(w, "Document ID is required", http.StatusBadRequest)
		return
	}

	doc, err := h.ragSvc.GetDocument(r.Context(), docID)
	if err != nil {
		h.logger.Error("Failed to get document: %v", err)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

// ListDocuments handles document listing requests
func (h *RAGHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 10
	}

	docs, err := h.ragSvc.ListDocuments(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list documents: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"documents": docs,
		"limit":     limit,
		"offset":    offset,
	})
}

// UpdateDocument handles document update requests
func (h *RAGHandler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var doc domain.Document
	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		h.logger.Error("Failed to decode document: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := h.ragSvc.UpdateDocument(r.Context(), &doc); err != nil {
		h.logger.Error("Failed to update document: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Document updated successfully"})
}

// DeleteDocument handles document deletion requests
func (h *RAGHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	docID := r.URL.Query().Get("id")
	if docID == "" {
		http.Error(w, "Document ID is required", http.StatusBadRequest)
		return
	}

	if err := h.ragSvc.DeleteDocument(r.Context(), docID); err != nil {
		h.logger.Error("Failed to delete document: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Document deleted successfully"})
}
