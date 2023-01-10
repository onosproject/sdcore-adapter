// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package synchronizer implements a synchronizer for Aether models
package synchronizer

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/onosproject/analytics/pkg/kafkaClient"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"github.com/onosproject/sdcore-adapter/pkg/promkafka"
)

var (
	kafkaURI   = flag.String("kafka_uri", "", "URI of kafka")
	kafkaTopic = flag.String("kafka_topic", "sdcore", "kafka topic to fetch from")
)

// given an IMSI, return the simCard that has that IMSI, or nil if none exists
func (s *Synchronizer) getSimCardFromSiteByImsi(site *Site, imsi string) *SimCard {
	for _, sim := range site.SimCard {
		if sim.Imsi != nil && *sim.Imsi == imsi {
			return sim
		}
	}
	return nil
}

// given an IMSI, return the device that has a simcart that has that IMSI, or nil if none exists
func (s *Synchronizer) getDeviceFromSiteByImsi(site *Site, imsi string) *Device {
	sim := s.getSimCardFromSiteByImsi(site, imsi)
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

// given an IMSI, find the device by searching all enterprises and sites
func (s *Synchronizer) getDeviceByImsi(config *gnmi.ConfigForest, imsi string) *Device {
	for _, entRoot := range config.Configs {
		enterprise := entRoot.(*RootDevice)
		for _, site := range enterprise.Site {
			dev := s.getDeviceFromSiteByImsi(site, imsi)
			if dev != nil {
				return dev
			}
		}
	}

	return nil
}

func (s *Synchronizer) handleKafkaIPAddress(config *gnmi.ConfigForest, event *promkafka.IPAddressEvent) {
	config.Mu.Lock()
	defer config.Mu.Unlock()

	device := s.getDeviceByImsi(config, event.Imsi)
	if device == nil {
		// Can happen if the device is reported in prometheus, but does not exist in
		// ROC. Rare in real life, but plausible in test infrastructure.
		// Message is currently disabled, as it is fairly noisy.
		log.Debugf("Failed to find device %v", event.Imsi)
		return
	}

	log.Debugf("Handling kafka update %s: ip=%s, connected=%v", *device.DeviceId, event.IPAddress, event.Connected)

	if device.State == nil {
		device.State = &DeviceState{}
	}
	device.State.IpAddress = &event.IPAddress

	if event.Connected {
		device.State.Connected = aStr("Yes")
	} else {
		device.State.Connected = aStr("No")
	}
}

// Repeatedly loop, receiving messages from Kafka
func (s *Synchronizer) receiveKafkaLoop(config *gnmi.ConfigForest) {
	log.Info("starting kafka receiver loop")
	for {
		select {
		case stringMsg := <-s.kafkaMsgChannel:
			var event promkafka.IPAddressEvent
			err := json.Unmarshal([]byte(stringMsg), &event)
			if err != nil {
				log.Warnf("Error unmarshaling: %v", err)
			} else {
				go s.handleKafkaIPAddress(config, &event)
			}
		case err := <-s.kafkaErrorChannel:
			log.Warnf("Kafka Error: %v", err)
		}
	}
}

// Start opstate synchronization thread
func (s *Synchronizer) startOpstate(config *gnmi.ConfigForest) {
	s.opstateStarted = true

	if *kafkaURI == "" {
		log.Info("no kafkaURI specified; not starting kafka client")
		return
	}

	log.Info("starting opstate processor on topic %s for URI %s", *kafkaTopic, *kafkaURI)

	go kafkaClient.StartTopicReader(context.Background(),
		s.kafkaMsgChannel,
		s.kafkaErrorChannel,
		[]string{*kafkaURI},
		*kafkaTopic,
		"opstate",
	)

	go s.receiveKafkaLoop(config)
}
