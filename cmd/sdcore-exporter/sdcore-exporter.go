// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package main

import (
	"github.com/onosproject/sdcore-adapter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"time"
)

func main() {

	collector.RecordMetrics(2 * time.Second, "starbucks_newyork_cameras")
	collector.RecordMetrics(2 * time.Second, "starbucks_seattle_cameras")
	collector.RecordMetrics(2 * time.Second, "acme_chicago_robots")
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":2112", nil); err != nil {
		log.Fatal(err)
	}
}
