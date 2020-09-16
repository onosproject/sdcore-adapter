// Copyright 2019-present Open Networking Foundation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Binary sdcore_adapter implements a target that communicates with
// aether-config via gNMI, and pushes configuration to SD-CORE via
// json.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	log "github.com/golang/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/onosproject/sdcore-adapter/pkg/gnmi_target"

	"github.com/google/gnxi/utils/credentials"

	pb "github.com/openconfig/gnmi/proto/gnmi"

	"github.com/onosproject/sdcore-adapter/pkg/synchronizer"
)

var (
	bindAddr       = flag.String("bind_address", ":10161", "Bind to address:port or just :port")
	configFile     = flag.String("config", "", "IETF JSON file for target startup config")
	outputFileName = flag.String("output", "", "JSON file to save output to")
)

func main() {
	// Initialize the synchronizer's service-specific code.
	sync := synchronizer.NewSynchronizer(*outputFileName)

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

	var configData []byte
	if *configFile != "" {
		var err error
		configData, err = ioutil.ReadFile(*configFile)
		if err != nil {
			log.Exitf("error in reading config file: %v", err)
		}
	}

	s, err := gnmi_target.NewServer(model, configData, sync)

	if err != nil {
		log.Exitf("error in creating gnmi target: %v", err)
	}
	pb.RegisterGNMIServer(g, s)
	reflection.Register(g)

	// Perform initial synchronization, in particular if there is any data
	// that was supplied as part of the --config flag.
	s.Synchronize()

	log.Infof("starting to listen on %s", *bindAddr)
	listen, err := net.Listen("tcp", *bindAddr)
	if err != nil {
		log.Exitf("failed to listen: %v", err)
	}

	log.Info("starting to serve")
	if err := g.Serve(listen); err != nil {
		log.Exitf("failed to serve: %v", err)
	}

}
