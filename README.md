[![Build Status](https://github.com/itcomusic/metrics/workflows/main/badge.svg)](https://github.com/itcomusic/metrics/actions)
[![GoDoc](https://pkg.go.dev/badge/github.com/itcomusic/metrics.svg)](http://pkg.go.dev/github.com/itcomusic/metrics)
[![Coverage](https://coveralls.io/repos/github/itcomusic/metrics/badge.svg)](https://coveralls.io/github/itcomusic/metrics)


# metrics - lightweight package for exporting metrics in Prometheus format

* current package has not modified [original](https://github.com/VictoriaMetrics/metrics/releases/tag/v1.35.2) (v1.35.2)
* add compatibility Prometheus histograms, `metrics.NewHistogramStatic`
* add ability pre-define buckets `metrics.DefBuckets`, `metrics.LinearBuckets`, `metrics.ExponentialBuckets`, `metrics.ExponentialBucketsRange`
