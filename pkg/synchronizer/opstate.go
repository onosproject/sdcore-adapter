// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package synchronizer implements a synchronizer for Aether models
package synchronizer

import (
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"github.com/onosproject/sdcore-adapter/pkg/metrics"
	"strconv"
	"time"
)

// Information we keep track of when we call prometheus. The endpoint of the
// prometheus server, as well as the enterprise and site that the opstate will
// be pushed to.
type prometheusInfo struct {
	entID    string
	siteID   string
	endpoint string
}

// given an IMSI, return the simCard that has that IMSI, or nil if none exists
func (s *Synchronizer) getSimCardByImsi(site *Site, imsi uint64) *SimCard {
	for _, sim := range site.SimCard {
		if sim.Imsi != nil && *sim.Imsi == imsi {
			return sim
		}
	}
	return nil
}

// given an IMSI, return the device that has a simcart that has that IMSI, or nil if none exists
func (s *Synchronizer) getDeviceByImsi(site *Site, imsi uint64) *Device {
	sim := s.getSimCardByImsi(site, imsi)
	if sim == nil {
		return nil
	}
	for _, dev := range site.Device {
		if *dev.SimCard == *sim.SimId {
			return dev
		}
	}
	return nil
}

// get a prometheus metrics fetcher, caching previously-created fetchers
func (s *Synchronizer) getFetcher(endpoint string) (*metrics.Fetcher, error) {
	mf, okay := s.prometheus[endpoint]
	if okay {
		return mf, nil
	}

	mf, err := metrics.NewFetcher(endpoint)
	if err != nil {
		return nil, err
	}
	s.prometheus[endpoint] = mf
	return mf, nil
}

// update Device.IPAddress opstate
func (s *Synchronizer) updateDeviceIPAddress(config *gnmi.ConfigForest, entID string, siteID string, imsi uint64, mobileIP string, connected float64) {
	// Hold the lock while updating the config tree
	config.Mu.Lock()
	defer config.Mu.Unlock()

	entRoot, okay := config.Configs[entID]
	if !okay {
		// Rare corner case -- the enterprise has disappeared. Maybe some other thread deleted it
		log.Warnf("Failed to get enterprise %s", entID)
		return
	}
	enterprise := entRoot.(*RootDevice)

	site, okay := enterprise.Site[siteID]
	if !okay {
		// Rare corner case -- the site has disappeared. Maybe some other thread deleted it
		log.Warnf("Failed to get site %s of enterprise %s", siteID, entID)
		return
	}

	device := s.getDeviceByImsi(site, imsi)
	if device == nil {
		// Can happen if the device is reported in prometheus, but does not exist in
		// ROC. Rare in real life, but plausible in test infrastructure.
		// Message is currently disabled, as it is fairly noisy.
		//log.Debugf("Failed to find device %v in site %s of enterprise %s", imsi, siteID, entID)
		return
	}

	if device.State == nil {
		device.State = &DeviceState{}
	}
	device.State.IpAddress = aStr(mobileIP)

	if connected != 0 {
		device.State.Connected = aStr("Yes")
	} else {
		device.State.Connected = aStr("No")
	}
}

// poll promtheus for metrics, and add them to the config tree
func (s *Synchronizer) pollPrometheus(config *gnmi.ConfigForest, prom *prometheusInfo) error {
	log.Debugf("Polling prometheus site %s enterprise %s endpoint %s", prom.siteID, prom.entID, prom.endpoint)

	fetcher, err := s.getFetcher(prom.endpoint)
	if err != nil {
		return err
	}
	v, err := fetcher.GetVector("subscribers_info")
	if err != nil {
		return err
	}

	// vector is []*Sample
	// sample is {Metric Value:SampleValue Time}
	// metric is a labelset
	// labelsite is map[labelname]labelvalue
	// samplevalue is float64

	for _, sample := range v {
		// labels will be imsi and mobile_ip
		imsiString, okay := sample.Metric["imsi"]
		if !okay {
			log.Warn("Failed to get imsi from sample")
			continue
		}
		mobileIP, okay := sample.Metric["mobile_ip"]
		if !okay {
			log.Warn("Failed to get mobile_ip from sample")
			continue
		}

		imsi, err := strconv.ParseUint(string(imsiString), 10, 64)
		if err != nil {
			log.Warn("Failed to convert imsi %v from sample", imsiString)
			continue
		}

		s.updateDeviceIPAddress(config, prom.entID, prom.siteID, imsi, string(mobileIP), float64(sample.Value))
	}

	return nil
}

// build a list of all prometheus instances that we want to poll
func (s *Synchronizer) buildPrometheusList(config *gnmi.ConfigForest) []*prometheusInfo {
	config.Mu.Lock()
	defer config.Mu.Unlock()

	prometheusList := []*prometheusInfo{}

	// Note 1: This query could be executed in parallel using goroutines
	//
	// Note 2: If there is one ACC that is used by multiple sites, for example the Pronto scenario,
	//         then we will execute the query multiple times, which is inefficient.

	for entID, entRoot := range config.Configs {
		enterprise := entRoot.(*RootDevice)
		for siteID, site := range enterprise.Site {
			if site.ConnectivityService != nil &&
				site.ConnectivityService.Core_4G != nil &&
				site.ConnectivityService.Core_4G.AccPrometheusUrl != nil &&
				*site.ConnectivityService.Core_4G.AccPrometheusUrl != "" {
				prometheusList = append(prometheusList, &prometheusInfo{entID: entID, siteID: siteID, endpoint: *site.ConnectivityService.Core_4G.AccPrometheusUrl})
			}
		}
	}

	return prometheusList
}

func (s *Synchronizer) pollAllPrometheus(config *gnmi.ConfigForest) {
	promList := s.buildPrometheusList(config)

	// Walk the list of prometheus instances. We do this without holding the
	// lock, so we do not block the config trees while waiting on prometheus.
	// We will lock the tree when we go to update a value.
	for _, promInfo := range promList {
		err := s.pollPrometheus(config, promInfo)
		if err != nil {
			log.Warnf("Error while polling prometheus: %v", err)
		}
	}
}

// Repeatedly loop, polling prometheus for changes.
func (s *Synchronizer) pollOpstateLoop(config *gnmi.ConfigForest) {
	for {
		s.pollAllPrometheus(config)
		time.Sleep(10 * time.Second)
	}
}

// Start opstate synchronization thread
func (s *Synchronizer) startOpstate(config *gnmi.ConfigForest) {
	s.opstateStarted = true
	go s.pollOpstateLoop(config)
}
