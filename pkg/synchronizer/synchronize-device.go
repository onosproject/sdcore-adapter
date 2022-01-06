// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizer

import (
	"github.com/openconfig/ygot/ygot"
	"time"

	models "github.com/onosproject/config-models/modelplugin/aether-2.0.0/aether_2_0_0"
)

// SynchronizeDevice synchronizes a device. Two sets of error state are returned:
//   1) pushFailures -- a count of pushes that failed to the core. Synchronizer should retry again later.
//   2) error -- a fatal error that occurred during synchronization.
func (s *Synchronizer) SynchronizeDevice(config ygot.ValidatedGoStruct) (int, error) {
	device := config.(*models.Device)

	if device.Enterprise == nil {
		log.Info("No enteprises")
		return 0, nil
	}

	if device.ConnectivityService == nil {
		log.Info("No connectivity services")
		return 0, nil
	}

	// For a given ConnectivityService, we want to know the list of Enterprises
	// that use it. Precompute this so we can pass a list of valid Enterprises
	// along to SynchronizeConnectivityService.
	csEntMap := map[string]map[string]bool{}
	for entID, ent := range device.Enterprise.Enterprise {
		for csID := range ent.ConnectivityService {
			m, okay := csEntMap[csID]
			if !okay {
				m = map[string]bool{}
				csEntMap[csID] = m
			}
			m[entID] = true
		}
	}

	pushFailures := 0

	// All errors are treated as nonfatal, logged, and synchronization continues with the next connectivity service.
	// PushFailures are counted and reported to the caller, who can decide whether to retry.

	for csID, cs := range device.ConnectivityService.ConnectivityService {
		if (cs.Core_5GEndpoint == nil) || (*cs.Core_5GEndpoint == "") {
			log.Warnf("Skipping connectivity service %s because it has no 5G Endpoint", *cs.Id)
			continue
		}

		// Get the list of valid Enterprises for this CS.
		// Note: This could return an empty map if there is a CS that no
		//   enterprises are linked to . In that case, we can still push models
		//   that are not directly related to an enterprise, such as profiles.
		m := csEntMap[csID]

		tStart := time.Now()
		KpiSynchronizationTotal.WithLabelValues(csID).Inc()

		pushFailures += s.SynchronizeConnectivityService(device, cs, m)

		KpiSynchronizationDuration.WithLabelValues(csID).Observe(time.Since(tStart).Seconds())
	}

	return pushFailures, nil
}

// SynchronizeConnectivityService synchronizes a connectivity service
func (s *Synchronizer) SynchronizeConnectivityService(device *models.Device, cs *models.OnfConnectivityService_ConnectivityService_ConnectivityService, validEnterpriseIds map[string]bool) int {
	log.Infof("Synchronizing Connectivity Service %s", *cs.Id)

	pushFailures := 0

	// All errors are treated as nonfatal, logged, and synchronization continues with the next model.
	// PushFailures are counted and reported to the caller, who can decide whether to retry.

	if device.DeviceGroup != nil {
		pushFailures += s.SynchronizeAllDeviceGroups(device, cs, validEnterpriseIds)
	}

	if device.Vcs != nil {
		pushFailures += s.SynchronizeAllVcs(device, cs, validEnterpriseIds)
	}

	return pushFailures
}
