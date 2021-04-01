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

type SecuritySettings struct {
	caCert            string
	clientCert        string
	clientKey         string
	tlsDisabled       bool
	hostCheckDisabled bool
}

func GetDefaultSecuritySettings() *SecuritySettings {
	s := SecuritySettings{}

	s.caCert = *caCert
	s.clientCert = *clientCert
	s.clientKey = *clientKey
	s.hostCheckDisabled = *hostCheckDisabled
	s.tlsDisabled = *tlsDisabled

	return &s
}

func readCerts(sec *SecuritySettings, q client.Query) error {
	if sec.tlsDisabled {
		q.TLS = nil
		return nil
	}

	if sec.hostCheckDisabled {
		q.TLS.InsecureSkipVerify = true
	}

	if sec.caCert != "" {
		certPool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(sec.caCert)
		if err != nil {
			return fmt.Errorf("could not read %q: %s", sec.caCert, err)
		}
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			return fmt.Errorf("failed to append CA certificates")
		}

		log.Debugf("Successfully read and configured caCert %s", sec.caCert)

		q.TLS.RootCAs = certPool
	}

	if sec.clientCert != "" || sec.clientKey != "" {
		if sec.clientCert == "" || sec.clientKey == "" {
			return fmt.Errorf("--client_crt and --client_key must be set with file locations")
		}
		certificate, err := tls.LoadX509KeyPair(sec.clientCert, sec.clientKey)
		if err != nil {
			return fmt.Errorf("could not load client key pair: %s", err)
		}

		log.Debugf("Successfully read and configured clientCert %s and clientKey %s", sec.clientCert, sec.clientKey)

		q.TLS.Certificates = []tls.Certificate{certificate}
	}

	return nil
}
