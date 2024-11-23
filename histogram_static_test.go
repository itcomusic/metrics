package metrics

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestHistogramStaticSerial(t *testing.T) {
	name := `TestHistogramStaticSerial`
	h := NewHistogramStatic(name, []float64{100, 113.6, 129.2, 146.8, 166.8, 189.6, 215.4, 219.8})

	// Verify that the histogram is visible in the output of WritePrometheus when it has no data.
	var bb bytes.Buffer
	WritePrometheus(&bb, false)
	result := bb.String()
	if !strings.Contains(result, name) {
		t.Fatalf("histogram %s shouldn be visible in the WritePrometheus output; got\n%s", name, result)
	}

	// Write data to histogram
	for i := 98; i < 221; i++ {
		h.Update(float64(i))
	}

	// Make sure the histogram prints <prefix>_bucket on marshalTo call
	testMarshalTo(t, h, "prefix", `prefix_bucket{le="1.000e+02"} 3
prefix_bucket{le="1.136e+02"} 16
prefix_bucket{le="1.292e+02"} 32
prefix_bucket{le="1.468e+02"} 49
prefix_bucket{le="1.668e+02"} 69
prefix_bucket{le="1.896e+02"} 92
prefix_bucket{le="2.154e+02"} 118
prefix_bucket{le="2.198e+02"} 122
prefix_bucket{le="+Inf"} 123
prefix_sum 19557
prefix_count 123
`)
	testMarshalTo(t, h, `	  m{foo="bar"}`, `	  m_bucket{foo="bar",le="1.000e+02"} 3
	  m_bucket{foo="bar",le="1.136e+02"} 16
	  m_bucket{foo="bar",le="1.292e+02"} 32
	  m_bucket{foo="bar",le="1.468e+02"} 49
	  m_bucket{foo="bar",le="1.668e+02"} 69
	  m_bucket{foo="bar",le="1.896e+02"} 92
	  m_bucket{foo="bar",le="2.154e+02"} 118
	  m_bucket{foo="bar",le="2.198e+02"} 122
	  m_bucket{foo="bar",le="+Inf"} 123
	  m_sum{foo="bar"} 19557
	  m_count{foo="bar"} 123
`)

	// Verify Reset
	h.Reset()
	bb.Reset()
	WritePrometheus(&bb, false)
	result = bb.String()
	if !strings.Contains(result, name) {
		t.Fatalf("unexpected histogram %s in the WritePrometheus output; got\n%s", name, result)
	}

	// Verify supported ranges
	for e10 := -100; e10 < 100; e10++ {
		for offset := 0; offset < bucketsPerDecimal; offset++ {
			m := 1 + math.Pow(bucketMultiplier, float64(offset))
			f1 := m * math.Pow10(e10)
			h.Update(f1)
			f2 := (m + 0.5*bucketMultiplier) * math.Pow10(e10)
			h.Update(f2)
			f3 := (m + 2*bucketMultiplier) * math.Pow10(e10)
			h.Update(f3)
		}
	}
	h.UpdateDuration(time.Now().Add(-time.Minute))

	// Verify edge cases
	h.Update(0)
	h.Update(math.Inf(1))
	h.Update(math.Inf(-1))
	h.Update(math.NaN())
	h.Update(-123)
	// See https://github.com/VictoriaMetrics/VictoriaMetrics/issues/1096
	h.Update(math.Float64frombits(0x3e112e0be826d695))

	// Make sure the histogram becomes visible in the output of WritePrometheus,
	// since now it contains values.
	bb.Reset()
	WritePrometheus(&bb, false)
	result = bb.String()
	if !strings.Contains(result, name) {
		t.Fatalf("missing histogram %s in the WritePrometheus output; got\n%s", name, result)
	}
}

func TestHistogramStaticConcurrent(t *testing.T) {
	name := "HistogramStaticConcurrent"
	h := NewHistogramStatic(name, []float64{0.6813, 0.7743, 0.8799, 1, 1.136, 1.292, 1.468})
	err := testConcurrent(func() error {
		for f := 0.6; f < 1.4; f += 0.1 {
			h.Update(f)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	testMarshalTo(t, h, "prefix", `prefix_bucket{le="6.813e-01"} 5
prefix_bucket{le="7.743e-01"} 10
prefix_bucket{le="8.799e-01"} 15
prefix_bucket{le="1.000e+00"} 25
prefix_bucket{le="1.136e+00"} 30
prefix_bucket{le="1.292e+00"} 35
prefix_bucket{le="1.468e+00"} 40
prefix_bucket{le="+Inf"} 40
prefix_sum 38
prefix_count 40
`)

	var labels []string
	var counts []uint64
	h.VisitBuckets(func(label string, count uint64) {
		labels = append(labels, label)
		counts = append(counts, count)
	})
	labelsExpected := []string{
		"6.813e-01",
		"7.743e-01",
		"8.799e-01",
		"1.000e+00",
		"1.136e+00",
		"1.292e+00",
		"1.468e+00",
		"+Inf",
	}
	if !reflect.DeepEqual(labels, labelsExpected) {
		t.Fatalf("unexpected labels; got %v; want %v", labels, labelsExpected)
	}
	countsExpected := []uint64{5, 5, 5, 10, 5, 5, 5, 0}
	if !reflect.DeepEqual(counts, countsExpected) {
		t.Fatalf("unexpected counts; got %v; want %v", counts, countsExpected)
	}
}

func TestHistogramStaticWithTags(t *testing.T) {
	name := `TestHistogramStatic{tag="foo"}`
	h := NewHistogramStatic(name, []float64{})
	h.Update(123)

	var bb bytes.Buffer
	WritePrometheus(&bb, false)
	result := bb.String()
	namePrefixWithTag := `TestHistogramStatic_bucket{tag="foo",le="+Inf"} 1` + "\n"
	if !strings.Contains(result, namePrefixWithTag) {
		t.Fatalf("missing histogram %s in the WritePrometheus output; got\n%s", namePrefixWithTag, result)
	}
}

func TestGetOrCreateHistogramStaticSerial(t *testing.T) {
	name := "GetOrCreateHistogramStaticSerial"
	if err := testGetOrCreateHistogramStatic(name); err != nil {
		t.Fatal(err)
	}
}

func TestGetOrCreateHistogramStaticConcurrent(t *testing.T) {
	name := "GetOrCreateHistogramStaticConcurrent"
	err := testConcurrent(func() error {
		return testGetOrCreateHistogramStatic(name)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestHistogramStaticInvalidBuckets(t *testing.T) {
	name := "HistogramStaticInvalidBuckets"
	expectPanic(t, name, func() {
		NewHistogramStatic(name, []float64{123, -234})
	})
}

func TestGetOrCreateHistogramStaticInvalidBuckets(t *testing.T) {
	name := "GetOrCreateHistogramStaticInvalidBuckets"
	expectPanic(t, name, func() {
		GetOrCreateHistogramStatic(name, []float64{123, -234})
	})
}

func testGetOrCreateHistogramStatic(name string) error {
	h1 := GetOrCreateHistogramStatic(name, []float64{1})
	for i := 0; i < 10; i++ {
		h2 := GetOrCreateHistogramStatic(name, []float64{1})
		if h1 != h2 {
			return fmt.Errorf("unexpected histogram returned; got %p; want %p", h2, h1)
		}
	}
	return nil
}
