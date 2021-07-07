// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

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

var (
	PercentActiveSubscribers *float64 // Nil if unset, otherwise value set from form
)

type ExporterApi struct {
}

// Serves up the index.html page that contains the knob. For such a short page, it's
// easy enough to put the page contents inline and simplifiy distribution. If the page
// becomes more complex, then consider putting it in a separate file.
func (m *ExporterApi) index(w http.ResponseWriter, r *http.Request) {
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
	<input type="text" data-angleoffset=-125 data-anglearc=250 data-fgcolor="#66EE66" value="50" class="dial">
	
	<script>
		$(".dial").knob({
		'release' : function (sendpostresp) {
			$.ajax({
				url: "/postActiveSubscribers",
				type: "POST",
				data: { 
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
func (m *ExporterApi) postActiveSubscribers(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Warnf("ParseForm() err: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Infof("Post from website: r.PostFrom = %v", r.PostForm)

	value := r.FormValue("value")
	if value == "" {
		return
	}

	f64, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Warnf("PostForm() ParseFloat Err %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	PercentActiveSubscribers = &f64

	log.Infof("Set Active Subscribers to %f", *PercentActiveSubscribers)
}

func (m *ExporterApi) handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/index.html", m.index).Methods("GET")
	myRouter.HandleFunc("/postActiveSubscribers", m.postActiveSubscribers).Methods("POST")
	err := http.ListenAndServe(":8081", myRouter)
	if err != nil {
		log.Panicf("ExporterAPI Error %v", err)
	}
}

func StartExporterAPI() {
	m := ExporterApi{}
	go m.handleRequests()
}
