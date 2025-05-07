package helper

import "iter"

// BackwardByBatches returns an iterator that processes the slice in reverse order
// by batches, yielding slices of items with their indices.
// This more closely matches your original implementation with batches as separate units.
func BackwardByBatches[Slice ~[]E, E any](s Slice, batchSize int) iter.Seq2[int, []struct {
	Index int
	Value E
}] {
	return func(yield func(int, []struct {
		Index int
		Value E
	}) bool) {
		totalBatches := (len(s) + batchSize - 1) / batchSize // ceiling division

		for batchIdx := totalBatches - 1; batchIdx >= 0; batchIdx-- {
			// Calculate start and end positions for this batch
			startPos := batchIdx * batchSize
			endPos := startPos + batchSize
			if endPos > len(s) {
				endPos = len(s)
			}

			// Create batch of items with their indices in reverse order
			var batch []struct {
				Index int
				Value E
			}
			for itemIdx := endPos - 1; itemIdx >= startPos; itemIdx-- {
				batch = append(batch, struct {
					Index int
					Value E
				}{
					Index: itemIdx,
					Value: s[itemIdx],
				})
			}

			if !yield(batchIdx, batch) {
				return
			}
		}
	}
}
