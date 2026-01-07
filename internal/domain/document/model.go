package document

import "time"

type Document struct {
	ID         string    `json:"id" bson:"_id,omitempty"`
	UserID     string    `json:"user_id" bson:"user_id"`
	Title      string    `json:"title" bson:"title"`
	Content    string    `json:"content" bson:"content"`
	Source     string    `json:"source" bson:"source"`
	UploadedAt time.Time `json:"uploaded_at" bson:"uploaded_at"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
	IsActive   bool      `json:"is_active" bson:"is_active"`
	Metadata   string    `json:"metadata" bson:"metadata"`
}

type Chunk struct {
	ID          string    `json:"id" bson:"_id,omitempty"`
	DocumentID  string    `json:"document_id" bson:"document_id"`
	ChunkIndex  int       `json:"chunk_index" bson:"chunk_index"`
	Content     string    `json:"content" bson:"content"`
	Embedding   []float64 `json:"embedding" bson:"embedding"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
}

type RAGQuery struct {
	Query     string  `json:"query"`
	TopK      int     `json:"top_k"`
	Threshold float64 `json:"threshold"`
}

type RAGResponse struct {
	Answer           string  `json:"answer"`
	RelevantChunks   []Chunk `json:"relevant_chunks"`
	ConfidenceScore  float64 `json:"confidence_score"`
	ProcessingTimeMs int64   `json:"processing_time_ms"`
}
