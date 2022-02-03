// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package collector

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

// Give mock-sdcore-exporter some knobs that can be manually turned.
//
// NOTE: Turning a knob currently affects all VCS, but we could address that
// with a drop-down. There's also no mechanism to set the initial value of the
// knobs. This is intended to be a simple interface to facilitate demos.
//
// If a value is nil,  then it has not been set via the API, and the collector
// will randomize it.

var (
	// PercentActiveSubscribers sets the percentage of active subscribers
	PercentActiveSubscribers *float64

	// PercentUpThroughput sets the percentage of upstream throughput
	PercentUpThroughput *float64

	// PercentDownThroughput sets the percentage of downstream throughput
	PercentDownThroughput *float64

	// PercentUpLatency sets the percentage of upstream latency
	PercentUpLatency *float64

	// PercentDownLatency sets the percentage of downstream latency
	PercentDownLatency *float64
)

// ExporterAPI is an API for remote configuring the mock pusher.
type ExporterAPI struct {
}

// Serves up the index.html page that contains the knob. For such a short page, it's
// easy enough to put the page contents inline and simplify distribution. If the page
// becomes more complex, then consider putting it in a separate file.
func (m *ExporterAPI) index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
	<!DOCTYPE html>
	<html lang="en">
	
	<head>
	<meta charset="utf-8" />
	<script src="http://ajax.googleapis.com/ajax/libs/jquery/1.9.1/jquery.min.js"></script>
	<script src="https://cdnjs.cloudflare.com/ajax/libs/jQuery-Knob/1.2.13/jquery.knob.min.js"></script>
	</head>
	
	<body>
	<p>Percent Active Subscribers</p>
	<input type="text" data-angleoffset=-125 data-anglearc=250 data-fgcolor="#66EE66" value="50" class="activeDial">
	<p>Percent Upstream Throughput</p>
	<input type="text" data-angleoffset=-125 data-anglearc=250 data-fgcolor="#66EE66" value="50" class="upThroughputDial">
	<p>Percent Downstream Throughput</p>
	<input type="text" data-angleoffset=-125 data-anglearc=250 data-fgcolor="#66EE66" value="50" class="downThroughputDial">
	
	<script>
		$(".activeDial").knob({
			'release' : function (sendpostresp) {
				$.ajax({
					url: "/updateKnob",
					type: "POST",
					data: {
						knob: "active",
						value: sendpostresp 
					}		
				});
			}
		});
		
		$(".upThroughputDial").knob({
			'release' : function (sendpostresp) {
				$.ajax({
					url: "/updateKnob",
					type: "POST",
					data: {
						knob: "upThroughput",
						value: sendpostresp 
					}		
				});
			}
		});

		$(".downThroughputDial").knob({
			'release' : function (sendpostresp) {
				$.ajax({
					url: "/updateKnob",
					type: "POST",
					data: {
						knob: "downThroughput",
						value: sendpostresp 
					}		
				});
			}
		});			
	</script>
	
	</body>
	</html>	
	`)
}

// The Knob posts to this endpoint.
func (m *ExporterAPI) postUpdateKnob(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Warnf("ParseForm() err: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Infof("Post from website: r.PostFrom = %v", r.PostForm)

	value := r.FormValue("value")
	if value == "" {
		log.Warn("No value")
		return
	}

	knob := r.FormValue("knob")
	if knob == "" {
		log.Warn("No knob")
		return
	}

	f64, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Warnf("PostForm() ParseFloat Err %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch knob {
	case "active":
		PercentActiveSubscribers = &f64
	case "upThroughput":
		PercentUpThroughput = &f64
	case "downThroughput":
		PercentDownThroughput = &f64
	case "upLatency":
		PercentUpLatency = &f64
	case "downLatency":
		PercentUpThroughput = &f64
	default:
		log.Warnf("Unknown knob %s", knob)
		http.Error(w, "unknown knob", http.StatusInternalServerError)
	}

	log.Infof("Set Knob %s to %f", knob, value)
}

func (m *ExporterAPI) handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", m.index).Methods("GET")
	myRouter.HandleFunc("/updateKnob", m.postUpdateKnob).Methods("POST")
	err := http.ListenAndServe(":8081", myRouter)
	if err != nil {
		log.Panicf("ExporterAPI Error %v", err)
	}
}

// StartExporterAPI starts the exporter API, serving requests.
func StartExporterAPI() {
	m := ExporterAPI{}
	go m.handleRequests()
}
