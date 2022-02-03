// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizer

import (
	"github.com/openconfig/ygot/ygot"
	"time"
)

// SynchronizeDevice synchronizes a device. Two sets of error state are returned:
//   1) pushFailures -- a count of pushes that failed to the core. Synchronizer should retry again later.
//   2) error -- a fatal error that occurred during synchronization.
func (s *Synchronizer) SynchronizeDevice(config ygot.ValidatedGoStruct) (int, error) {
	device := config.(*RootDevice)

	pushFailures := 0

	if device.Enterprises == nil {
		log.Info("No enteprises")
		return 0, nil
	}

	if device.ConnectivityServices == nil {
		log.Info("No connectivity services")
		return 0, nil
	}

	for _, cs := range device.ConnectivityServices.ConnectivityService {
		tStart := time.Now()
		KpiSynchronizationTotal.WithLabelValues(*cs.ConnectivityServiceId).Inc()

		scope := &AetherScope{
			ConnectivityService: cs,
			RootDevice:          device}
		for _, enterprise := range device.Enterprises.Enterprise {
			// Does this enterprise use the current ConnectivityService?
			// If not, skip it
			hasConnectivityService := false
			for csID := range enterprise.ConnectivityService {
				if csID == *cs.ConnectivityServiceId {
					hasConnectivityService = true
				}
			}
			if !hasConnectivityService {
				continue
			}

			scope.Enterprise = enterprise
			for _, site := range enterprise.Site {
				scope.Site = site
				for _, dg := range site.DeviceGroup {
					dgPushErrors, err := s.SynchronizeDeviceGroup(scope, dg)
					if err != nil {
						log.Warnf("DG %s failed to synchronize Core: %s", *dg.DeviceGroupId, err)
					}
					pushFailures += dgPushErrors
				}
			sliceLoop:
				for _, slice := range site.Slice {
					slicePushFailures, err := s.SynchronizeSlice(scope, slice)
					pushFailures += slicePushFailures
					if err != nil {
						log.Warnf("VCS %s failed to synchronize Core: %s", *slice.SliceId, err)
						// Do not try to synchronize the UPF, if we've already failed
						continue sliceLoop
					}

					upfPushFailures, err := s.SynchronizeSliceUPF(scope, slice)
					pushFailures += upfPushFailures
					if err != nil {
						log.Warnf("Slice %s failed to synchronize UPF: %s", *slice.SliceId, err)
						continue sliceLoop
					}
				}
			}
		}

		KpiSynchronizationDuration.WithLabelValues(*cs.ConnectivityServiceId).Observe(time.Since(tStart).Seconds())
	}

	return pushFailures, nil
}
