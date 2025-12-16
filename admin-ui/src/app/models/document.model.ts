export interface Document {
  id: string;
  title: string;
  content: string;
  source: string;
  uploaded_at: string;
  updated_at: string;
  is_active: boolean;
  metadata: string;
}

export interface RAGQuery {
  query: string;
  top_k: number;
  threshold: number;
}

export interface RAGResponse {
  answer: string;
  relevant_chunks: DocumentChunk[];
  confidence_score: number;
  processing_time_ms: number;
}

export interface DocumentChunk {
  id: string;
  document_id: string;
  chunk_index: number;
  content: string;
  embedding: number[];
  created_at: string;
}
