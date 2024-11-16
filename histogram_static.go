package metrics

import (
	"fmt"
	"io"
	"math"
	"sort"
	"sync"
	"time"
)

// DefBuckets are the default Histogram buckets. The default buckets are
// tailored to broadly measure the response time (in seconds) of a network
// service.
var DefBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

// LinearBuckets creates 'count' regular buckets, each 'width' wide, where the
// lowest bucket has an upper bound of 'start'. The final +Inf bucket is not
// counted and not included in the returned slice.
//
// start, width must not be negative, and count must be positive.
func LinearBuckets(start, width float64, count int) []float64 {
	if start < 0 {
		panic("BUG: start negative")
	}
	if width < 0 {
		panic("BUG: width negative")
	}
	if count < 1 {
		panic("BUG: count not positive")
	}
	buckets := make([]float64, count)
	for i := range buckets {
		buckets[i] = start
		start += width
	}
	return buckets
}

// ExponentialBuckets creates 'count' regular buckets, where the lowest bucket
// has an upper bound of 'start' and each following bucket's upper bound is
// 'factor' times the previous bucket's upper bound. The final +Inf bucket is
// not counted and not included in the returned slice.
//
// start must be positive, factor must be greater than 1, and count must be positive.
func ExponentialBuckets(start, factor float64, count int) []float64 {
	if start <= 0 {
		panic("BUG: start not positive")
	}
	if factor <= 1 {
		panic("BUG: factor not greater than 1")
	}
	if count < 1 {
		panic("BUG: count not positive")
	}
	buckets := make([]float64, count)
	for i := range buckets {
		buckets[i] = start
		start *= factor
	}
	return buckets
}

// ExponentialBucketsRange creates 'count' buckets, where the lowest bucket is
// 'min' and the highest bucket is 'max'. The final +Inf bucket is not counted
// and not included in the returned slice.
//
// min must be positive, max must be greater than min, count must be positive.
func ExponentialBucketsRange(minBucket, maxBucket float64, count int) []float64 {
	if minBucket <= 0 {
		panic("BUG: minBucket not positive")
	}
	if maxBucket <= minBucket {
		panic("BUG: maxBucket not greater than minBucket")
	}
	if count < 1 {
		panic("BUG: count not positive")
	}

	// Formula for exponential buckets.
	// max = min*growthFactor^(bucketCount-1)

	// We know max/min and highest bucket. Solve for growthFactor.
	growthFactor := math.Pow(maxBucket/minBucket, 1.0/float64(count-1))

	// Now that we know growthFactor, solve for each bucket.
	buckets := make([]float64, count)
	for i := 1; i <= count; i++ {
		buckets[i-1] = minBucket * math.Pow(growthFactor, float64(i-1))
	}
	return buckets
}

// HistogramStatic is a histogram for non-negative values with statically created buckets.
//
// Each bucket contains a counter for values in the given range.
// Each non-empty bucket is exposed via the following metric:
//
//	<metric_name>_bucket{<optional_tags>,le="<end>"} <counter>
//
// Where:
//
//   - <metric_name> is the metric name passed to NewHistogram
//   - <optional_tags> is optional tags for the <metric_name>, which are passed to NewHistogram
//   - <end> - (less or equal) values for the given bucket
//   - <counter> - the number of hits to the given bucket during Update* calls
//
// Zero histogram is usable.
type HistogramStatic struct {
	// Mu gurantees synchronous update for all the counters and sum.
	mu sync.Mutex

	// buckets contains counters for histogram buckets
	buckets []leBucket

	// upper is the number of values, which hit the upper bucket +Inf
	upper uint64

	// sum is the sum of all the values put into Histogram
	sum float64
}

// Reset resets the given histogram.
func (h *HistogramStatic) Reset() {
	h.mu.Lock()
	for i := range h.buckets {
		h.buckets[i].count = 0
	}
	h.upper = 0
	h.sum = 0
	h.mu.Unlock()
}

// Update updates h with v.
//
// Negative values and NaNs are ignored.
func (h *HistogramStatic) Update(v float64) {
	if math.IsNaN(v) || v < 0 {
		// Skip NaNs and negative values.
		return
	}
	h.mu.Lock()
	h.sum += v

	if len(h.buckets) == 0 || v > h.buckets[len(h.buckets)-1].le {
		h.upper++
	} else {
		idx := sort.Search(len(h.buckets), func(i int) bool {
			return v <= h.buckets[i].le
		})
		h.buckets[idx].count++
	}
	h.mu.Unlock()
}

// VisitBuckets calls f for all buckets with counters.
//
// le contains "<end>" end with bucket bounds. The lower bound
// isn't included in the bucket, while the upper bound is included.
// This is required to be compatible with Prometheus-style histogram buckets
// with `le` (less or equal) labels.
func (h *HistogramStatic) VisitBuckets(f func(le string, count uint64)) {
	h.mu.Lock()

	for _, b := range h.buckets {
		f(fmt.Sprintf("%.3e", b.le), b.count)
	}

	f("+Inf", h.upper)
	h.mu.Unlock()
}

// NewHistogramStatic creates and returns new histogram with the given name and buckets.
//
// name must be valid Prometheus-compatible metric with possible labels.
// For instance,
//
//   - foo
//   - foo{bar="baz"}
//   - foo{bar="baz",aaa="b"}
//
// The returned histogram is safe to use from concurrent goroutines.
func NewHistogramStatic(name string, buckets []float64) *HistogramStatic {
	return defaultSet.NewHistogramStatic(name, buckets)
}

// GetOrCreateHistogramStatic returns registered histogram with the given name
// or creates new histogram if the registry doesn't contain histogram with
// the given name.
//
// name must be valid Prometheus-compatible metric with possible labels.
// For instance,
//
//   - foo
//   - foo{bar="baz"}
//   - foo{bar="baz",aaa="b"}
//
// The returned histogram is safe to use from concurrent goroutines.
//
// Performance tip: prefer NewHistogramStatic instead of GetOrCreateHistogramStatic.
func GetOrCreateHistogramStatic(name string, buckets []float64) *HistogramStatic {
	return defaultSet.GetOrCreateStaticHistogram(name, buckets)
}

// UpdateDuration updates request duration based on the given startTime.
func (h *HistogramStatic) UpdateDuration(startTime time.Time) {
	d := time.Since(startTime).Seconds()
	h.Update(d)
}

func (h *HistogramStatic) marshalTo(prefix string, w io.Writer) {
	countTotal := uint64(0)
	h.VisitBuckets(func(le string, count uint64) {
		tag := fmt.Sprintf("le=%q", le)
		metricName := addTag(prefix, tag)
		name, labels := splitMetricName(metricName)
		countTotal += count
		fmt.Fprintf(w, "%s_bucket%s %d\n", name, labels, countTotal)
	})
	if countTotal == 0 {
		return
	}
	name, labels := splitMetricName(prefix)
	sum := h.getSum()
	if float64(int64(sum)) == sum {
		fmt.Fprintf(w, "%s_sum%s %d\n", name, labels, int64(sum))
	} else {
		fmt.Fprintf(w, "%s_sum%s %g\n", name, labels, sum)
	}
	fmt.Fprintf(w, "%s_count%s %d\n", name, labels, countTotal)
}

func (h *HistogramStatic) getSum() float64 {
	h.mu.Lock()
	sum := h.sum
	h.mu.Unlock()
	return sum
}

func (h *HistogramStatic) metricType() string {
	return "histogram"
}

type leBucket struct {
	le    float64
	count uint64
}
