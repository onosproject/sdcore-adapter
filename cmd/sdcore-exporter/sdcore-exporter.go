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
	collector.RecordSiteMetrics(2*time.Second, "acme-chicago")
	collector.RecordSiteMetrics(2*time.Second, "starbucks-seattle")
	collector.RecordSiteMetrics(2*time.Second, "starbucks-newyork")

	collector.RecordMetrics(2*time.Second, "starbucks-newyork-cameras")
	collector.RecordMetrics(2*time.Second, "starbucks-seattle-cameras")
	collector.RecordMetrics(2*time.Second, "acme-chicago-robots")

	// period, vcs, imsis, upMbps, downMbps, upLatency, downLatency
	collector.RecordUEMetrics(2*time.Second, "starbucks-newyork-cameras", []string{
		"40",
		"41",
		"50",
		"51",
		"52",
		"53",
		"54",
		"55",
	}, 100000, 1000, 10, 10)

	collector.RecordUEMetrics(2*time.Second, "starbucks-seattle-cameras", []string{
		"00",
		"01",
		"02",
		"03",
		"10",
		"11",
		"12",
		"13",
		"14",
	}, 100000, 1000, 10, 10)

	collector.RecordUEMetrics(2*time.Second, "starbucks-seattle-pos", []string{
		"20",
		"21",
		"22",
		"30",
		"31",
		"32",
		"33",
		"34",
	}, 100000, 1000, 10, 10)

	collector.RecordUEMetrics(2*time.Second, "starbucks-newyork-pos", []string{
		"60",
		"61",
		"70",
		"71",
		"72",
		"73",
	}, 100000, 1000, 10, 10)

	collector.RecordUEMetrics(2*time.Second, "acme-chicago-robots", []string{
		"0",
		"1",
		"2",
		"3",
		"10",
		"11",
		"12",
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
