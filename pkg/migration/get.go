// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

/*
 * Library functions to facilitate gNMI get operations.
 */

package migration

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/openconfig/gnmi/client"
	gclient "github.com/openconfig/gnmi/client/gnmi"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
)

// MockGetFunction can be used to mock a gNMI Get Operation for unit testing.
type MockGetFunction func(*gpb.GetRequest) (*gpb.GetResponse, error)

// MockGet can be set to a mocking function to mock ExecuteGet.
var MockGet MockGetFunction

// ExecuteGet executes gNMI Get request against the server named by addr.
func ExecuteGet(ctx context.Context, r *gpb.GetRequest, addr string) (*gpb.GetResponse, error) {
	// for ease of unit testing
	if MockGet != nil {
		return MockGet(r)
	}

	q := client.Query{TLS: &tls.Config{}}

	err := readCerts(q)
	if err != nil {
		return nil, err
	}

	ctx = getAuthContext(ctx)

	c, err := gclient.New(ctx, client.Destination{
		Addrs:       []string{addr},
		Timeout:     q.Timeout,
		Credentials: q.Credentials,
		TLS:         q.TLS,
	})

	if err != nil {
		return nil, fmt.Errorf("could not create a gNMI client: %v", err)
	}

	response, err := c.(*gclient.Client).Get(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("target returned RPC error for Get(%q): %v", r.String(), err)
	}

	return response, nil
}

// GetPath executes a gNMI Get Operation using the named path and target.
func GetPath(ctx context.Context, path string, target string, addr string) (*gpb.TypedValue, error) {
	req := &gpb.GetRequest{
		Path:     []*gpb.Path{StringToPath(path, target)},
		Encoding: gpb.Encoding_JSON,
	}

	resp, err := ExecuteGet(ctx, req, addr)
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
