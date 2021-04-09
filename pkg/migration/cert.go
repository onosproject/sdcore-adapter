// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

/*
 * Various certificate-related functions that configure the gNMI client used by migration.
 */

package migration

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/openconfig/gnmi/client"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
)

var (
	// Certificate files.
	caCert            = flag.String("ca_crt", "", "CA certificate file. Used to verify server TLS certificate.")
	clientCert        = flag.String("client_crt", "", "Client certificate file. Used for client certificate-based authentication.")
	clientKey         = flag.String("client_key", "", "Client private key file. Used for client certificate-based authentication.")
	tlsDisabled       = flag.Bool("tlsDisabled", false, "When set, caCert, clientCert & clientKey will be ignored")
	hostCheckDisabled = flag.Bool("hostCheckDisabled", false, "When set, host name in server cert will not be verified")
	authHeader        = flag.String("ah", "", "Authorization token to use when contacting aether-config")
)

// Best practice: Keys used in context.WithValue should use unique types to
// prevent collision.
type contextKey int

const (
	authContextKey contextKey = iota
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

		log.Debugf("Successfully read and configured caCert %s", *caCert)

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

		log.Debugf("Successfully read and configured clientCert %s and clientKey %s", *clientCert, *clientKey)

		q.TLS.Certificates = []tls.Certificate{certificate}
	}

	return nil
}

// Add an authorization header to a context
func WithAuthorization(ctx context.Context, auth string) context.Context {
	return context.WithValue(ctx, authContextKey, auth)
}

// Propagate through the authorization header into HTTP metadata to make an HTTP request,
// either from the context, or from the default command-line options.
func getAuthContext(ctx context.Context) context.Context {
	reqAuthHeader := OverrideFromContext(ctx, authContextKey, *authHeader).(string)
	if reqAuthHeader != "" {
		md := make(metadata.MD)
		md.Set("authorization", reqAuthHeader)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	return ctx
}
