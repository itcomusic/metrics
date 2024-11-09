package metrics

import (
	"bytes"
	"math"
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
