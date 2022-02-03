// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package gnmiclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/grpc/retry"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/openconfig/gnmi/client"
	gclient "github.com/openconfig/gnmi/client/gnmi"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"os"
	"strings"
	"time"
)

var log = logging.GetLogger("gnmiclient")

const (
	authorization = "Authorization"
	secretname    = "subscriber-proxy-keycloak-secret"
)

// GnmiInterface - abstract definition of the Gnmi interface
//go:generate mockgen -destination=../test/mocks/mock_gnmi.go -package=mocks github.com/onosproject/sdcore-adapter/pkg/gnmiclient GnmiInterface
type GnmiInterface interface {
	// GetPath - gNMI Get
	GetPath(ctx context.Context, path string, target string, addr string) (*gpb.TypedValue, error)
	// Update for gNMI Set
	Update(ctx context.Context, prefix *gpb.Path, target string, addr string, updates []*gpb.Update) error
	// Delete for gNMI Set
	Delete(ctx context.Context, prefix *gpb.Path, target string, addr string, deletes []*gpb.Path) error
	// CloseClient - close the gNMI connection
	CloseClient()
	// Address Get the address
	Address() string
}

// Gnmi - concrete implementation
type Gnmi struct {
	address string
	Client  client.Impl
}

// NewGnmi - create one gNMI client and keep it open
func NewGnmi(addr string, timeout time.Duration) (GnmiInterface, error) {
	gnmi := new(Gnmi)
	gnmi.address = addr
	var err error

	q := client.Query{TLS: &tls.Config{}, Timeout: timeout}

	if err = readCerts(q); err != nil {
		return nil, err
	}

	ctx := getAuthContext(context.Background())

	gnmi.Client, err = gclient.New(ctx, client.Destination{
		Addrs:       []string{addr},
		Timeout:     q.Timeout,
		Credentials: q.Credentials,
		TLS:         q.TLS,
	})
	if err != nil {
		return nil, errors.NewInternal("could not create a gNMI client: %v", err)
	}

	return gnmi, nil
}

//NewGnmiWithInterceptor - create one gNMI client and keep it open with retry mechanism
func NewGnmiWithInterceptor(addr string, timeout time.Duration) (GnmiInterface, string, error) {
	gnmi := new(Gnmi)
	gnmi.address = addr
	var err error

	q := client.Query{TLS: &tls.Config{}, Timeout: timeout}

	if err = readCerts(q); err != nil {
		return nil, "", err
	}

	ctx := context.Background()
	token := ""
	openIDIssuer := os.Getenv("OIDC_SERVER_URL")
	if len(strings.TrimSpace(openIDIssuer)) > 0 {
		token, err = GetAccessToken(openIDIssuer, secretname)
		if err != nil {
			return nil, token, err
		}
		token = "Bearer " + token
		ctx = metadata.AppendToOutgoingContext(ctx, authorization, token)
		log.Debug("[INFO] Added Bearer Token to context ")
	} else {
		ctx = getAuthContext(ctx)
	}

	opts := []grpc.DialOption{}

	opts = append(opts, grpc.WithUnaryInterceptor(retry.RetryingUnaryClientInterceptor(retry.WithInterval(100*time.Millisecond))))

	d := client.Destination{
		Addrs:       []string{addr},
		Timeout:     q.Timeout,
		Credentials: q.Credentials,
		TLS:         q.TLS,
	}

	switch d.TLS {
	case nil:
		opts = append(opts, grpc.WithInsecure())
	default:
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(d.TLS)))
	}

	gCtx, cancel := context.WithTimeout(ctx, q.Timeout)
	defer cancel()

	addr = ""
	if len(d.Addrs) != 0 {
		addr = d.Addrs[0]
	}
	conn, err := grpc.DialContext(gCtx, addr, opts...)
	if err != nil {
		return nil, token, errors.NewInternal(fmt.Sprintf("Dialer(%s, %v): %v", addr, d.Timeout, err))
	}

	gnmi.Client, err = gclient.NewFromConn(ctx, conn, d)

	if err != nil {
		return nil, token, errors.NewInternal(fmt.Sprintf("could not create a gNMI client: %v", err))
	}
	log.Info("[INFO] gnmi client connected !!! ")

	return gnmi, token, nil
}

// CloseClient - close the gNMI Client when finished
func (g *Gnmi) CloseClient() {
	if err := g.Client.Close(); err != nil {
		log.Errorf("Unable to close gnmi client %s", err)
	}
}

// Address - get the server address
func (g *Gnmi) Address() string {
	return g.address
}

// GetPath - gNMI Get
func (g *Gnmi) GetPath(ctx context.Context, path string, target string, addr string) (*gpb.TypedValue, error) {
	req := &gpb.GetRequest{
		Path:     []*gpb.Path{StringToPath(path, target)},
		Encoding: gpb.Encoding_JSON,
	}

	resp, err := executeGet(ctx, req, addr)
	if err != nil {
		return nil, err
	}

	if len(resp.Notification) == 0 {
		return nil, nil
	}

	if len(resp.Notification[0].Update) == 0 {
		return nil, nil
	}

	return resp.Notification[0].Update[0].Val, nil
}

// Update for gNMI Set
func (g *Gnmi) Update(ctx context.Context, prefix *gpb.Path, target string, addr string, updates []*gpb.Update) error {
	req := &gpb.SetRequest{
		Prefix: prefix,
		Update: updates,
	}

	_, err := executeSet(ctx, req, addr)
	return err
}

// Delete for gNMI Set
func (g *Gnmi) Delete(ctx context.Context, prefix *gpb.Path, target string, addr string, deletes []*gpb.Path) error {
	req := &gpb.SetRequest{
		Prefix: prefix,
		Delete: deletes,
	}

	_, err := executeSet(ctx, req, addr)
	return err
}
