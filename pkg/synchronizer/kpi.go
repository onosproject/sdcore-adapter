// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package synchronizer

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// KpiSynchronizationTotal is the count of Synchronizations
	KpiSynchronizationTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "synchronization_total",
		Help: "The total number of synchronizations",
	},
		[]string{"cs"},
	)

	// KpiSynchronizationDuration is a histogram of duration of synchronizations
	KpiSynchronizationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "synchronization_duration",
		Help: "The duration of synchronizations",
	},
		[]string{"cs"},
	)

	// KpiSynchronizationFailedTotal is a count of failed synchronizations
	KpiSynchronizationFailedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "synchronization_failed_total",
		Help: "The total number of failed synchronizations",
	},
		[]string{"cs"},
	)

	// KpiSynchronizationResourceTotal is the total number of resources synchronized.
	KpiSynchronizationResourceTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "synchronization_resource_total",
		Help: "The total number of resources synchronized",
	},
		[]string{"cs", "kind"},
	)
)
