// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

/*
 * Various certificate-related functions that configure the gNMI client used by migration.
 */

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
	caCert            = flag.String("ca_crt", "", "CA certificate file. Used to verify server TLS certificate.")
	clientCert        = flag.String("client_crt", "", "Client certificate file. Used for client certificate-based authentication.")
	clientKey         = flag.String("client_key", "", "Client private key file. Used for client certificate-based authentication.")
	tlsDisabled       = flag.Bool("tlsDisabled", false, "When set, caCert, clientCert & clientKey will be ignored")
	hostCheckDisabled = flag.Bool("hostCheckDisabled", false, "When set, host name in server cert will not be verified")
)

func readCerts(q client.Query) error {
	if *tlsDisabled {
		q.TLS = nil
		return nil
	}

	if *hostCheckDisabled {
		q.TLS.InsecureSkipVerify = true
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

		log.Infof("Successfully read and configured caCert %s", *caCert)

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

		log.Infof("Successfully read and configured clientCert %s and clientKey %s", *clientCert, *clientKey)

		q.TLS.Certificates = []tls.Certificate{certificate}
	}

	return nil
}
