[![Build Status](https://github.com/itcomusic/metrics/workflows/main/badge.svg)](https://github.com/itcomusic/metrics/actions)
[![GoDoc](https://pkg.go.dev/badge/github.com/itcomusic/metrics.svg)](http://pkg.go.dev/github.com/itcomusic/metrics)
[![codecov](https://codecov.io/gh/itcomusic/metrics/branch/master/graph/badge.svg)](https://codecov.io/gh/itcomusic/metrics)


# metrics - lightweight package for exporting metrics in Prometheus format

* current package has not modified original code
* add compatibility Prometheus histograms, `metrics.NewHistogramStatic`
* add ability pre-define buckets `metrics.DefBuckets`, `metrics.LinearBuckets`, `metrics.ExponentialBuckets`, `metrics.ExponentialBucketsRange`
