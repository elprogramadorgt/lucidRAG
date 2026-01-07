package vectormath

import (
	"math"
	"testing"
)

func TestCosineSimilarityIdentical(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{1, 0, 0}

	sim := CosineSimilarity(a, b)
	if math.Abs(sim-1.0) > 0.001 {
		t.Errorf("Expected similarity 1.0 for identical vectors, got %f", sim)
	}
}

func TestCosineSimilarityOrthogonal(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{0, 1, 0}

	sim := CosineSimilarity(a, b)
	if math.Abs(sim) > 0.001 {
		t.Errorf("Expected similarity 0 for orthogonal vectors, got %f", sim)
	}
}

func TestCosineSimilarityOpposite(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{-1, 0, 0}

	sim := CosineSimilarity(a, b)
	if math.Abs(sim+1.0) > 0.001 {
		t.Errorf("Expected similarity -1.0 for opposite vectors, got %f", sim)
	}
}

func TestCosineSimilarityDifferentLengths(t *testing.T) {
	a := []float64{1, 0}
	b := []float64{1, 0, 0}

	sim := CosineSimilarity(a, b)
	if sim != 0 {
		t.Errorf("Expected 0 for different length vectors, got %f", sim)
	}
}

func TestCosineSimilarityEmpty(t *testing.T) {
	a := []float64{}
	b := []float64{}

	sim := CosineSimilarity(a, b)
	if sim != 0 {
		t.Errorf("Expected 0 for empty vectors, got %f", sim)
	}
}

func TestCosineSimilarityZeroVector(t *testing.T) {
	a := []float64{0, 0, 0}
	b := []float64{1, 0, 0}

	sim := CosineSimilarity(a, b)
	if sim != 0 {
		t.Errorf("Expected 0 for zero vector, got %f", sim)
	}
}

func TestEuclideanDistanceIdentical(t *testing.T) {
	a := []float64{1, 2, 3}
	b := []float64{1, 2, 3}

	dist := EuclideanDistance(a, b)
	if dist != 0 {
		t.Errorf("Expected distance 0 for identical vectors, got %f", dist)
	}
}

func TestEuclideanDistanceKnown(t *testing.T) {
	a := []float64{0, 0}
	b := []float64{3, 4}

	dist := EuclideanDistance(a, b)
	if math.Abs(dist-5.0) > 0.001 {
		t.Errorf("Expected distance 5.0 (3-4-5 triangle), got %f", dist)
	}
}

func TestEuclideanDistanceDifferentLengths(t *testing.T) {
	a := []float64{1, 0}
	b := []float64{1, 0, 0}

	dist := EuclideanDistance(a, b)
	if dist != math.MaxFloat64 {
		t.Errorf("Expected MaxFloat64 for different length vectors, got %f", dist)
	}
}

func TestEuclideanDistanceEmpty(t *testing.T) {
	a := []float64{}
	b := []float64{}

	dist := EuclideanDistance(a, b)
	if dist != math.MaxFloat64 {
		t.Errorf("Expected MaxFloat64 for empty vectors, got %f", dist)
	}
}

func TestTopKBySimilarity(t *testing.T) {
	query := []float64{1, 0, 0}
	vectors := [][]float64{
		{1, 0, 0},    // identical
		{0.9, 0.1, 0}, // similar
		{0, 1, 0},     // orthogonal
		{0.5, 0.5, 0}, // moderate
	}

	results := TopKBySimilarity(query, vectors, 2, 0.5)

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// First result should be the identical vector
	if results[0].Index != 0 {
		t.Errorf("Expected first result to be index 0, got %d", results[0].Index)
	}
}

func TestTopKBySimilarityThreshold(t *testing.T) {
	query := []float64{1, 0, 0}
	vectors := [][]float64{
		{1, 0, 0},  // similarity 1.0
		{0, 1, 0},  // similarity 0.0
		{-1, 0, 0}, // similarity -1.0
	}

	results := TopKBySimilarity(query, vectors, 10, 0.5)

	if len(results) != 1 {
		t.Errorf("Expected 1 result above threshold, got %d", len(results))
	}
}

func TestTopKBySimilarityZeroK(t *testing.T) {
	query := []float64{1, 0, 0}
	vectors := [][]float64{{1, 0, 0}}

	results := TopKBySimilarity(query, vectors, 0, 0)
	if results != nil {
		t.Errorf("Expected nil for k=0, got %v", results)
	}
}

func TestTopKBySimilarityEmptyVectors(t *testing.T) {
	query := []float64{1, 0, 0}
	vectors := [][]float64{}

	results := TopKBySimilarity(query, vectors, 5, 0)
	if results != nil {
		t.Errorf("Expected nil for empty vectors, got %v", results)
	}
}

func TestNormalizeVector(t *testing.T) {
	v := []float64{3, 4}
	normalized := NormalizeVector(v)

	// Check unit length
	var norm float64
	for _, val := range normalized {
		norm += val * val
	}
	norm = math.Sqrt(norm)

	if math.Abs(norm-1.0) > 0.001 {
		t.Errorf("Expected unit vector (norm=1), got norm=%f", norm)
	}
}

func TestNormalizeVectorEmpty(t *testing.T) {
	v := []float64{}
	normalized := NormalizeVector(v)

	if len(normalized) != 0 {
		t.Errorf("Expected empty result for empty input")
	}
}

func TestNormalizeVectorZero(t *testing.T) {
	v := []float64{0, 0, 0}
	normalized := NormalizeVector(v)

	// Zero vector should remain zero
	for i, val := range normalized {
		if val != 0 {
			t.Errorf("Expected 0 at index %d, got %f", i, val)
		}
	}
}

func TestScoredItemZeroValue(t *testing.T) {
	var item ScoredItem

	if item.Index != 0 {
		t.Error("Expected zero value Index to be 0")
	}
	if item.Score != 0 {
		t.Error("Expected zero value Score to be 0")
	}
}
