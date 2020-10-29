// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package migration

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/openconfig/gnmi/client"
	"io/ioutil"
)

var (
	// Certificate files.
	caCert      = flag.String("ca_crt", "", "CA certificate file. Used to verify server TLS certificate.")
	clientCert  = flag.String("client_crt", "", "Client certificate file. Used for client certificate-based authentication.")
	clientKey   = flag.String("client_key", "", "Client private key file. Used for client certificate-based authentication.")
	tlsDisabled = flag.Bool("tlsDisabled", false, "When set, caCert, clientCert & clientKey will be ignored")
)

func readCerts(q client.Query) error {
	if *tlsDisabled {
		q.TLS = nil
		return nil
	}

	if *caCert != "" {
		certPool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(*caCert)
		if err != nil {
			return fmt.Errorf("could not read %q: %s", *caCert, err)
		}
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			return fmt.Errorf("failed to append CA certificates")
		}

		q.TLS.RootCAs = certPool
	}

	if *clientCert != "" || *clientKey != "" {
		if *clientCert == "" || *clientKey == "" {
			return fmt.Errorf("--client_crt and --client_key must be set with file locations")
		}
		certificate, err := tls.LoadX509KeyPair(*clientCert, *clientKey)
		if err != nil {
			return fmt.Errorf("could not load client key pair: %s", err)
		}

		q.TLS.Certificates = []tls.Certificate{certificate}
	}

	return nil
}
