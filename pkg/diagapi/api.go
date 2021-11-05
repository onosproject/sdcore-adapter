// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package diagapi

/*
 * api.go: an out-of-band diagnostic API
 *
 * Examples:
 *   # dump the current cache to a file
 *   curl http://localhost:8080/cache > state.json
 *
 *   # delete the contents of the cache
 *   curl --header "Content-Type: application/json" -X DELETE http://localhost:8080/cache
 *
 *   # load the cache from a file
 *   curl --header "Content-Type: application/json" -X POST --data @state.json http://localhost:8080/cache
 *
 *   # pull state from onos-config and populate the cache
 *   curl --header "Content-Type: application/json" -X POST "http://localhost:8080/pull?target=connectivity-service-v2&aetherConfigAddr=onos-config:5150"
 *
 *   # pull state from onos-config and populate the cache, with auth token
 *   AUTH=<...stuff...>
 *   curl --header "Content-Type: application/json" --header "Authorization: Bearer $AUTH" -X POST "http://localhost:8080/pull?target=connectivity-service-v2&aetherConfigAddr=onos-config:5150"
 */

import (
	"context"
	"fmt"
	"github.com/onosproject/sdcore-adapter/pkg/gnmiclient"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

var log = logging.GetLogger("diagapi")

// TargetInterface is an interface to a gNMI Target
type TargetInterface interface {
	ExecuteCallbacks(reason gnmi.ConfigCallbackType, path *pb.Path) error
	GetJSON() ([]byte, error)
	PutJSON([]byte) error
}

// DiagnosticAPI is an api for performing diagnostic operations on the synchronizer
type DiagnosticAPI struct {
	targetServer            TargetInterface
	defaultTarget           string
	defaultAetherConfigAddr string
}

func (m *DiagnosticAPI) reSync(w http.ResponseWriter, r *http.Request) {
	// TODO: tell the target server to synchronize
	_ = r
	err := m.targetServer.ExecuteCallbacks(gnmi.Forced, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "SUCCESS")
}

func (m *DiagnosticAPI) getCache(w http.ResponseWriter, r *http.Request) {
	_ = r
	jsonDump, err := m.targetServer.GetJSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonDump)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (m *DiagnosticAPI) deleteCache(w http.ResponseWriter, r *http.Request) {
	reqBody := []byte("{}")
	err := m.targetServer.PutJSON(reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "SUCCESS")
}

func (m *DiagnosticAPI) postCache(w http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = m.targetServer.PutJSON(reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "SUCCESS")
}

func (m *DiagnosticAPI) pullFromOnosConfig(w http.ResponseWriter, r *http.Request) {
	queryArgs := r.URL.Query()

	target := queryArgs.Get("target")
	if target == "" {
		target = m.defaultTarget
	}

	aetherConfigAddr := queryArgs.Get("aetherConfigAddr")
	if aetherConfigAddr == "" {
		aetherConfigAddr = m.defaultAetherConfigAddr
	}

	log.Infof("Pull, aetherConfig=%s, target=%s", aetherConfigAddr, target)

	ctx := context.Background()
	if auth := r.Header.Get("Authorization"); auth != "" {
		ctx = gnmiclient.WithAuthorization(ctx, auth)
	}

	srcVal, err := gnmiclient.GetPath(ctx, "", target, aetherConfigAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	srcJSONBytes := srcVal.GetJsonVal()

	err = m.targetServer.PutJSON(srcJSONBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "SUCCESS")
}

func (m *DiagnosticAPI) handleRequests(port uint) {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/synchronize", m.reSync).Methods("POST")
	myRouter.HandleFunc("/cache", m.getCache).Methods("GET")
	myRouter.HandleFunc("/cache", m.postCache).Methods("POST")
	myRouter.HandleFunc("/cache", m.deleteCache).Methods("DELETE")
	myRouter.HandleFunc("/pull", m.pullFromOnosConfig).Methods("POST")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), myRouter))
}

// StartDiagnosticAPI starts the Diagnostic API, serving requests
func StartDiagnosticAPI(targetServer TargetInterface,
	defaultAetherConfigAddr string,
	defaultTarget string,
	port uint) {
	m := DiagnosticAPI{targetServer: targetServer,
		defaultAetherConfigAddr: defaultAetherConfigAddr,
		defaultTarget:           defaultTarget}
	go m.handleRequests(port)
}
