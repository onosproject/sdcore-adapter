// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizerv3

import (
	//"bytes"
	//"encoding/json"
	//"errors"
	"fmt"
	"github.com/openconfig/ygot/ygot"
	//"net/http"
	//"os"
	"time"

	models "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
)

type IpDomain struct {
	Dnn         string `json:"dnn"`
	Pool        string `json:"ue-ip-pool"`
	AdminStatus string `json:"admin-status"`
	DnsPrimary  string `json:"dns-primary"`
	Mtu         uint32 `json:"mtu"`
}

type DeviceGroup struct {
	Imsis        []string `json:"imsis"`
	IpDomainName string   `json:"ip-domain-name"`
	SiteInfo     string   `json:"site-info"`
	IpDomain     IpDomain `json:"ip-domain-expanded"`
}

type SliceId struct {
	Sst uint32 `json:"sst"`
	Sd  uint32 `json:"sd"`
}

type Qos struct {
	Uplink   uint64 `json:"uplink"`
	Downlink uint64 `json:"downlink"`
}

type GNodeB struct {
	Name string `json:"name"`
	Tac  uint32 `json:"tac"`
}

type Plmn struct {
	Mcc uint32 `json:"mcc"`
	Mnc uint32 `json:"mnc"`
}

type Upf struct {
	Name string `json:"upf-name"`
	Port uint32 `json:"upf-port"`
}

type SiteInfo struct {
	SiteName string   `json:"site-name"`
	Plmn     Plmn     `json:"plmn"`
	GNodeBs  []GNodeB `json:"gNodeBs"`
	Upf      Upf      `json:"upf"`
}

type Application struct {
	Name      string `json:"app-name"`
	Endpoint  string `json:"endpoint"`
	StartPort uint32 `json:"start-port"`
	EndPort   uint32 `json:"end-port"`
	Protocol  uint32 `json:"protocol"`
}

type Slice struct {
	Id                SliceId       `json:"slice-id"`
	Qos               Qos           `json:"qos"`
	DeviceGroup       string        `json:"device-group"`
	SiteInfo          SiteInfo      `json:"site-info"`
	DenyApplication   []string      `json:"deny-application"`
	PermitApplication []string      `json:"permitted-applications"`
	Applications      []Application `json:"application-information"`
}

// NOTE: This function is nearly identical with the v2 synchronizer. Refactor?
func (s *Synchronizer) SynchronizeDevice(config ygot.ValidatedGoStruct) error {
	device := config.(*models.Device)

	if device.Enterprise == nil {
		log.Info("No enteprises")
		return nil
	}

	if device.ConnectivityService == nil {
		log.Info("No connectivity services")
		return nil
	}

	// For a given ConnectivityService, we want to know the list of Enterprises
	// that use it. Precompute this so we can pass a list of valid Enterprises
	// along to SynchronizeConnectivityService.
	csEntMap := map[string]map[string]bool{}
	for entId, ent := range device.Enterprise.Enterprise {
		for csId := range ent.ConnectivityService {
			m, okay := csEntMap[csId]
			if !okay {
				m = map[string]bool{}
				csEntMap[csId] = m
			}
			m[entId] = true
		}
	}

	errors := []error{}
	for csId, cs := range device.ConnectivityService.ConnectivityService {
		// Get the list of valid Enterprises for this CS.
		// Note: This could return an empty map if there is a CS that no
		//   enterprises are linked to . In that case, we can still push models
		//   that are not directly related to an enterprise, such as profiles.
		m := csEntMap[csId]

		tStart := time.Now()
		synchronizationTotal.WithLabelValues(csId).Inc()

		err := s.SynchronizeConnectivityService(device, cs, m)
		if err != nil {
			synchronizationFailedTotal.WithLabelValues(csId).Inc()
			// If there are errors, then build a list of them and continue to try
			// to synchronize other connectivity services.
			errors = append(errors, err)
		} else {
			synchronizationDuration.WithLabelValues(csId).Observe(time.Since(tStart).Seconds())
		}
	}

	if len(errors) == 0 {
		return nil
	} else {
		return fmt.Errorf("synchronization errors: %v", errors)
	}
}

func (s *Synchronizer) SynchronizeConnectivityService(device *models.Device, cs *models.ConnectivityService_ConnectivityService_ConnectivityService, validEnterpriseIds map[string]bool) error {
	log.Infof("Synchronizing Connectivity Service %s", *cs.Id)

	if device.DeviceGroup != nil {
		//SynchronizeDeviceGroups(device.DeviceGroup)
	}
	if device.Vcs != nil {
		//SynchronizeVCS(device.Vcs)
	}

	return nil
}
