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

	"github.com/onosproject/onos-lib-go/pkg/logging"
)

var log = logging.GetLogger("collector")

func RecordMetrics(period time.Duration, vcdID string) {
	vcsLatency.WithLabelValues(vcdID).Set(21.0)
	vcsJitter.WithLabelValues(vcdID).Set(3.0)
	vcsThroughput.WithLabelValues(vcdID).Set(10000)
	rand.Seed(time.Now().UnixNano())

	go func() {
		for {
			vcsLatency.WithLabelValues(vcdID).Add(float64(rand.Intn(11)-5) / 5.0)
			vcsJitter.WithLabelValues(vcdID).Add(float64(rand.Intn(11)-5) / 100.0)
			vcsThroughput.WithLabelValues(vcdID).Add(float64(rand.Intn(11)-5) * 10)
			time.Sleep(period)
		}
	}()
}

func RecordUEMetrics(period time.Duration, vcsId string, imsiList []string, downThroughput float64, upThroughput float64, upLatency float64, downLatency float64) {
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

				var stateIndex int
				if PercentActiveSubscribers == nil {
					// Randomize
					stateIndex = rand.Int() % 3
				} else {
					// Someone turned the knob in the mock-sdcore-exporter control ui.
					// PercentActiveSubscribers is from api.go
					thresh := float64(len(imsiList)) * (*PercentActiveSubscribers) / 100.0
					if float64(i) <= thresh {
						// active
						stateIndex = 0
					} else {
						// inactive
						stateIndex = 1

					}

				}
				counts[stateIndex] += 1
				for j, state := range states {
					if j == stateIndex {
						smfPduSessionProfile.WithLabelValues(imsi, ip, state, "upf", vcsId).Set(1)
					} else {
						smfPduSessionProfile.WithLabelValues(imsi, ip, state, "upf", vcsId).Set(0)
					}
				}

				if stateIndex == 0 {
					// active UE reports throughput and latency
					if PercentUpThroughput == nil {
						// randomize, between 75% and 100% of upThroughput argument
						ueThroughput.WithLabelValues(imsi, vcsId, "upstream").Set(upThroughput * float64(75+rand.Intn(25)) / 100.0)
					} else {
						// Someone turned the knob in the mock-sdcore-exporter control ui.
						ueThroughput.WithLabelValues(imsi, vcsId, "upstream").Set(upThroughput * (*PercentUpThroughput))
					}
					if PercentDownThroughput == nil {
						// randomize, between 75% and 100% of downThroughput argument
						ueThroughput.WithLabelValues(imsi, vcsId, "upstream").Set(downThroughput * float64(75+rand.Intn(25)) / 100.0)
					} else {
						// Someone turned the knob in the mock-sdcore-exporter control ui.
						ueThroughput.WithLabelValues(imsi, vcsId, "upstream").Set(downThroughput * (*PercentDownThroughput))
					}
					if PercentUpLatency == nil {
						// randomize, between 75% and 100% of latency argument
						ueLatency.WithLabelValues(imsi, vcsId, "upstream").Set(upLatency * float64(75+rand.Intn(25)) / 100.0)
					} else {
						// Someone turned the knob in the mock-sdcore-exporter control ui.
						ueLatency.WithLabelValues(imsi, vcsId, "upstream").Set(upLatency * (*PercentUpLatency))
					}
					if PercentDownLatency == nil {
						// randomize, between 75% and 100% of latency argument
						ueLatency.WithLabelValues(imsi, vcsId, "downstream").Set(downLatency * float64(75+rand.Intn(25)) / 100.0)
					} else {
						// Someone turned the knob in the mock-sdcore-exporter control ui.
						ueLatency.WithLabelValues(imsi, vcsId, "downstream").Set(downLatency * (*PercentDownLatency))
					}
				} else {
					// inactive UE has no throughput or latency
					ueThroughput.WithLabelValues(imsi, vcsId, "upstream").Set(0)
					ueThroughput.WithLabelValues(imsi, vcsId, "downstream").Set(0)
					ueLatency.WithLabelValues(imsi, vcsId, "upstream").Set(0)
					ueLatency.WithLabelValues(imsi, vcsId, "downstream").Set(0)
				}
			}
			smfPduSessions.Set(counts[0]) // counts[0] is active UEs
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

	// SMF metrics are actual metrics from SD-Core
	smfPduSessionProfile = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "smf_pdu_session_profile",
		Help: "smf pdu session profile",
	}, []string{"id", "ip", "state", "upf", "slice"})
	smfPduSessions = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "smf_pdu_sessions",
		Help: "smf pdu session count",
	})

	// UE throughput and latencies are hypothetical per-UE values
	ueThroughput = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ue_throughput",
		Help: "ue_throughput",
	}, []string{"id", "slice", "direction"})
	ueLatency = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ue_latency",
		Help: "ue_latency",
	}, []string{"id", "slice", "direction"})
)
