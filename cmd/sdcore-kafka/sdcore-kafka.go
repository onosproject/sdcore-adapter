// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

/*
 * sdcore-kafka: A tool for importing sdcore data from Prometheus and posting it to Kafka
 *
 * To Install a Kafka for use with this command:
 *
 * helm repo add bitnami https://charts.bitnami.com/bitnami
 * helm install -n micro-onos kafka bitnami/kafka
 *
 *  consumers: kafka.micro-onos.svc.cluster.local
 *  producers: kafka-0.kafka-headless.micro-onos.svc.cluster.local:9092
 *
 * To manually run promethes-to-kafka:
 *     cd  /usr/local/bin && ./sdcore-kafka --endpoint "http://aether-roc-umbrella-prometheus-acc-server:80"
 *
 * To exampine with Kafkacat:
 *     kafkacat -C -b kafka.micro-onos.svc.cluster.local -t sdcore
 *
 */

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/internal/pkg/version"
	"github.com/onosproject/sdcore-adapter/pkg/promkafka"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	kafkaURI  = flag.String("kafka_uri", "kafka-0.kafka-headless.micro-onos.svc.cluster.local:9092", "URI of kafka")
	endpoints arrayFlags
)

var log = logging.GetLogger("sdcore-kafka")

func main() {
	flag.Usage = func() {
		_, err := fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		if err != nil {
			log.Errorf("error: %+v", err)
		}
		flag.PrintDefaults()
	}
	flag.Var(&endpoints, "endpoint", "prometheus endpoints")
	flag.Parse()

	log.Infof("sdcore-kafka")
	version.LogVersion("  ")

	pk := promkafka.NewPromKafka(
		endpoints,
		promkafka.WithKafkaURI(*kafkaURI))

	pk.Start()

	for {
		time.Sleep(10 * time.Second)
	}
}
