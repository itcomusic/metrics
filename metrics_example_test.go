package metrics_test

import (
	"net/http"

	"github.com/itcomusic/metrics"
)

func ExampleWritePrometheus() {
	// Export all the registered metrics in Prometheus format at `/metrics` http path.
	http.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		metrics.WritePrometheus(w, true)
	})
}
