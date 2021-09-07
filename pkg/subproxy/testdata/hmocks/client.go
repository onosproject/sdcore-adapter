// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package hmocks

import "net/http"

type HttpMockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *HttpMockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

var (
	GetDoFunc func(req *http.Request) (*http.Response, error)
)
