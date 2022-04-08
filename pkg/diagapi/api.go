// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

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
 *
 *   # change the synchronizer log level
 *   curl -v -X POST http://localhost:8080/loglevel/root --data "DEBUG"
 */

import (
	"context"
	"fmt"
	"github.com/onosproject/sdcore-adapter/pkg/gnmiclient"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

var log = logging.GetLogger("diagapi")

// TargetInterface is an interface to a gNMI Target
type TargetInterface interface {
	ExecuteCallbacks(reason gnmi.ConfigCallbackType, target string, path *pb.Path) error
	GetJSON(string) ([]byte, error)
	PutJSON(string, []byte) error
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
	err := m.targetServer.ExecuteCallbacks(gnmi.Forced, gnmi.AllTargets, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "SUCCESS")
}

func (m *DiagnosticAPI) getCache(w http.ResponseWriter, r *http.Request) {
	queryArgs := r.URL.Query()

	target := queryArgs.Get("target")
	if target == "" {
		target = m.defaultTarget
	}

	jsonDump, err := m.targetServer.GetJSON(target)
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
	queryArgs := r.URL.Query()

	target := queryArgs.Get("target")
	if target == "" {
		target = m.defaultTarget
	}

	reqBody := []byte("{}")
	err := m.targetServer.PutJSON(target, reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "SUCCESS")
}

func (m *DiagnosticAPI) postCache(w http.ResponseWriter, r *http.Request) {
	queryArgs := r.URL.Query()

	target := queryArgs.Get("target")
	if target == "" {
		target = m.defaultTarget
	}

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = m.targetServer.PutJSON(target, reqBody)
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

	err = m.targetServer.PutJSON(target, srcJSONBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "SUCCESS")
}

// this method is not exported in onos logger
func splitLoggerName(name string) []string {
	names := strings.Split(name, "/")
	return names
}

// this method is not exported in onos logger
func levelStringToLevel(l string) (logging.Level, error) {
	switch strings.ToUpper(l) {
	case logging.DebugLevel.String():
		return logging.DebugLevel, nil
	case logging.InfoLevel.String():
		return logging.InfoLevel, nil
	case logging.WarnLevel.String():
		return logging.WarnLevel, nil
	case logging.ErrorLevel.String():
		return logging.ErrorLevel, nil
	case logging.FatalLevel.String():
		return logging.FatalLevel, nil
	case logging.PanicLevel.String():
		return logging.PanicLevel, nil
	case logging.DPanicLevel.String():
		return logging.DPanicLevel, nil
	}
	return logging.ErrorLevel, fmt.Errorf("Unknown level %s", l)
}

func (m *DiagnosticAPI) getLogLevel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["logger"]
	names := splitLoggerName(name)
	logger := logging.GetLogger(names...)

	_, err := w.Write([]byte(logger.GetLevel().String()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (m *DiagnosticAPI) setLogLevel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["logger"]
	names := splitLoggerName(name)
	logger := logging.GetLogger(names...)

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	level, err := levelStringToLevel(string(reqBody))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.SetLevel(level)

	_, err = w.Write([]byte(logger.GetLevel().String()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (m *DiagnosticAPI) handleRequests(port uint) {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/synchronize", m.reSync).Methods("POST")
	myRouter.HandleFunc("/cache", m.getCache).Methods("GET")
	myRouter.HandleFunc("/cache", m.postCache).Methods("POST")
	myRouter.HandleFunc("/cache", m.deleteCache).Methods("DELETE")
	myRouter.HandleFunc("/pull", m.pullFromOnosConfig).Methods("POST")
	myRouter.HandleFunc("/loglevel/{logger}", m.getLogLevel).Methods("GET")
	myRouter.HandleFunc("/loglevel/{logger}", m.setLogLevel).Methods("POST")
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
