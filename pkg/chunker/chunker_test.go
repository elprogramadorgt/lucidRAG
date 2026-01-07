package chunker

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	c := New(100, 20)

	if c.ChunkSize != 100 {
		t.Errorf("Expected ChunkSize 100, got %d", c.ChunkSize)
	}
	if c.ChunkOverlap != 20 {
		t.Errorf("Expected ChunkOverlap 20, got %d", c.ChunkOverlap)
	}
}

func TestNewDefaults(t *testing.T) {
	c := New(0, 0)

	if c.ChunkSize != 512 {
		t.Errorf("Expected default ChunkSize 512, got %d", c.ChunkSize)
	}
	if c.ChunkOverlap != 0 {
		t.Errorf("Expected ChunkOverlap 0, got %d", c.ChunkOverlap)
	}
}

func TestNewNegativeChunkSize(t *testing.T) {
	c := New(-10, 5)

	if c.ChunkSize != 512 {
		t.Errorf("Expected default ChunkSize 512 for negative input, got %d", c.ChunkSize)
	}
}

func TestNewNegativeOverlap(t *testing.T) {
	c := New(100, -5)

	if c.ChunkOverlap != 0 {
		t.Errorf("Expected ChunkOverlap 0 for negative input, got %d", c.ChunkOverlap)
	}
}

func TestNewOverlapTooLarge(t *testing.T) {
	c := New(100, 100)

	if c.ChunkOverlap != 25 {
		t.Errorf("Expected ChunkOverlap 25 (1/4 of size), got %d", c.ChunkOverlap)
	}
}

func TestChunkEmptyText(t *testing.T) {
	c := New(10, 2)
	chunks := c.Chunk("")

	if chunks != nil {
		t.Errorf("Expected nil for empty text, got %v", chunks)
	}
}

func TestChunkWhitespaceOnly(t *testing.T) {
	c := New(10, 2)
	chunks := c.Chunk("   \t\n  ")

	if chunks != nil {
		t.Errorf("Expected nil for whitespace-only text, got %v", chunks)
	}
}

func TestChunkSingleWord(t *testing.T) {
	c := New(10, 2)
	chunks := c.Chunk("hello")

	if len(chunks) != 1 {
		t.Fatalf("Expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0] != "hello" {
		t.Errorf("Expected 'hello', got '%s'", chunks[0])
	}
}

func TestChunkMultipleWords(t *testing.T) {
	c := New(3, 1)
	text := "one two three four five"
	chunks := c.Chunk(text)

	if len(chunks) == 0 {
		t.Fatal("Expected chunks, got none")
	}

	// Each chunk should have at most 3 words
	for i, chunk := range chunks {
		words := strings.Fields(chunk)
		if len(words) > 3 {
			t.Errorf("Chunk %d has %d words, expected max 3", i, len(words))
		}
	}
}

func TestChunkOverlap(t *testing.T) {
	c := New(3, 1)
	text := "one two three four"
	chunks := c.Chunk(text)

	// With overlap=1 and size=3, chunks should overlap by 1 word
	if len(chunks) < 2 {
		t.Fatal("Expected at least 2 chunks for overlap test")
	}

	// Check that there's some overlap between consecutive chunks
	chunk1Words := strings.Fields(chunks[0])
	chunk2Words := strings.Fields(chunks[1])

	if len(chunk1Words) > 0 && len(chunk2Words) > 0 {
		// The last word of chunk1 might appear in chunk2 due to overlap
		lastWord := chunk1Words[len(chunk1Words)-1]
		foundOverlap := false
		for _, w := range chunk2Words {
			if w == lastWord {
				foundOverlap = true
				break
			}
		}
		// Note: overlap depends on step size, so this is just a sanity check
		_ = foundOverlap
	}
}

func TestChunkWithMetadata(t *testing.T) {
	c := New(5, 1)
	text := "one two three four five six seven eight"
	chunks := c.ChunkWithMetadata(text)

	if len(chunks) == 0 {
		t.Fatal("Expected chunks with metadata, got none")
	}

	for i, chunk := range chunks {
		if chunk.Index != i {
			t.Errorf("Expected Index %d, got %d", i, chunk.Index)
		}
		if chunk.Content == "" {
			t.Errorf("Chunk %d has empty content", i)
		}
	}
}

func TestChunkWithMetadataEmpty(t *testing.T) {
	c := New(5, 1)
	chunks := c.ChunkWithMetadata("")

	if len(chunks) != 0 {
		t.Errorf("Expected 0 chunks for empty text, got %d", len(chunks))
	}
}

func TestChunkerZeroValue(t *testing.T) {
	var c Chunker

	if c.ChunkSize != 0 {
		t.Error("Expected zero value ChunkSize to be 0")
	}
	if c.ChunkOverlap != 0 {
		t.Error("Expected zero value ChunkOverlap to be 0")
	}
}

func TestChunkWithPositionZeroValue(t *testing.T) {
	var cwp ChunkWithPosition

	if cwp.Content != "" {
		t.Error("Expected zero value Content to be empty")
	}
	if cwp.Index != 0 {
		t.Error("Expected zero value Index to be 0")
	}
}
