// Package chunker provides text chunking utilities for RAG indexing.
package chunker

import (
	"strings"
	"unicode"
)

// Chunker splits text into overlapping chunks for embedding.
type Chunker struct {
	ChunkSize    int
	ChunkOverlap int
}

// New creates a new Chunker with the given size and overlap.
func New(chunkSize, chunkOverlap int) *Chunker {
	if chunkSize <= 0 {
		chunkSize = 512
	}
	if chunkOverlap < 0 {
		chunkOverlap = 0
	}
	if chunkOverlap >= chunkSize {
		chunkOverlap = chunkSize / 4
	}

	return &Chunker{
		ChunkSize:    chunkSize,
		ChunkOverlap: chunkOverlap,
	}
}

// Chunk splits text into overlapping word-based chunks.
func (c *Chunker) Chunk(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	words := tokenize(text)
	if len(words) == 0 {
		return nil
	}

	var chunks []string
	step := c.ChunkSize - c.ChunkOverlap
	if step <= 0 {
		step = 1
	}

	for i := 0; i < len(words); i += step {
		end := i + c.ChunkSize
		if end > len(words) {
			end = len(words)
		}

		chunk := strings.Join(words[i:end], " ")
		chunk = strings.TrimSpace(chunk)
		if chunk != "" {
			chunks = append(chunks, chunk)
		}

		if end == len(words) {
			break
		}
	}

	return chunks
}

func tokenize(text string) []string {
	var words []string
	var currentWord strings.Builder

	for _, r := range text {
		if unicode.IsSpace(r) {
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		} else {
			currentWord.WriteRune(r)
		}
	}

	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

// ChunkWithMetadata splits text into chunks with position information.
func (c *Chunker) ChunkWithMetadata(text string) []ChunkWithPosition {
	chunks := c.Chunk(text)
	result := make([]ChunkWithPosition, len(chunks))

	for i, chunk := range chunks {
		result[i] = ChunkWithPosition{
			Content: chunk,
			Index:   i,
		}
	}

	return result
}

// ChunkWithPosition represents a text chunk with its index position.
type ChunkWithPosition struct {
	Content string
	Index   int
}
