// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/closedloop"
)

// Example:
// sample-rocapp -client_key=/etc/sdcore-adapter/certs/tls.key -client_crt=/etc/sdcore-adapter/certs/tls.crt -ca_crt=/etc/sdcore-adapter/certs/tls.cacert -hostCheckDisabled

// If you want to forward the prometheus, for manual inspection
// kubectl -n micro-onos port-forward services/aether-roc-umbrella-prometheus-server --address 0.0.0.0 8180:80

var (
	aetherConfigAddr   = flag.String("aether_config_addr", "", "If specified, pull initial state from aether-config at this address")
	aetherConfigTarget = flag.String("aether_config_target", "connectivity-service-v2", "Target to use when pulling from aether-config")
	sliceName          = flag.String("slice_name", "starbuck-newyork-cameras", "Target to use when pulling from aether-config")
)

var log = logging.GetLogger("sample-rocapp")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	conf := &closedloop.ClosedLoopConfig{}
	err := conf.LoadFromYamlFile("/etc/sample-rocapp.yaml") // TODO: Config file name is hardcoded; consider adding a command-line parser
	if err != nil {
		panic(err)
	}

	log.Infof("Loaded Config %+v", conf)

	control := closedloop.NewClosedLoopControl(conf)

	for {
		err = control.Evaluate()
		if err != nil {
			log.Errorf("Error: %v", err)
		}
		time.Sleep(5 * time.Second)
	}
}
