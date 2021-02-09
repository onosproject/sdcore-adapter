// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/google/gnxi/utils/credentials"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"github.com/onosproject/sdcore-adapter/pkg/synchronizer"
	"github.com/onosproject/sdcore-adapter/pkg/target"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	bindAddr       = flag.String("bind_address", ":10161", "Bind to address:port or just :port")
	metricAddr     = flag.String("metric_address", ":9851", "Prometheus metric endpoint bind to address:port or just :port")
	configFile     = flag.String("config", "", "IETF JSON file for target startup config")
	outputFileName = flag.String("output", "", "JSON file to save output to")
	_              = flag.String("spgw_endpoint", "", "Endpoint to post SPGW-C JSON to - DEPRECATED") // TODO: remove me
	postDisable    = flag.Bool("post_disable", false, "Disable posting to connectivity service endpoints")
	postTimeout    = flag.Duration("post_timeout", time.Second*10, "Timeout duration when making post requests")
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
func synchronizerWrapper(s *synchronizer.Synchronizer) gnmi.ConfigCallback {
	return func(config ygot.ValidatedGoStruct, callbackType gnmi.ConfigCallbackType) error {
		err := s.Synchronize(config, callbackType)
		if err != nil {
			// Report the error, but do not send the error upstream.
			log.Warnf("Error during synchronize: %v", err)
		}
		return nil
	}
}

func main() {
	// Initialize the synchronizer's service-specific code.
	sync := synchronizer.NewSynchronizer(*outputFileName, !*postDisable, *postTimeout)

	// The synchronizer will convey its list of models.
	model := sync.GetModels()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Supported models:\n")
		for _, m := range model.SupportedModels() {
			fmt.Fprintf(os.Stderr, "  %s\n", m)
		}
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	opts := credentials.ServerCredentials()
	g := grpc.NewServer(opts...)

	// outputFileName may have changed after processing arguments
	sync.SetOutputFileName(*outputFileName)
	sync.SetPostEnable(!*postDisable)
	sync.SetPostTimeout(*postTimeout)

	sync.Start()

	var configData []byte
	if *configFile != "" {
		var err error
		configData, err = ioutil.ReadFile(*configFile)
		if err != nil {
			log.Fatalf("error in reading config file: %v", err)
		}
	}

	s, err := target.NewTarget(model, configData, synchronizerWrapper(sync))

	if err != nil {
		log.Fatalf("error in creating gnmi target: %v", err)
	}
	pb.RegisterGNMIServer(g, s)
	reflection.Register(g)

	log.Info("starting metric handler")
	go serveMetrics()

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
