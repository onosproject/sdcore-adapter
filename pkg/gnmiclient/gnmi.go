// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package gnmiclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/openconfig/gnmi/client"
	gclient "github.com/openconfig/gnmi/client/gnmi"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"time"
)

var log = logging.GetLogger("gnmiclient")

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
		return nil, fmt.Errorf("could not create a gNMI client: %v", err)
	}

	return gnmi, nil
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
