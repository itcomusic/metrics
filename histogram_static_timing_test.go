package metrics

import (
	"testing"
)

func BenchmarkHistogramStaticUpdate(b *testing.B) {
	// Define the fixed base buckets
	buckets := []float64{87.99, 100, 113.6, 129.2, 146.8, 166.8, 189.6, 215.4, 244.8}

	h := GetOrCreateHistogramStatic("BenchmarkHistogramStaticUpdate", buckets)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			h.Update(float64(i))
			i++
		}
	})
}
