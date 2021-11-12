// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/onosproject/sdcore-adapter/pkg/gnmiclient"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/google/gnxi/utils/credentials"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/diagapi"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	synchronizer "github.com/onosproject/sdcore-adapter/pkg/synchronizer"
	synchronizerv4 "github.com/onosproject/sdcore-adapter/pkg/synchronizer/v4"
	"github.com/onosproject/sdcore-adapter/pkg/target"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	bindAddr           = flag.String("bind_address", ":10161", "Bind to address:port or just :port")
	metricAddr         = flag.String("metric_address", ":9851", "Prometheus metric endpoint bind to address:port or just :port")
	configFile         = flag.String("config", "", "IETF JSON file for target startup config")
	outputFileName     = flag.String("output", "", "JSON file to save output to")
	_                  = flag.String("spgw_endpoint", "", "Endpoint to post SPGW-C JSON to - DEPRECATED") // TODO: remove me
	postDisable        = flag.Bool("post_disable", false, "Disable posting to connectivity service endpoints")
	postTimeout        = flag.Duration("post_timeout", time.Second*10, "Timeout duration when making post requests")
	aetherConfigAddr   = flag.String("aether_config_addr", "", "If specified, pull initial state from aether-config at this address")
	aetherConfigTarget = flag.String("aether_config_target", "connectivity-service-v4", "Target to use when pulling from aether-config")
	modelVersion       = flag.String("model_version", "v4", "Version of modeling to use") // DEPRECATED
	showModelList      = flag.Bool("show_models", false, "Show list of available modes")
	diagsPort          = flag.Uint("diags_port", 8080, "Port to use for Diagnostics API")
)

var log = logging.GetLogger("sdcore-adapter")

func serveMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(*metricAddr, nil); err != nil {
		log.Fatalf("failed to serve metrics: %v", err)
	}
}

// Synchronize and eat the error. This lets aether-config know we applied the
// configuration, but leaves us to retry applying it to the southbound device
// ourselves.
func synchronizerWrapper(s synchronizer.SynchronizerInterface) gnmi.ConfigCallback {
	return func(config ygot.ValidatedGoStruct, callbackType gnmi.ConfigCallbackType, path *pb.Path) error {
		err := s.Synchronize(config, callbackType, path)
		if err != nil {
			// Report the error, but do not send the error upstream.
			log.Warnf("Error during synchronize: %v", err)
		}
		return nil
	}
}

func main() {
	var sync synchronizer.SynchronizerInterface

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	// Initialize the synchronizer's service-specific code.
	if *modelVersion != "v4" {
		log.Panicf("modelVersion cannot be anything other than v4")
	}

	log.Infof("Initializing synchronizer")
	sync = synchronizerv4.NewSynchronizer(*outputFileName, !*postDisable, *postTimeout)

	// The synchronizer will convey its list of models.
	model := sync.GetModels()

	if *showModelList {
		fmt.Fprintf(os.Stdout, "Supported models:\n")
		for _, m := range model.SupportedModels() {
			fmt.Fprintf(os.Stdout, "  %s\n", m)
		}
		return
	}

	opts := credentials.ServerCredentials()
	g := grpc.NewServer(opts...)

	// outputFileName may have changed after processing arguments
	sync.SetOutputFileName(*outputFileName)
	sync.SetPostEnable(!*postDisable)
	sync.SetPostTimeout(*postTimeout)

	sync.Start()

	if (*configFile != "") && (*aetherConfigAddr != "") {
		log.Fatalf("use --configfile or --aetherConfigAddr, but not both")
	}

	var configData []byte

	// Optional: pull initial config from a local file
	if *configFile != "" {
		var err error
		configData, err = ioutil.ReadFile(*configFile)
		if err != nil {
			log.Fatalf("error in reading config file: %v", err)
		}
	}

	// Optional: pull initial config from onos-config
	if *aetherConfigAddr != "" {
		log.Infof("Fetching initial state from %s, target %s", *aetherConfigAddr, *aetherConfigTarget)
		// The migration library has the functions for fetching from onos-config
		srcVal, err := gnmiclient.GetPath(context.Background(), "", *aetherConfigTarget, *aetherConfigAddr)
		if err != nil {
			log.Fatalf("Error fetching initial data from onos-config: %s", err.Error())
			return
		}

		configData = srcVal.GetJsonVal()

		log.Infof("Fetched config: %s", string(configData))
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c)

	s, err := target.NewTarget(model, configData, synchronizerWrapper(sync))
	if err != nil {
		log.Fatalf("error in creating gnmi target: %v", err)
	}
	go func() {
		for {
			oscall := <-c
			if oscall.String() == "terminated" || oscall.String() == "interrupt" {
				log.Warnf("system call:%+v", oscall)
				s.Close()
				os.Exit(0)
			}
		}
	}()

	pb.RegisterGNMIServer(g, s)
	reflection.Register(g)

	log.Info("starting metric handler")
	go serveMetrics()

	log.Infof("starting out-of-band API on %d", *diagsPort)
	diagapi.StartDiagnosticAPI(s, *aetherConfigAddr, *aetherConfigTarget, *diagsPort)

	log.Infof("starting to listen on %s", *bindAddr)
	listen, err := net.Listen("tcp", *bindAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Info("starting to serve")
	if err := g.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
