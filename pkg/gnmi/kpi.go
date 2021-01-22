package gnmi

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	gnmiRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_gnmi_requests_total",
		Help: "The total number of GNMI requests",
	},
		[]string{"method"},
	)

	gnmiRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "api_gnmi_requests_duration",
		Help: "The duration of GNMI requests",
	},
		[]string{"method"},
	)

	gnmiRequestsFailedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_gnmi_requests_failed_total",
		Help: "The total number of failed GNMI requests",
	},
		[]string{"method"},
	)
)
