// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"math/rand"
	"time"
)


func RecordMetrics(period time.Duration, vcdID string) {
	vcsLatency.WithLabelValues(vcdID).Set(21.0)
	vcsJitter.WithLabelValues(vcdID).Set(3.0)
	vcsThroughput.WithLabelValues(vcdID).Set(10000)

	go func() {
		for {
			vcsLatency.WithLabelValues(vcdID).Add(float64(rand.Intn(10) - 5)/5.0)
			vcsJitter.WithLabelValues(vcdID).Add(float64(rand.Intn(10) - 5)/100.0)
			vcsThroughput.WithLabelValues(vcdID).Add(float64(rand.Intn(10) - 5)*10)
			time.Sleep(period)
		}
	}()
}

var (
	vcsLatency = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vcs_latency",
		Help: "VCS Latency",
	}, []string{"vcs_id"})
	vcsJitter = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vcs_jitter",
		Help: "VCS Jitter",
	}, []string{"vcs_id"})
	vcsThroughput = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vcs_throughput",
		Help: "VCS Throughput",
	}, []string{"vcs_id"})
)
