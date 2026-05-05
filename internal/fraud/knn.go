package fraud

import (
	"math"
	"runtime"
	"sort"
	"sync"

	"github.com/Menon04/rinha-de-backend-2026-golang/internal/dataset"
)

const k = 5

type neighbor struct {
	dist  float32
	fraud bool
}

// euclidean computes squared euclidean distance (no sqrt needed for ranking).
func euclidean(a, b *[14]float32) float32 {
	var sum float32
	for i := 0; i < 14; i++ {
		d := a[i] - b[i]
		sum += d * d
	}
	return sum
}

// Score runs KNN with k=5 and returns fraud_score and approved.
func Score(query [14]float32, refs []dataset.Ref) (float32, bool) {
	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}

	chunkSize := (len(refs) + workers - 1) / workers
	results := make([][]neighbor, workers)
	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			start := id * chunkSize
			end := start + chunkSize
			if end > len(refs) {
				end = len(refs)
			}

			top := make([]neighbor, 0, k+1)
			maxDist := float32(math.MaxFloat32)

			for i := start; i < end; i++ {
				d := euclidean(&query, &refs[i].Vector)
				if len(top) < k || d < maxDist {
					top = append(top, neighbor{dist: d, fraud: refs[i].Fraud})
					sort.Slice(top, func(a, b int) bool { return top[a].dist < top[b].dist })
					if len(top) > k {
						top = top[:k]
					}
					if len(top) == k {
						maxDist = top[k-1].dist
					}
				}
			}
			results[id] = top
		}(w)
	}
	wg.Wait()

	// merge all partial top-k
	merged := make([]neighbor, 0, workers*k)
	for _, r := range results {
		merged = append(merged, r...)
	}
	sort.Slice(merged, func(a, b int) bool { return merged[a].dist < merged[b].dist })
	if len(merged) > k {
		merged = merged[:k]
	}

	var fraudCount float32
	for _, n := range merged {
		if n.fraud {
			fraudCount++
		}
	}

	score := fraudCount / float32(k)
	return score, score < 0.6
}
