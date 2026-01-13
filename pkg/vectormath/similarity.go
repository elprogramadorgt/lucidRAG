package vectormath

import (
	"math"
	"sort"
)

func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func EuclideanDistance(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return math.MaxFloat64
	}

	var sum float64
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	return math.Sqrt(sum)
}

type ScoredItem struct {
	Index int
	Score float64
}

func TopKBySimilarity(query []float64, vectors [][]float64, k int, threshold float64) []ScoredItem {
	if k <= 0 || len(vectors) == 0 {
		return nil
	}

	scores := make([]ScoredItem, 0, len(vectors))
	for i, v := range vectors {
		score := CosineSimilarity(query, v)
		if score >= threshold {
			scores = append(scores, ScoredItem{
				Index: i,
				Score: score,
			})
		}
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	if len(scores) > k {
		scores = scores[:k]
	}

	return scores
}

func NormalizeVector(v []float64) []float64 {
	if len(v) == 0 {
		return v
	}

	var norm float64
	for _, val := range v {
		norm += val * val
	}
	norm = math.Sqrt(norm)

	if norm == 0 {
		return v
	}

	result := make([]float64, len(v))
	for i, val := range v {
		result[i] = val / norm
	}
	return result
}
