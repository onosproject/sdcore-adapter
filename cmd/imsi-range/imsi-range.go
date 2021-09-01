// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0
package main

import (
	"flag"
	"fmt"
	"github.com/onosproject/sdcore-adapter/pkg/gnmiclient"
	"github.com/prometheus/common/log"
	"os"
	"time"

	"github.com/onosproject/sdcore-adapter/pkg/imsirange"
)

var (
	aetherConfigTarget = flag.String("aether_config_target", "connectivity-service-v3",
		"Target to use when pulling from aether-config")
	aetherConfigAddr = flag.String("onos_config_url", "onos-config.micro-onos.svc.cluster.local:5150",
		"url of onos-config")
)

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

	gnmiClient, err := gnmiclient.NewGnmi(*aetherConfigAddr, time.Second*5)
	if err != nil {
		log.Fatalf("Error opening gNMI client %s", err.Error())
	}
	defer gnmiClient.CloseClient()

	ims := imsirange.NewIMSIRange(*aetherConfigAddr, *aetherConfigTarget)

	device, err := ims.GetDevice(gnmiClient)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = ims.CollapseImsi(device, gnmiClient)
	if err != nil {
		log.Error(err.Error())
	}
}
