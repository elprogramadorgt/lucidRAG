package vectormath

import (
	"math"
	"testing"
)

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
		delta    float64
	}{
		{
			name:     "identical vectors",
			a:        []float64{1, 0, 0},
			b:        []float64{1, 0, 0},
			expected: 1.0,
			delta:    0.001,
		},
		{
			name:     "orthogonal vectors",
			a:        []float64{1, 0, 0},
			b:        []float64{0, 1, 0},
			expected: 0.0,
			delta:    0.001,
		},
		{
			name:     "opposite vectors",
			a:        []float64{1, 0, 0},
			b:        []float64{-1, 0, 0},
			expected: -1.0,
			delta:    0.001,
		},
		{
			name:     "similar vectors",
			a:        []float64{1, 1, 0},
			b:        []float64{1, 0, 0},
			expected: 0.707,
			delta:    0.01,
		},
		{
			name:     "different length vectors",
			a:        []float64{1, 0},
			b:        []float64{1, 0, 0},
			expected: 0,
			delta:    0.001,
		},
		{
			name:     "empty vectors",
			a:        []float64{},
			b:        []float64{},
			expected: 0,
			delta:    0.001,
		},
		{
			name:     "zero vector a",
			a:        []float64{0, 0, 0},
			b:        []float64{1, 0, 0},
			expected: 0,
			delta:    0.001,
		},
		{
			name:     "zero vector b",
			a:        []float64{1, 0, 0},
			b:        []float64{0, 0, 0},
			expected: 0,
			delta:    0.001,
		},
		{
			name:     "both zero vectors",
			a:        []float64{0, 0, 0},
			b:        []float64{0, 0, 0},
			expected: 0,
			delta:    0.001,
		},
		{
			name:     "negative values",
			a:        []float64{-1, -1, -1},
			b:        []float64{-1, -1, -1},
			expected: 1.0,
			delta:    0.001,
		},
		{
			name:     "mixed positive negative",
			a:        []float64{1, -1, 1},
			b:        []float64{-1, 1, -1},
			expected: -1.0,
			delta:    0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CosineSimilarity(tt.a, tt.b)
			if math.Abs(result-tt.expected) > tt.delta {
				t.Errorf("CosineSimilarity(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestEuclideanDistance(t *testing.T) {
	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
		delta    float64
	}{
		{
			name:     "identical vectors",
			a:        []float64{1, 2, 3},
			b:        []float64{1, 2, 3},
			expected: 0,
			delta:    0.001,
		},
		{
			name:     "3-4-5 triangle",
			a:        []float64{0, 0},
			b:        []float64{3, 4},
			expected: 5.0,
			delta:    0.001,
		},
		{
			name:     "unit distance",
			a:        []float64{0, 0, 0},
			b:        []float64{1, 0, 0},
			expected: 1.0,
			delta:    0.001,
		},
		{
			name:     "different length vectors",
			a:        []float64{1, 0},
			b:        []float64{1, 0, 0},
			expected: math.MaxFloat64,
			delta:    0,
		},
		{
			name:     "empty vectors",
			a:        []float64{},
			b:        []float64{},
			expected: math.MaxFloat64,
			delta:    0,
		},
		{
			name:     "negative coordinates",
			a:        []float64{-1, -1},
			b:        []float64{1, 1},
			expected: 2.828,
			delta:    0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EuclideanDistance(tt.a, tt.b)
			if tt.delta == 0 {
				if result != tt.expected {
					t.Errorf("EuclideanDistance(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
				}
			} else if math.Abs(result-tt.expected) > tt.delta {
				t.Errorf("EuclideanDistance(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestTopKBySimilarity(t *testing.T) {
	tests := []struct {
		name       string
		query      []float64
		vectors    [][]float64
		k          int
		threshold  float64
		wantLen    int
		wantFirst  int
		wantNil    bool
	}{
		{
			name:  "basic top-2",
			query: []float64{1, 0, 0},
			vectors: [][]float64{
				{1, 0, 0},    // identical - score 1.0
				{0.9, 0.1, 0}, // similar
				{0, 1, 0},     // orthogonal - score 0
				{0.5, 0.5, 0}, // moderate
			},
			k:         2,
			threshold: 0.5,
			wantLen:   2,
			wantFirst: 0,
		},
		{
			name:  "threshold filters out",
			query: []float64{1, 0, 0},
			vectors: [][]float64{
				{1, 0, 0},  // score 1.0
				{0, 1, 0},  // score 0.0
				{-1, 0, 0}, // score -1.0
			},
			k:         10,
			threshold: 0.5,
			wantLen:   1,
			wantFirst: 0,
		},
		{
			name:      "k=0 returns nil",
			query:     []float64{1, 0, 0},
			vectors:   [][]float64{{1, 0, 0}},
			k:         0,
			threshold: 0,
			wantNil:   true,
		},
		{
			name:      "empty vectors returns nil",
			query:     []float64{1, 0, 0},
			vectors:   [][]float64{},
			k:         5,
			threshold: 0,
			wantNil:   true,
		},
		{
			name:  "k larger than results",
			query: []float64{1, 0, 0},
			vectors: [][]float64{
				{1, 0, 0},
				{0.9, 0.1, 0},
			},
			k:         10,
			threshold: 0,
			wantLen:   2,
		},
		{
			name:  "all below threshold",
			query: []float64{1, 0, 0},
			vectors: [][]float64{
				{0, 1, 0},
				{0, 0, 1},
			},
			k:         5,
			threshold: 0.9,
			wantLen:   0,
		},
		{
			name:  "negative k returns nil",
			query: []float64{1, 0, 0},
			vectors: [][]float64{
				{1, 0, 0},
			},
			k:       -1,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TopKBySimilarity(tt.query, tt.vectors, tt.k, tt.threshold)

			if tt.wantNil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}

			if len(result) != tt.wantLen {
				t.Errorf("Expected length %d, got %d", tt.wantLen, len(result))
				return
			}

			if tt.wantLen > 0 && tt.wantFirst >= 0 {
				if result[0].Index != tt.wantFirst {
					t.Errorf("Expected first index %d, got %d", tt.wantFirst, result[0].Index)
				}
			}

			// Verify sorted order
			for i := 1; i < len(result); i++ {
				if result[i].Score > result[i-1].Score {
					t.Errorf("Results not sorted: score[%d]=%f > score[%d]=%f",
						i, result[i].Score, i-1, result[i-1].Score)
				}
			}
		})
	}
}

func TestNormalizeVector(t *testing.T) {
	tests := []struct {
		name        string
		v           []float64
		expectNorm1 bool
		expectSame  bool
	}{
		{
			name:        "standard vector",
			v:           []float64{3, 4},
			expectNorm1: true,
		},
		{
			name:        "already normalized",
			v:           []float64{1, 0, 0},
			expectNorm1: true,
		},
		{
			name:        "empty vector",
			v:           []float64{},
			expectSame:  true,
		},
		{
			name:        "zero vector",
			v:           []float64{0, 0, 0},
			expectSame:  true,
		},
		{
			name:        "negative values",
			v:           []float64{-3, -4},
			expectNorm1: true,
		},
		{
			name:        "single element",
			v:           []float64{5},
			expectNorm1: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeVector(tt.v)

			if tt.expectSame {
				if len(result) != len(tt.v) {
					t.Errorf("Expected same length")
				}
				for i := range result {
					if result[i] != tt.v[i] {
						t.Errorf("Expected same values")
					}
				}
				return
			}

			if tt.expectNorm1 {
				var norm float64
				for _, val := range result {
					norm += val * val
				}
				norm = math.Sqrt(norm)

				if math.Abs(norm-1.0) > 0.001 {
					t.Errorf("Expected unit vector (norm=1), got norm=%f", norm)
				}
			}
		})
	}
}

func TestScoredItemZeroValue(t *testing.T) {
	var item ScoredItem

	if item.Index != 0 {
		t.Errorf("Expected zero value Index to be 0, got %d", item.Index)
	}
	if item.Score != 0 {
		t.Errorf("Expected zero value Score to be 0, got %f", item.Score)
	}
}

func BenchmarkCosineSimilarity(b *testing.B) {
	a := make([]float64, 1536)
	vec := make([]float64, 1536)
	for i := range a {
		a[i] = float64(i) / 1536
		vec[i] = float64(1536-i) / 1536
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CosineSimilarity(a, vec)
	}
}

func BenchmarkTopKBySimilarity(b *testing.B) {
	query := make([]float64, 1536)
	vectors := make([][]float64, 1000)
	for i := range query {
		query[i] = float64(i) / 1536
	}
	for i := range vectors {
		vectors[i] = make([]float64, 1536)
		for j := range vectors[i] {
			vectors[i][j] = float64((i+j)%1536) / 1536
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TopKBySimilarity(query, vectors, 10, 0.5)
	}
}
