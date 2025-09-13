package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Package-level registry and metric instruments. Call Init() before use.

var (
	registry              *prometheus.Registry
	tickTimeMsHist        prometheus.Histogram
	snapshotBytesHist     prometheus.Histogram
	entitiesInAOIHist     prometheus.Histogram
	handoverLatencyMsHist prometheus.Histogram
	wsConnectedGauge      prometheus.Gauge
	handoversTotalCounter prometheus.Counter
)

// Init (re)initializes the metrics registry and instruments. Safe to call in tests.
func Init() {
	registry = prometheus.NewRegistry()

	tickTimeMsHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "sim",
		Name:      "tick_time_ms",
		Help:      "Duration of simulation tick in milliseconds.",
		Buckets:   []float64{0.1, 0.25, 0.5, 1, 2, 5, 10, 20, 50, 100},
	})

	snapshotBytesHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "ws",
		Name:      "snapshot_bytes",
		Help:      "Size of WS state snapshot JSON payload in bytes.",
		Buckets:   []float64{256, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536},
	})

	entitiesInAOIHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "sim",
		Name:      "entities_in_aoi",
		Help:      "Number of entities returned per AOI query.",
		Buckets:   []float64{0, 1, 2, 4, 8, 16, 32, 64, 128, 256},
	})

	handoverLatencyMsHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "sim",
		Name:      "handover_latency_ms",
		Help:      "Latency from handover decision to client handover emission.",
		Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 25, 50, 100, 250, 500},
	})

	wsConnectedGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "ws",
		Name:      "connected",
		Help:      "Current number of WebSocket clients connected.",
	})

	handoversTotalCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "sim",
		Name:      "handovers_total",
		Help:      "Total handover events processed.",
	})

	registry.MustRegister(
		tickTimeMsHist,
		snapshotBytesHist,
		entitiesInAOIHist,
		handoverLatencyMsHist,
		wsConnectedGauge,
		handoversTotalCounter,
	)
}

// Handler returns an HTTP handler that serves Prometheus metrics.
func Handler() http.Handler {
	if registry == nil {
		Init()
	}
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

// ObserveTickDuration records a simulation tick duration.
func ObserveTickDuration(d time.Duration) {
	if tickTimeMsHist == nil {
		return
	}
	tickTimeMsHist.Observe(float64(d) / float64(time.Millisecond))
}

// ObserveSnapshotBytes records the size of a WS state snapshot in bytes.
func ObserveSnapshotBytes(n int) {
	if snapshotBytesHist == nil {
		return
	}
	snapshotBytesHist.Observe(float64(n))
}

// ObserveEntitiesInAOI records the number of entities returned in an AOI query.
func ObserveEntitiesInAOI(n int) {
	if entitiesInAOIHist == nil {
		return
	}
	entitiesInAOIHist.Observe(float64(n))
}

// ObserveHandoverLatency records the measured handover latency.
func ObserveHandoverLatency(d time.Duration) {
	if handoverLatencyMsHist == nil {
		return
	}
	handoverLatencyMsHist.Observe(float64(d) / float64(time.Millisecond))
}

// IncWSConnected increments the active WS client gauge.
func IncWSConnected() {
	if wsConnectedGauge == nil {
		return
	}
	wsConnectedGauge.Inc()
}

// DecWSConnected decrements the active WS client gauge.
func DecWSConnected() {
	if wsConnectedGauge == nil {
		return
	}
	wsConnectedGauge.Dec()
}

// IncHandovers increments handover counter.
func IncHandovers() {
	if handoversTotalCounter == nil {
		return
	}
	handoversTotalCounter.Inc()
}
