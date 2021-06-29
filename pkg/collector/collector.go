// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package collector

import (
	"fmt"
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
			vcsLatency.WithLabelValues(vcdID).Add(float64(rand.Intn(10)-5) / 5.0)
			vcsJitter.WithLabelValues(vcdID).Add(float64(rand.Intn(10)-5) / 100.0)
			vcsThroughput.WithLabelValues(vcdID).Add(float64(rand.Intn(10)-5) * 10)
			time.Sleep(period)
		}
	}()
}

func RecordSmfMetrics(period time.Duration, vcsId string, imsiList []string) {
	go func() {
		states := []string{"active", "inactive", "idle"}
		for {
			counts := []float64{0, 0, 0}
			for i, imsi := range imsiList {
				ip := fmt.Sprintf("1.2.3.%d", i)

				// For a given UE, it can be in either "active", "inactive", or "idle" state. It
				// can only be in one state at a time.
				//
				// Pick a state, set its metric to 1, leave the other metrics at 0.

				stateIndex := rand.Int() % 3
				counts[stateIndex] += 1
				for j, state := range states {
					if j == stateIndex {
						smfPduSessionProfile.WithLabelValues(imsi, ip, state, "upf", vcsId).Set(1)
					} else {
						smfPduSessionProfile.WithLabelValues(imsi, ip, state, "upf", vcsId).Set(0)
					}
				}
			}
			smfPduSessions.Set(counts[0]) // counts[0] are active UEs
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
	smfPduSessionProfile = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "smf_pdu_session_profile",
		Help: "smf pdu session profile",
	}, []string{"id", "ip", "state", "upf", "slice"})
	smfPduSessions = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "smf_pdu_sessions",
		Help: "smf pdu session count",
	})
)
