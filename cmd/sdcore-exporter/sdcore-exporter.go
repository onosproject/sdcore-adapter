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

	collector.RecordMetrics(2*time.Second, "starbucks-newyork-cameras")
	collector.RecordMetrics(2*time.Second, "starbucks-seattle-cameras")
	collector.RecordMetrics(2*time.Second, "acme-chicago-robots")

	// period, vcs, imsis, upMbps, downMbps, upLatency, downLatency
	collector.RecordUEMetrics(2*time.Second, "starbucks-newyork-cameras", []string{
		"170029313275040",
		"170029313275041",
		"170029313275050",
		"170029313275051",
		"170029313275052",
		"170029313275053",
		"170029313275054",
		"170029313275055",
	}, 100000, 1000, 10, 10)

	collector.RecordUEMetrics(2*time.Second, "starbucks-seattle-cameras", []string{
		"170029313275000",
		"170029313275001",
		"170029313275002",
		"170029313275003",
		"170029313275010",
		"170029313275011",
		"170029313275012",
		"170029313275013",
		"170029313275014",
	}, 100000, 1000, 10, 10)

	collector.RecordUEMetrics(2*time.Second, "starbucks-seattle-pos", []string{
		"170029313275020",
		"170029313275021",
		"170029313275022",
		"170029313275030",
		"170029313275031",
		"170029313275032",
		"170029313275033",
		"170029313275034",
	}, 100000, 1000, 10, 10)

	collector.RecordUEMetrics(2*time.Second, "starbucks-newyork-pos", []string{
		"170029313275060",
		"170029313275061",
		"170029313275070",
		"170029313275071",
		"170029313275072",
		"170029313275073",
	}, 100000, 1000, 10, 10)

	collector.RecordUEMetrics(2*time.Second, "acme-chicago-robots", []string{
		"13698808332993000",
		"13698808332993001",
		"13698808332993002",
		"13698808332993003",
		"13698808332993010",
		"13698808332993011",
		"13698808332993012",
	}, 100000, 1000, 10, 10)

	var imsi uint64
	gameImsis := []string{}
	for imsi = 130029313275060; imsi <= 130029313275160; imsi++ {
		gameImsis = append(gameImsis, strconv.FormatUint(imsi, 10))
	}
	collector.RecordUEMetrics(2*time.Second, "zynga-sfo-vrgames", gameImsis, 1000, 100000, 10, 10)

	collector.StartExporterAPI()

	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":2112", nil); err != nil {
		log.Fatal(err)
	}
}
