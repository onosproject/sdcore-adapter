// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package promkafka implements a prometheus-to-kafka gateway to take sd-core messages from
// prometheus and post them to Kafka.
package promkafka

import (
	"github.com/onosproject/analytics/pkg/kafkaClient"
	"github.com/onosproject/analytics/pkg/messages"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/metrics"
	"strconv"
	"time"
)

var log = logging.GetLogger("promkafka")

// Information we keep track of when we call prometheus. The endpoint of the
// prometheus server, as well as the enterprise and site that the opstate will
// be pushed to.
type prometheusInfo struct {
	endpoint string
}

// get a prometheus metrics fetcher, caching previously-created fetchers
func (s *PromKafka) getFetcher(endpoint string) (*metrics.Fetcher, error) {
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

func (s *PromKafka) updateDeviceIPAddress(imsi uint64, mobileIP string, connected float64) error {
	ev := IPAddressEvent{Type: "ip-assignment",
		Imsi:      imsi,
		Connected: connected > 0,
		IPAddress: mobileIP}

	cached, okay := s.ipCache[imsi]
	if okay && cached.Connected == ev.Connected && cached.IPAddress == ev.IPAddress {
		return nil
	}

	b, err := messages.GetJson(ev)
	if err != nil {
		return err
	}

	err = s.kafkaWriter.SendMessage(b)

	s.ipCache[imsi] = &ev

	return err
}

// poll promtheus for metrics, and add them to the config tree
func (s *PromKafka) pollPrometheus(prom *prometheusInfo) error {
	log.Debugf("Polling prometheus endpoint %s", prom.endpoint)

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

		// do stuff here
		err = s.updateDeviceIPAddress(imsi, string(mobileIP), float64(sample.Value))
		if err != nil {
			log.Warnf("Failed to update kafka for imsi %v: %v", imsi, err)
		}
	}

	return nil
}

// build a list of all prometheus instances that we want to poll
func (s *PromKafka) buildPrometheusList() []*prometheusInfo {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	prometheusList := []*prometheusInfo{}

	for _, endpoint := range s.endpoints {
		prometheusList = append(prometheusList, &prometheusInfo{endpoint: endpoint})
	}

	return prometheusList
}

func (s *PromKafka) pollAllPrometheus() {
	promList := s.buildPrometheusList()

	// Walk the list of prometheus instances. We do this without holding the
	// lock, so we do not block the config trees while waiting on prometheus.
	// We will lock the tree when we go to update a value.
	for _, promInfo := range promList {
		err := s.pollPrometheus(promInfo)
		if err != nil {
			log.Warnf("Error while polling prometheus: %v", err)
		}
	}
}

// Repeatedly loop, polling prometheus for changes.
func (s *PromKafka) pollLoop() {
	for {
		s.pollAllPrometheus()
		time.Sleep(10 * time.Second)
	}
}

// Start opstate synchronization thread
func (s *PromKafka) Start() {
	go s.pollLoop()
}

// WithKafkaURI sets the URI of the Kafka to push to
func WithKafkaURI(uri string) PromKafkaOption {
	return func(p *PromKafka) {
		p.kafkaURI = uri
	}
}

// NewPromKafka creates a new PromKafka
func NewPromKafka(endpoints []string, opts ...PromKafkaOption) *PromKafka {
	p := &PromKafka{
		endpoints:  endpoints,
		prometheus: map[string]*metrics.Fetcher{},
		kafkaURI:   "default-kafka-uri",
		kafkaTopic: "sdcore",
		ipCache:    map[uint64]*IPAddressEvent{},
	}

	for _, opt := range opts {
		opt(p)
	}

	p.kafkaWriter = kafkaClient.GetWriter(p.kafkaURI, p.kafkaTopic)

	return p
}
