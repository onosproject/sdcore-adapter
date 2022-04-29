// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

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

	// KpiSliceBitrate is the Configured MBR for slices
	KpiSliceBitrate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "slice_bitrate",
		Help: "per-Slice configured maximum bitrate",
	},
		[]string{"enterprise", "site", "slice", "direction"},
	)

	// KpiApplicationBitrate is the Configured MBR for Applications
	KpiApplicationBitrate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "application_bitrate",
		Help: "per-Device per-Application configured maximim bitrate",
	},
		[]string{"enterprise", "site", "slice", "application", "endpoint", "direction"},
	)

	// KpiDeviceGroupBitrate is the Configured MBR for Device Groups
	KpiDeviceGroupBitrate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "device_group_bitrate",
		Help: "per-Device configured maximum bitrate",
	},
		[]string{"enterprise", "site", "devicegroup", "direction"},
	)
)

// Helper routines follow for reporting prometheus metrics. These will make it easier when
// if we move away from prometheus and toward the Analytics Engine.

func (s *Synchronizer) reportApplicationBitrate(scope *AetherScope, slice *Slice, app *Application, endpoint *ApplicationEndpoint, direction string, value uint64) {
	KpiApplicationBitrate.WithLabelValues(*scope.Enterprise.EnterpriseId,
		*scope.Site.SiteId,
		*slice.SliceId,
		*app.ApplicationId,
		*endpoint.EndpointId,
		direction).Set(float64(value))
}

func (s *Synchronizer) reportDeviceGroupBitrate(scope *AetherScope, dg *DeviceGroup, direction string, value uint64) {
	KpiDeviceGroupBitrate.WithLabelValues(*scope.Enterprise.EnterpriseId,
		*scope.Site.SiteId,
		*dg.DeviceGroupId,
		direction).Set(float64(value))
}

func (s *Synchronizer) reportSliceBitrate(scope *AetherScope, slice *Slice, direction string, value uint64) {
	KpiSliceBitrate.WithLabelValues(*scope.Enterprise.EnterpriseId,
		*scope.Site.SiteId,
		*slice.SliceId,
		direction).Set(float64(value))
}
