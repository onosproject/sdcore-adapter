// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package collector

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type ExporterApi struct {
}

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
					foo: sendpostresp 
				}		
			});
		}
	});
	</script>
	
	</body>
	</html>	
	`)
}

func (m *ExporterApi) postActiveSubscribers(w http.ResponseWriter, r *http.Request) {
	queryArgs := r.URL.Query()

	/*target := queryArgs.Get("target")
	if target == "" {
		target = m.defaultTarget
	}*/

	_ = queryArgs

	fmt.Fprintf(w, "SUCCESS")
}

func (m *ExporterApi) handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/index.html", m.index).Methods("GET")
	myRouter.HandleFunc("/postActiveSubscribers", m.postActiveSubscribers).Methods("POST")
	log.Fatal(http.ListenAndServe(":8081", myRouter))
}

func StartExporterAPI() {
	m := ExporterApi{}
	go m.handleRequests()
}
