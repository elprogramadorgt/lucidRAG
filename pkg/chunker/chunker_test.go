package chunker

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name            string
		chunkSize       int
		chunkOverlap    int
		wantChunkSize   int
		wantChunkOverlap int
	}{
		{
			name:             "valid values",
			chunkSize:        100,
			chunkOverlap:     20,
			wantChunkSize:    100,
			wantChunkOverlap: 20,
		},
		{
			name:             "zero chunk size defaults to 512",
			chunkSize:        0,
			chunkOverlap:     20,
			wantChunkSize:    512,
			wantChunkOverlap: 20,
		},
		{
			name:             "negative chunk size defaults to 512",
			chunkSize:        -10,
			chunkOverlap:     20,
			wantChunkSize:    512,
			wantChunkOverlap: 20,
		},
		{
			name:             "negative overlap defaults to 0",
			chunkSize:        100,
			chunkOverlap:     -5,
			wantChunkSize:    100,
			wantChunkOverlap: 0,
		},
		{
			name:             "overlap >= chunk size defaults to chunk size / 4",
			chunkSize:        100,
			chunkOverlap:     100,
			wantChunkSize:    100,
			wantChunkOverlap: 25,
		},
		{
			name:             "overlap > chunk size defaults to chunk size / 4",
			chunkSize:        100,
			chunkOverlap:     150,
			wantChunkSize:    100,
			wantChunkOverlap: 25,
		},
		{
			name:             "zero overlap is valid",
			chunkSize:        100,
			chunkOverlap:     0,
			wantChunkSize:    100,
			wantChunkOverlap: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.chunkSize, tt.chunkOverlap)
			if c.ChunkSize != tt.wantChunkSize {
				t.Errorf("ChunkSize = %d, want %d", c.ChunkSize, tt.wantChunkSize)
			}
			if c.ChunkOverlap != tt.wantChunkOverlap {
				t.Errorf("ChunkOverlap = %d, want %d", c.ChunkOverlap, tt.wantChunkOverlap)
			}
		})
	}
}

func TestChunk(t *testing.T) {
	tests := []struct {
		name         string
		chunkSize    int
		chunkOverlap int
		text         string
		wantLen      int
		wantNil      bool
		checkFirst   string
	}{
		{
			name:         "empty string returns nil",
			chunkSize:    10,
			chunkOverlap: 2,
			text:         "",
			wantNil:      true,
		},
		{
			name:         "whitespace only returns nil",
			chunkSize:    10,
			chunkOverlap: 2,
			text:         "   \t\n  ",
			wantNil:      true,
		},
		{
			name:         "short text returns single chunk",
			chunkSize:    10,
			chunkOverlap: 2,
			text:         "hello world",
			wantLen:      1,
			checkFirst:   "hello world",
		},
		{
			name:         "text splits into multiple chunks",
			chunkSize:    3,
			chunkOverlap: 1,
			text:         "one two three four five",
			wantLen:      2,
			checkFirst:   "one two three",
		},
		{
			name:         "no overlap",
			chunkSize:    2,
			chunkOverlap: 0,
			text:         "a b c d e f",
			wantLen:      3,
			checkFirst:   "a b",
		},
		{
			name:         "single word",
			chunkSize:    5,
			chunkOverlap: 1,
			text:         "hello",
			wantLen:      1,
			checkFirst:   "hello",
		},
		{
			name:         "preserves multiple spaces as single separator",
			chunkSize:    10,
			chunkOverlap: 0,
			text:         "hello    world",
			wantLen:      1,
			checkFirst:   "hello world",
		},
		{
			name:         "handles tabs and newlines",
			chunkSize:    10,
			chunkOverlap: 0,
			text:         "hello\tworld\ntest",
			wantLen:      1,
			checkFirst:   "hello world test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.chunkSize, tt.chunkOverlap)
			result := c.Chunk(tt.text)

			if tt.wantNil {
				if result != nil {
					t.Errorf("Chunk() = %v, want nil", result)
				}
				return
			}

			if len(result) != tt.wantLen {
				t.Errorf("len(Chunk()) = %d, want %d, result=%v", len(result), tt.wantLen, result)
			}

			if tt.checkFirst != "" && len(result) > 0 {
				if result[0] != tt.checkFirst {
					t.Errorf("Chunk()[0] = %q, want %q", result[0], tt.checkFirst)
				}
			}
		})
	}
}

func TestChunkOverlap(t *testing.T) {
	c := New(3, 1)
	text := "one two three four five six"
	chunks := c.Chunk(text)

	// With chunk size 3 and overlap 1, step = 2
	// Chunk 0: words 0-2 = "one two three"
	// Chunk 1: words 2-4 = "three four five"
	// Chunk 2: words 4-6 = "five six"

	if len(chunks) < 2 {
		t.Fatalf("Expected at least 2 chunks, got %d", len(chunks))
	}

	// Check that overlap works - "three" should appear in first and second chunks
	if !strings.Contains(chunks[0], "three") {
		t.Error("First chunk should contain 'three'")
	}
	if !strings.Contains(chunks[1], "three") {
		t.Error("Second chunk should contain 'three' (overlap)")
	}
}

