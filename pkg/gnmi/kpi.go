// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

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

	/*
	 * Don't use a single metric with a "failure" and "success" label. Instead implement two
	 * metrics so they are easily subtracted.
	 *    https://prometheus.io/docs/instrumenting/writing_exporters/
	 */

	gnmiRequestsFailedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_gnmi_requests_failed_total",
		Help: "The total number of failed GNMI requests",
	},
		[]string{"method"},
	)
)
