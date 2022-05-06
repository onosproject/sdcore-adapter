// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package promkafka implements a prometheus-to-kafka gateway to take sd-core messages from
// prometheus and post them to Kafka.
package promkafka

import (
	"github.com/onosproject/analytics/pkg/kafkaClient"
	"github.com/onosproject/sdcore-adapter/pkg/metrics"
	"sync"
)

// PromKafka is a prometheus-to-kafka gateway
type PromKafka struct {
	endpoints []string

	// Promehtues fetchers for each endpoint
	prometheus map[string]*metrics.Fetcher

	ipCache map[uint64]*IPAddressEvent

	Mu sync.RWMutex // mu is the RW lock to protect the access to config

	kafkaURI    string
	kafkaTopic  string
	kafkaWriter kafkaClient.Writer
}

// PromKafkaOption is for options passed when creating a new PromKafka
type PromKafkaOption func(c *PromKafka) // nolint
