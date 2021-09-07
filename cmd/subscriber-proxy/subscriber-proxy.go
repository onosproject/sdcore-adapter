// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0
package main

import (
	"flag"
	"fmt"
	"github.com/onosproject/sdcore-adapter/pkg/gnmiclient"
	"github.com/onosproject/sdcore-adapter/pkg/subproxy"
	"github.com/prometheus/common/log"
	"os"
	"time"
)

var (
	bindPort           = flag.String("bind_port", ":5001", "Bind to just :port")
	postTimeout        = flag.Duration("post_timeout", time.Second*10, "Timeout duration when making post requests")
	aetherConfigTarget = flag.String("aether_config_target", "connectivity-service-v3", "Target to use when pulling from aether-config")
	baseWebConsoleURL  = flag.String("webconsole_url", "http://webui.omec.svc.cluster.local:5000", "base url for webui service address")
	aetherConfigAddr   = flag.String("onos_config_url", "onos-config.micro-onos.svc.cluster.local:5150", "url of onos-config")
)

// Main
func main() {
	flag.Usage = func() {
		_, err := fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		if err != nil {
			log.Info("fail to parse the flags")
			return
		}
		flag.PrintDefaults()
	}
	flag.Parse()
	gnmiClient, err := gnmiclient.NewGnmi(*aetherConfigAddr, time.Second*15)
	if err != nil {
		log.Fatalf("Error opening gNMI client %s", err.Error())
		return
	}
	defer gnmiClient.CloseClient()

	fmt.Println("debug 1", gnmiClient)

	proxy := subproxy.NewSubscriberProxy(*aetherConfigTarget, *baseWebConsoleURL, *aetherConfigAddr, gnmiClient, *postTimeout)

	proxy.StartSubscriberProxy(*bindPort, "/api/subscriber/:ueId")
}
