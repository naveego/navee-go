package logging

import "github.com/prometheus/client_golang/prometheus"

var requestDurationsHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_request_duration_seconds",
	Help: "Histogram of API latency.",
}, []string{"method", "route"})

var requestBytesHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "http_request_bytes",
	Help:    "Histogram of API incoming bytes, labelled by method and route.",
	Buckets: prometheus.LinearBuckets(0, 1000, 5),
}, []string{"method", "route"})

var responseBytesHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "http_response_bytes",
	Help:    "Histogram of API outgoing bytes, labelled by method and route.",
	Buckets: prometheus.LinearBuckets(0, 1000, 5),
}, []string{"method", "route"})

func init() {
	prometheus.MustRegister(requestDurationsHistogram)
	prometheus.MustRegister(requestBytesHistogram)
	prometheus.MustRegister(responseBytesHistogram)
}
