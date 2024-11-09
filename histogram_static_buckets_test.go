package metrics

import (
	"math"
	"reflect"
	"testing"
)

func TestLinearBuckets(t *testing.T) {
	t.Parallel()

	got := LinearBuckets(15, 5, 6)
	want := []float64{15, 20, 25, 30, 35, 40}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected linear buckets: got %v, want %v", got, want)
	}
}

func TestExponentialBuckets(t *testing.T) {
	t.Parallel()

	got := ExponentialBuckets(100, 1.2, 3)
	want := []float64{100, 120, 144}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected exponential buckets: got %v, want %v", got, want)
	}
}

func TestExponentialBucketsRange(t *testing.T) {
	t.Parallel()

	got := ExponentialBucketsRange(1, 100, 10)
	want := []float64{
		1.0, 1.6681, 2.7825, 4.6415, 7.7426, 12.9154, 21.5443,
		35.9381, 59.9484, 100.0000,
	}
	const epsilon = 0.0001
	if !almostEqualFloat64s(got, want, epsilon) {
		t.Errorf("expected exponential buckets range: got %v, want %v (epsilon %f)", got, want, epsilon)
	}
}

func TestBucketsParamInvalid(t *testing.T) {
	t.Parallel()

	expectPanic(t, "LinearBuckets_start", func() { LinearBuckets(-1, 5, 6) })
	expectPanic(t, "LinearBuckets_width", func() { LinearBuckets(15, -1, 6) })
	expectPanic(t, "LinearBuckets_count", func() { LinearBuckets(15, 5, 0) })

	expectPanic(t, "ExponentialBuckets_start", func() { ExponentialBuckets(0, 1.2, 3) })
	expectPanic(t, "ExponentialBuckets_factor", func() { ExponentialBuckets(100, 1, 3) })
	expectPanic(t, "ExponentialBuckets_count", func() { ExponentialBuckets(100, 1.2, 0) })

	expectPanic(t, "ExponentialBucketsRange_minBucket", func() { ExponentialBucketsRange(0, 100, 10) })
	expectPanic(t, "ExponentialBucketsRange_maxBucket", func() { ExponentialBucketsRange(1, 1, 10) })
	expectPanic(t, "ExponentialBucketsRange_count", func() { ExponentialBucketsRange(1, 100, 0) })
}

// minNormalFloat64 is the smallest positive normal value of type float64.
var minNormalFloat64 = math.Float64frombits(0x0010000000000000)

// AlmostEqualFloat64 returns true if a and b are equal within a relative error
// of epsilon. See http://floating-point-gui.de/errors/comparison/ for the
// details of the applied method.
func almostEqualFloat64(a, b, epsilon float64) bool {
	if a == b {
		return true
	}
	absA := math.Abs(a)
	absB := math.Abs(b)
	diff := math.Abs(a - b)
	if a == 0 || b == 0 || absA+absB < minNormalFloat64 {
		return diff < epsilon*minNormalFloat64
	}
	return diff/math.Min(absA+absB, math.MaxFloat64) < epsilon
}

// AlmostEqualFloat64s is the slice form of AlmostEqualFloat64.
func almostEqualFloat64s(a, b []float64, epsilon float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !almostEqualFloat64(a[i], b[i], epsilon) {
			return false
		}
	}
	return true
}
