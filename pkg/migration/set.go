// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package migration

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/openconfig/gnmi/client"
	gclient "github.com/openconfig/gnmi/client/gnmi"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
)

func ExecuteSet(r *gpb.SetRequest, addr string, ctx context.Context) (*gpb.SetResponse, error) {
	q := client.Query{TLS: &tls.Config{}}

	err := readCerts(q)
	if err != nil {
		return nil, err
	}

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

func Update(prefix *gpb.Path, target string, addr string, updates []*gpb.Update, ctx context.Context) error {
	req := &gpb.SetRequest{
		Prefix: prefix,
		Update: updates,
	}

	resp, err := ExecuteSet(req, addr, ctx)
	if err != nil {
		return err
	}

	_ = resp

	return nil
}

func Delete(prefix *gpb.Path, target string, addr string, deletes []*gpb.Path, ctx context.Context) error {
	req := &gpb.SetRequest{
		Prefix: prefix,
		Delete: deletes,
	}

	resp, err := ExecuteSet(req, addr, ctx)
	if err != nil {
		return err
	}

	_ = resp

	return nil
}
