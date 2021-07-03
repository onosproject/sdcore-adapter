// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package main

import (
	"github.com/onosproject/sdcore-adapter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"strconv"
	"time"
)

func main() {

	collector.RecordMetrics(2*time.Second, "starbucks_newyork_cameras")
	collector.RecordMetrics(2*time.Second, "starbucks_seattle_cameras")
	collector.RecordMetrics(2*time.Second, "acme_chicago_robots")

	collector.RecordSmfMetrics(2*time.Second, "starbuck-newyork-cameras", []string{
		"170029313275040",
		"170029313275041",
		"170029313275050",
		"170029313275051",
		"170029313275052",
		"170029313275053",
		"170029313275054",
		"170029313275055",
	})

	var imsi uint64
	gameImsis := []string{}
	for imsi = 130029313275060; imsi <= 130029313275160; imsi++ {
		gameImsis = append(gameImsis, strconv.FormatUint(imsi, 10))
	}
	collector.RecordSmfMetrics(2*time.Second, "zynga-sfo-vrgames", gameImsis)

	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":2112", nil); err != nil {
		log.Fatal(err)
	}
}
