// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

/*
 * Library functions to facilitate gNMI set operations.
 */

package gnmiclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/openconfig/gnmi/client"
	gclient "github.com/openconfig/gnmi/client/gnmi"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"time"
)

// MockSetFunction is a function that can be used for mocking executeSet in unit tests.
// Deprecated. Use MockGnmiInterface instead
type MockSetFunction func(*gpb.SetRequest) (*gpb.SetResponse, error)

// MockSet is a variable that will enable mocking of executeSet.
// Deprecated. Use MockGnmiInterface instead
var MockSet MockSetFunction

// executeSet executes a gNMI set request
func executeSet(ctx context.Context, r *gpb.SetRequest, addr string) (*gpb.SetResponse, error) {
	// for ease of unit testing
	if MockSet != nil {
		return MockSet(r)
	}

	q := client.Query{TLS: &tls.Config{}, Timeout: 5 * time.Second}

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

	response, err := c.(*gclient.Client).Set(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("target returned RPC error for Set(%q): %v", r.String(), err)
	}

	return response, nil
}

// Update performs a gNMI Update Set operation
// Deprecated. Use GnmiInterface instead
func Update(ctx context.Context, prefix *gpb.Path, target string, addr string, updates []*gpb.Update) error {
	req := &gpb.SetRequest{
		Prefix: prefix,
		Update: updates,
	}

	_, err := executeSet(ctx, req, addr)
	return err
}

// Delete performs a gNMI Delete set operation
// Deprecated. Use GnmiInterface instead
func Delete(ctx context.Context, prefix *gpb.Path, target string, addr string, deletes []*gpb.Path) error {
	req := &gpb.SetRequest{
		Prefix: prefix,
		Delete: deletes,
	}

	_, err := executeSet(ctx, req, addr)
	return err
}
