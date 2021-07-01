// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/onosproject/sdcore-adapter/pkg/closedloop"

	"github.com/onosproject/onos-lib-go/pkg/logging"
)

// kubectl -n micro-onos port-forward services/aether-roc-umbrella-prometheus-server --address 0.0.0.0 8180:80

// smf_pdu_session_profile{state="active",slice="starbucks_newyork_cameras"}==1 ... list all UEs that are active
// sum by (state) (smf_pdu_session_profile{slice="starbucks_newyork_cameras"})  ... coun the active, idle, inactive cameras
//

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

	/*
		mf, err := metrics.NewMetricsFetcher(*metricAddr)
		if err != nil {
			panic(err)
		}

		m, err := mf.GetSingleVector("SUM(smf_pdu_session_profile{slice=\"starbucks_newyork_cameras\",state=\"active\"}) <= 3")
		if err != nil {
			panic(err)
		}
		if m == nil {
			fmt.Print("nil\n")
		} else {
			fmt.Printf("%v\n", *m)
		}

		m, err = mf.GetSingleVector("SUM(smf_pdu_session_profile{slice=\"starbucks_newyork_cameras\",state=\"active\"}) > 3")
		if err != nil {
			panic(err)
		}
		if m == nil {
			fmt.Print("nil\n")
		} else {
			fmt.Printf("%v\n", *m)
		}

		um, err := mf.GetSliceUEMetrics("starbucks_newyork_cameras")
		if err != nil {
			panic(err)
		}
		fmt.Printf("%v\n", um)
	*/

	conf := closedloop.ClosedLoopConfig{}
	err := conf.LoadFromYamlFile("/etc/sample-rocapp.yaml")
	if err != nil {
		panic(err)
	}

	control := closedloop.NewClosedLoopControl(conf)

	fmt.Printf("%+v\b", conf)

	/*

		// Optional: pull initial config from onos-config
		if *aetherConfigAddr != "" {
			log.Infof("Fetching initial state from %s, target %s", *aetherConfigAddr, *aetherConfigTarget)
			// The migration library has the functions for fetching from onos-config
			srcVal, err := migration.GetPath("", *aetherConfigTarget, *aetherConfigAddr, context.Background())
			if err != nil {
				log.Fatalf("Error fetching initial data from onos-config: %s", err.Error())
				return
			}

			configData = srcVal.GetJsonVal()

			log.Infof("Fetched config: %s", string(configData))
		}*/
}
