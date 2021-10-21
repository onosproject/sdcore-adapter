// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizerv4 implements a synchronizer for converting sdcore gnmi to json
package synchronizerv4

import (
	"fmt"
	"github.com/openconfig/ygot/ygot"
	"time"

	models "github.com/onosproject/config-models/modelplugin/aether-4.0.0/aether_4_0_0"
	"github.com/onosproject/sdcore-adapter/pkg/synchronizer"
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
	errors := []error{}
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
		synchronizer.KpiSynchronizationTotal.WithLabelValues(csID).Inc()

		csPushFailures, err := s.SynchronizeConnectivityService(device, cs, m)
		pushFailures += csPushFailures
		if err != nil {
			synchronizer.KpiSynchronizationFailedTotal.WithLabelValues(csID).Inc()
			// If there are errors, then build a list of them and continue to try
			// to synchronize other connectivity services.
			errors = append(errors, err)
		} else {
			synchronizer.KpiSynchronizationDuration.WithLabelValues(csID).Observe(time.Since(tStart).Seconds())
		}
	}

	if len(errors) == 0 {
		return pushFailures, nil
	}

	return pushFailures, fmt.Errorf("synchronization errors: %v", errors)
}

// SynchronizeConnectivityService synchronizes a connectivity service
func (s *Synchronizer) SynchronizeConnectivityService(device *models.Device, cs *models.OnfConnectivityService_ConnectivityService_ConnectivityService, validEnterpriseIds map[string]bool) (int, error) {
	log.Infof("Synchronizing Connectivity Service %s", *cs.Id)

	var err error
	var dgPushFailures int
	var vcsPushFailures int

	if device.DeviceGroup != nil {
		dgPushFailures, err = s.SynchronizeAllDeviceGroups(device, cs, validEnterpriseIds)
		if err != nil {
			// nonfatal error -- we still want to try to synchronize VCS
			log.Warnf("ConnectivityService %s DeviceGroup Synchronization Error: %v", *cs.Id, err)
		}
	}
	if device.Vcs != nil {
		vcsPushFailures, err = s.SynchronizeAllVcs(device, cs, validEnterpriseIds)
		if err != nil {
			// nofatal error
			log.Warnf("ConnectivityService %s VCS Synchronization Error: %v", *cs.Id, err)
		}
	}

	return dgPushFailures + vcsPushFailures, nil
}
