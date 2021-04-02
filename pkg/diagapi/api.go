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
 */

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"github.com/onosproject/sdcore-adapter/pkg/migration"
)

var log = logging.GetLogger("diagapi")

type TargetInterface interface {
	ExecuteCallbacks(reason gnmi.ConfigCallbackType) error
	GetJSON() ([]byte, error)
	PutJSON([]byte) error
}

type DiagnosticApi struct {
	targetServer            TargetInterface
	defaultTarget           string
	defaultAetherConfigAddr string
}

func (m *DiagnosticApi) reSync(w http.ResponseWriter, r *http.Request) {
	// TODO: tell the target server to synchronize
	_ = r
	m.targetServer.ExecuteCallbacks(gnmi.Forced)
	fmt.Fprintf(w, "Okay")
}

func (m *DiagnosticApi) getCache(w http.ResponseWriter, r *http.Request) {
	_ = r
	jsonDump, err := m.targetServer.GetJSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonDump)
}

func (m *DiagnosticApi) deleteCache(w http.ResponseWriter, r *http.Request) {
	reqBody := []byte("{}")
	err := m.targetServer.PutJSON(reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "SUCCESS")
}

func (m *DiagnosticApi) postCache(w http.ResponseWriter, r *http.Request) {
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

func (m *DiagnosticApi) pullFromOnosConfig(w http.ResponseWriter, r *http.Request) {
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

	srcVal, err := migration.GetPath("", target, aetherConfigAddr, context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	srcJsonBytes := srcVal.GetJsonVal()

	err = m.targetServer.PutJSON(srcJsonBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "SUCCESS")
}

func (m *DiagnosticApi) handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/synchronize", m.reSync).Methods("POST")
	myRouter.HandleFunc("/cache", m.getCache).Methods("GET")
	myRouter.HandleFunc("/cache", m.postCache).Methods("POST")
	myRouter.HandleFunc("/cache", m.deleteCache).Methods("DELETE")
	myRouter.HandleFunc("/pull", m.pullFromOnosConfig).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", myRouter))
}

func StartDiagnosticAPI(targetServer TargetInterface,
	defaultAetherConfigAddr string,
	defaultTarget string) {
	m := DiagnosticApi{targetServer: targetServer,
		defaultAetherConfigAddr: defaultAetherConfigAddr,
		defaultTarget:           defaultTarget}
	go m.handleRequests()
}
