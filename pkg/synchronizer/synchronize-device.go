// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizer

import (
	"fmt"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"time"
)

// updateScopeFromSlice uses a slice object to determine the Generation (4G|5G) and the
// CoreEndpoint.
func (s *Synchronizer) updateScopeFromSlice(scope *AetherScope, slice *Slice) error {
	if scope.Site.ConnectivityService == nil {
		return fmt.Errorf("Site %s has no ConnectivityServices", *scope.Site.SiteId)
	}
	if slice.ConnectivityService == ConnectivityService4G {
		if scope.Site.ConnectivityService.Core_4G == nil {
			return fmt.Errorf("Site %s has no 4G ConnectivityServices", *scope.Site.SiteId)
		}
		if scope.Site.ConnectivityService.Core_4G.Endpoint == nil || *scope.Site.ConnectivityService.Core_4G.Endpoint == "" {
			return fmt.Errorf("Site %s 4G ConnectivityService has no endpoint", *scope.Site.SiteId)
		}
		scope.Generation = aStr("4G")
		scope.CoreEndpoint = scope.Site.ConnectivityService.Core_4G.Endpoint
		return nil
	} else if slice.ConnectivityService == ConnectivityService5G {
		if scope.Site.ConnectivityService.Core_5G == nil {
			return fmt.Errorf("Site %s has no 5G ConnectivityServices", *scope.Site.SiteId)
		}
		if scope.Site.ConnectivityService.Core_5G.Endpoint == nil || *scope.Site.ConnectivityService.Core_5G.Endpoint == "" {
			return fmt.Errorf("Site %s 5G ConnectivityService has no endpoint", *scope.Site.SiteId)
		}
		scope.Generation = aStr("5G")
		scope.CoreEndpoint = scope.Site.ConnectivityService.Core_5G.Endpoint
		return nil
	}
	return fmt.Errorf("Slice %s has unknown or undefined ConnectivityService", *slice.SliceId)
}

// updateScopeFromDeviceGroup uses a DeviceGroup object to determine the Generation (4G|5G) and the
// CoreEndpoint. It does this by finding some slice that uses the DeviceGroup.
func (s *Synchronizer) updateScopeFromDeviceGroup(scope *AetherScope, dg *DeviceGroup) error {
	for _, slice := range scope.Site.Slice {
		for dgID := range slice.DeviceGroup {
			if dgID == *dg.DeviceGroupId {
				// Each DG can only participate in one Slice, so as soon as we've found a slice
				// we're done.
				return s.updateScopeFromSlice(scope, slice)
			}
		}
	}
	// Not finding the a Slice that contains the DG is not an error. The caller will notice
	// that no CoreEndpoint was found, and react accordingly.
	// TODO smbaker: Consider optimizing out the return value if it's always nil.
	return nil
}

// SynchronizeDevice synchronizes a device. Two sets of error state are returned:
//   1) pushFailures -- a count of pushes that failed to the core. Synchronizer should retry again later.
//   2) error -- a fatal error that occurred during synchronization.
func (s *Synchronizer) SynchronizeDevice(allConfig gnmi.ConfigForest) (int, error) {
	pushFailures := 0
	for entID, enterpriseConfig := range allConfig {
		device := enterpriseConfig.(*RootDevice)

		tStart := time.Now()
		KpiSynchronizationTotal.WithLabelValues(entID).Inc()

		scope := &AetherScope{
			Enterprise: device}

		for _, site := range device.Site {
			scope.Site = site
		dgLoop:
			for _, dg := range site.DeviceGroup {
				err := s.updateScopeFromDeviceGroup(scope, dg)
				if err != nil {
					log.Warnf("DG %s error while resolving core endpoint: %s", *dg.DeviceGroupId, err)
					continue dgLoop
				}
				if scope.CoreEndpoint == nil {
					// This is not necessarily a problem; the DG might simply not be in use.
					log.Infof("DG %s is not related to any core: %s", *dg.DeviceGroupId, err)
					continue dgLoop
				}
				KpiSynchronizationResourceTotal.WithLabelValues(entID, "device-group").Inc()
				dgPushErrors, err := s.SynchronizeDeviceGroup(scope, dg)
				pushFailures += dgPushErrors
				if err != nil {
					log.Warnf("DG %s failed to synchronize Core: %s", *dg.DeviceGroupId, err)
					KpiSynchronizationFailedTotal.WithLabelValues(entID, "device-group", "core").Inc()
				}
			}
		sliceLoop:
			for _, slice := range site.Slice {
				err := s.updateScopeFromSlice(scope, slice)
				if err != nil {
					log.Warnf("DG %s error while resolving core endpoint: %s", *slice.SliceId, err)
					continue sliceLoop
				}
				if scope.CoreEndpoint == nil {
					// Slice should always be related to a connectivity service
					log.Warnf("Slice %s is not related to any core: %s", *slice.SliceId, err)
					continue sliceLoop
				}
				KpiSynchronizationResourceTotal.WithLabelValues(entID, "slice").Inc()
				slicePushFailures, err := s.SynchronizeSlice(scope, slice)
				pushFailures += slicePushFailures
				if err != nil {
					log.Warnf("VCS %s failed to synchronize Core: %s", *slice.SliceId, err)
					KpiSynchronizationFailedTotal.WithLabelValues(entID, "slice", "core").Inc()
					// Do not try to synchronize the UPF, if we've already failed
					continue sliceLoop
				}

				upfPushFailures, err := s.SynchronizeSliceUPF(scope, slice)
				pushFailures += upfPushFailures
				if err != nil {
					log.Warnf("Slice %s failed to synchronize UPF: %s", *slice.SliceId, err)
					KpiSynchronizationFailedTotal.WithLabelValues(entID, "slice", "upf").Inc()
				}
			}
		}

		KpiSynchronizationDuration.WithLabelValues(entID).Observe(time.Since(tStart).Seconds())
	}

	return pushFailures, nil
}
