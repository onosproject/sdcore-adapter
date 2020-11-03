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
	"strings"
)

func ExecuteGet(r *gpb.GetRequest, addr string, ctx context.Context) (*gpb.GetResponse, error) {
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

	response, err := c.(*gclient.Client).Get(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("target returned RPC error for Get(%q): %v", r.String(), err)
	}

	return response, nil
}

func StringToPath(s string, target string) *gpb.Path {
	elems := []*gpb.PathElem{}

	for _, part := range strings.Split(s, "/") {
		if len(part) > 0 {
			elem := &gpb.PathElem{Name: part}
			elems = append(elems, elem)
		}
	}

	return &gpb.Path{
		Target: target,
		Elem:   elems,
	}
}

func GetPath(path string, target string, addr string, ctx context.Context) (*gpb.TypedValue, error) {
	req := &gpb.GetRequest{
		Path:     []*gpb.Path{StringToPath(path, target)},
		Encoding: gpb.Encoding_JSON,
	}

	resp, err := ExecuteGet(req, addr, ctx)
	if err != nil {
		return nil, err
	}

	log.Infof("GET numNot=%d", len(resp.Notification))
	log.Infof("GET numUpdate=%d", len(resp.Notification[0].Update))

	//log.Infof("GET %v", resp.Notification[0].Update[0].Val.GetJsonVal())

	//json := resp.Notification[0].Update[0].Val.GetJsonVal()

	return resp.Notification[0].Update[0].Val, nil
}
