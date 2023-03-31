package ethclient

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	req     *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

const (
	labelApp     = "app"
	labelChain   = "chain"
	labelSuccess = "success"
	labelMethod  = "method"
	labelClient  = "client"
)

var (
	labels        = []string{labelMethod, labelClient, labelSuccess}
	latencyBucket = []float64{
		2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048,
	}
)

func newMetrics(appName string, chainName string) *metrics {
	return &metrics{
		req: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rpc_request_total",
				Help: "RPC requests counts",
				ConstLabels: prometheus.Labels{
					labelApp:   appName,
					labelChain: chainName,
				},
			}, labels),
		latency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "rpc_latency_milliseconds",
				Help:    "RPC request latency in milliseconds",
				Buckets: latencyBucket,
				ConstLabels: prometheus.Labels{
					labelApp:   appName,
					labelChain: chainName},
			}, labels),
	}
}

func (m *metrics) Register() {
	prometheus.MustRegister(m.req)
	prometheus.MustRegister(m.latency)
}

func (m *metrics) Unregister() {
	prometheus.Unregister(m.req)
	prometheus.Unregister(m.latency)
}

func (s *metrics) Observe(method string, startedAt time.Time, client string, successful bool) {
	s.req.With(prometheus.Labels{
		labelMethod:  method,
		labelClient:  client,
		labelSuccess: strconv.FormatBool(successful),
	}).Inc()
	s.latency.With(prometheus.Labels{
		labelMethod:  method,
		labelClient:  client,
		labelSuccess: strconv.FormatBool(successful),
	}).Observe(float64(time.Since(startedAt).Milliseconds()))
}