func TestChunkWithMetadata(t *testing.T) {
	tests := []struct {
		name         string
		chunkSize    int
		chunkOverlap int
		text         string
		wantLen      int
	}{
		{
			name:         "empty text",
			chunkSize:    10,
			chunkOverlap: 0,
			text:         "",
			wantLen:      0,
		},
		{
			name:         "single chunk",
			chunkSize:    10,
			chunkOverlap: 0,
			text:         "hello world",
			wantLen:      1,
		},
		{
			name:         "multiple chunks",
			chunkSize:    2,
			chunkOverlap: 0,
			text:         "a b c d e f",
			wantLen:      3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.chunkSize, tt.chunkOverlap)
			result := c.ChunkWithMetadata(tt.text)

			if len(result) != tt.wantLen {
				t.Errorf("len(ChunkWithMetadata()) = %d, want %d", len(result), tt.wantLen)
			}

			// Verify indices are sequential
			for i, chunk := range result {
				if chunk.Index != i {
					t.Errorf("ChunkWithMetadata()[%d].Index = %d, want %d", i, chunk.Index, i)
				}
				if chunk.Content == "" {
					t.Errorf("ChunkWithMetadata()[%d].Content is empty", i)
				}
			}
		})
	}
}

func TestChunkWithPositionFields(t *testing.T) {
	c := New(5, 0)
	text := "one two three"
	result := c.ChunkWithMetadata(text)

	if len(result) != 1 {
		t.Fatalf("Expected 1 chunk, got %d", len(result))
	}

	chunk := result[0]
	if chunk.Content != "one two three" {
		t.Errorf("Content = %q, want %q", chunk.Content, "one two three")
	}
	if chunk.Index != 0 {
		t.Errorf("Index = %d, want 0", chunk.Index)
	}
}

func TestTokenize(t *testing.T) {
	// Test tokenization behavior through Chunk
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "simple words",
			text:     "hello world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "multiple spaces",
			text:     "hello   world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "tabs",
			text:     "hello\tworld",
			expected: []string{"hello", "world"},
		},
		{
			name:     "newlines",
			text:     "hello\nworld",
			expected: []string{"hello", "world"},
		},
		{
			name:     "mixed whitespace",
			text:     "  hello \t world \n test  ",
			expected: []string{"hello", "world", "test"},
		},
		{
			name:     "unicode characters",
			text:     "héllo wörld 你好",
			expected: []string{"héllo", "wörld", "你好"},
		},
		{
			name:     "punctuation attached",
			text:     "hello, world!",
			expected: []string{"hello,", "world!"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use large chunk size to get all words in one chunk
			c := New(1000, 0)
			chunks := c.Chunk(tt.text)

			if len(chunks) == 0 {
				t.Fatal("Expected at least one chunk")
			}

			result := chunks[0]
			expected := strings.Join(tt.expected, " ")
			if result != expected {
				t.Errorf("Chunk() = %q, want %q", result, expected)
			}
		})
	}
}

func TestChunkEdgeCases(t *testing.T) {
	t.Run("very small step size", func(t *testing.T) {
		c := New(2, 1)
		text := "a b c d e"
		chunks := c.Chunk(text)

		// Step = chunkSize - overlap = 2 - 1 = 1
		// Should create multiple overlapping chunks
		if len(chunks) == 0 {
			t.Error("Expected chunks")
		}
	})

	t.Run("chunk size equals text words", func(t *testing.T) {
		c := New(3, 0)
		text := "one two three"
		chunks := c.Chunk(text)

		if len(chunks) != 1 {
			t.Errorf("Expected 1 chunk, got %d", len(chunks))
		}
		if chunks[0] != "one two three" {
			t.Errorf("Chunk = %q, want %q", chunks[0], "one two three")
		}
	})

	t.Run("chunk size larger than text", func(t *testing.T) {
		c := New(100, 10)
		text := "hello world"
		chunks := c.Chunk(text)

		if len(chunks) != 1 {
			t.Errorf("Expected 1 chunk, got %d", len(chunks))
		}
	})
}

func BenchmarkChunk(b *testing.B) {
	text := strings.Repeat("word ", 1000)
	c := New(100, 20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Chunk(text)
	}
}

func BenchmarkChunkWithMetadata(b *testing.B) {
	text := strings.Repeat("word ", 1000)
	c := New(100, 20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.ChunkWithMetadata(text)
	}
}
