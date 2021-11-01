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

// RecordSiteMetrics records Site-based metrics
func RecordSiteMetrics(period time.Duration, siteID string) {
	go func() {
		for {
			isDisconnected := rand.Intn(100)%9 == 0
			isInMaintenance := rand.Intn(100)%8 == 0

			if isDisconnected {
				edgeTestsOk.WithLabelValues(siteID).Set(0)
				edgeTestsDown.WithLabelValues(siteID).Set(1)
			} else {
				edgeTestsOk.WithLabelValues(siteID).Set(1)
				edgeTestsDown.WithLabelValues(siteID).Set(0)
			}
			// This is separate condition as even in Maintenanace , Tests can be running and can be successful
			if isInMaintenance {
				edgeMaintenanceWindow.WithLabelValues(siteID).Set(1)
			} else {
				edgeMaintenanceWindow.WithLabelValues(siteID).Set(0)
			}
			time.Sleep(period)
		}
	}()
}

// RecordSmallCellMetrics records status of eNB at Site
func RecordSmallCellMetrics(period time.Duration, enbname string) {
	go func() {
		for {
			count := float64(rand.Intn(10))
			smallCellStatus.WithLabelValues("Active", enbname).Set(count)
			time.Sleep(period)
		}
	}()
}

// RecordMetrics records VCS-based metrics
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

// RecordUEMetrics reports UE-based metrics
func RecordUEMetrics(period time.Duration, vcsID string, imsiList []string, upThroughput float64, downThroughput float64, upLatency float64, downLatency float64) {
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
				counts[stateIndex]++
				for j, state := range states {
					if j == stateIndex {
						smfPduSessionProfile.WithLabelValues(imsi, ip, state, "upf", vcsID).Set(1)
					} else {
						smfPduSessionProfile.WithLabelValues(imsi, ip, state, "upf", vcsID).Set(0)

					}
				}

				if stateIndex == 0 {
					// active UE reports throughput and latency
					if PercentUpThroughput == nil {
						// randomize, between 75% and 100% of upThroughput argument
						ueThroughput.WithLabelValues(imsi, vcsID, "upstream").Set(upThroughput * float64(75+rand.Intn(25)) / 100.0)
					} else {
						// Someone turned the knob in the mock-sdcore-exporter control ui.
						ueThroughput.WithLabelValues(imsi, vcsID, "upstream").Set(upThroughput * (*PercentUpThroughput))
					}
					if PercentDownThroughput == nil {
						// randomize, between 75% and 100% of downThroughput argument
						ueThroughput.WithLabelValues(imsi, vcsID, "downstream").Set(downThroughput * float64(75+rand.Intn(25)) / 100.0)
					} else {
						// Someone turned the knob in the mock-sdcore-exporter control ui.
						ueThroughput.WithLabelValues(imsi, vcsID, "downstream").Set(downThroughput * (*PercentDownThroughput))
					}
					if PercentUpLatency == nil {
						// randomize, between 75% and 100% of latency argument
						ueLatency.WithLabelValues(imsi, vcsID, "upstream").Set(upLatency * float64(75+rand.Intn(25)) / 100.0)
					} else {
						// Someone turned the knob in the mock-sdcore-exporter control ui.
						ueLatency.WithLabelValues(imsi, vcsID, "upstream").Set(upLatency * (*PercentUpLatency))
					}
					if PercentDownLatency == nil {
						// randomize, between 75% and 100% of latency argument
						ueLatency.WithLabelValues(imsi, vcsID, "downstream").Set(downLatency * float64(75+rand.Intn(25)) / 100.0)
					} else {
						// Someone turned the knob in the mock-sdcore-exporter control ui.
						ueLatency.WithLabelValues(imsi, vcsID, "downstream").Set(downLatency * (*PercentDownLatency))
					}
					ueSubscriberInfo.WithLabelValues(imsi, ip).Set(1)
				} else {
					// inactive UE has no throughput or latency
					ueThroughput.WithLabelValues(imsi, vcsID, "upstream").Set(0)
					ueThroughput.WithLabelValues(imsi, vcsID, "downstream").Set(0)
					ueLatency.WithLabelValues(imsi, vcsID, "upstream").Set(0)
					ueLatency.WithLabelValues(imsi, vcsID, "downstream").Set(0)
					ueSubscriberInfo.WithLabelValues(imsi, ip).Set(0)
				}
			}
			smfPduSessions.Set(counts[0]) // counts[0] is active UEs
			time.Sleep(period)
		}
	}()
}

var (
	// Site metrics
	edgeTestsOk = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "aetheredge_e2e_tests_ok",
		Help: "Edge Connectivity Tests",
	}, []string{"name"})
	edgeTestsDown = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "aetheredge_e2e_tests_down",
		Help: "Edge Connectivity Tests",
	}, []string{"name"})
	edgeMaintenanceWindow = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "aetheredge_in_maintenance_window",
		Help: "Edge Site in Maintenance",
	}, []string{"name"})
	smallCellStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mme_number_of_enb_attached",
		Help: "ENB Status",
	}, []string{"enb_state", "enbname"})
	//VCS-based metrics
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
	ueSubscriberInfo = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "subscriber_info",
		Help: "subscriber info",
	}, []string{"imsi", "ip"})
	ueThroughput = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ue_throughput",
		Help: "ue_throughput",
	}, []string{"id", "slice", "direction"})
	ueLatency = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ue_latency",
		Help: "ue_latency",
	}, []string{"id", "slice", "direction"})
)
