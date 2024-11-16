package metrics_test

import (
	"fmt"
	"time"

	"github.com/itcomusic/metrics"
)

func ExampleHistogramStatic() {
	// Define a histogram in global scope.
	var h = metrics.NewHistogramStatic(`request_duration_seconds{path="/foo/bar"}`, nil)

	// Update the histogram with the duration of processRequest call.
	startTime := time.Now()
	processRequest()
	h.UpdateDuration(startTime)
}

func ExampleHistogramStatic_vec() {
	for i := 0; i < 3; i++ {
		// Dynamically construct metric name and pass it to GetOrCreateHistogram.
		name := fmt.Sprintf(`response_size_bytes{path=%q}`, "/foo/bar")
		response := processRequest()
		metrics.GetOrCreateHistogramStatic(name, nil).Update(float64(len(response)))
	}
}
