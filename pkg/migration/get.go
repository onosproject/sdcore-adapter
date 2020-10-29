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

func ExecuteGet(r *gpb.GetRequest, ctx context.Context) error {
	q := client.Query{TLS: &tls.Config{}}

	err := readCerts(q)
	if err != nil {
		return err
	}

	// TODO: should this have been set in caller?
	r.Encoding = gpb.Encoding_PROTO // altnerate - Encoding_JSON

	c, err := gclient.New(ctx, client.Destination{
		Addrs:       q.Addrs,
		Target:      q.Target,
		Timeout:     q.Timeout,
		Credentials: q.Credentials,
		TLS:         q.TLS,
	})

	if err != nil {
		return fmt.Errorf("could not create a gNMI client: %v", err)
	}
	response, err := c.(*gclient.Client).Get(ctx, r)
	if err != nil {
		return fmt.Errorf("target returned RPC error for Get(%q): %v", r.String(), err)
	}
	_ = response
	//cfg.Display([]byte(proto.MarshalTextString(response)))
	return nil
}
