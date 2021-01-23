package synchronizer

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	synchronizationTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "synchronization_total",
		Help: "The total number of synchronizations",
	},
		[]string{"cs"},
	)

	synchronizationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "synchronization_duration",
		Help: "The duration of synchronizations",
	},
		[]string{"cs"},
	)

	synchronizationFailedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "synchronization_failed_total",
		Help: "The total number of failed synchronizations",
	},
		[]string{"cs"},
	)
)
