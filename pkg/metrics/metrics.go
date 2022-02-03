// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"context"
	"fmt"
	"time"

	promApi "github.com/prometheus/client_golang/api"
	promApiV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promModel "github.com/prometheus/common/model"
)

// Fetcher is a wrapper around prometheus.
type Fetcher struct {
	Address string
	client  promApi.Client
	v1api   promApiV1.API
}

// NewFetcher creates a new Fetcher at the given prometheus address.
func NewFetcher(address string) (*Fetcher, error) {
	mf := &Fetcher{Address: address}
	return mf, mf.Connect()
}

// Connect to the Fetcher.
func (m *Fetcher) Connect() error {
	var err error

	m.client, err = promApi.NewClient(promApi.Config{
		Address: m.Address,
	})
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}

	m.v1api = promApiV1.NewAPI(m.client)

	return nil
}

// GetMetrics executes a query.
func (m *Fetcher) GetMetrics(query string) (promModel.Value, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := m.v1api.Query(ctx, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("error querying Prometheus: %v", err)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}

	// result is a Value, which is an interface to ValueType and a String() method
	// Can cast to:
	//    Matrix, Vector, *Scalar, *String

	return result, nil
}

// GetVector execute a query and return the result as a vector.
func (m *Fetcher) GetVector(query string) (promModel.Vector, error) {
	result, err := m.GetMetrics(query)
	if err != nil {
		return nil, err
	}

	v := result.(promModel.Vector)

	return v, nil
}

// GetSingleVector execute a query, assume the result is a vector of size one, and return it.
func (m *Fetcher) GetSingleVector(query string) (*float64, error) {
	v, err := m.GetVector(query)
	if err != nil {
		return nil, err
	}

	if len(v) != 1 {
		// TODO: no result; should this be an error
		return nil, nil
	}

	floatVal := float64(v[0].Value)

	return &floatVal, nil
}

// GetScalar execute a query and return the result as a scalar.
func (m *Fetcher) GetScalar(query string) (*float64, error) {
	result, err := m.GetMetrics(query)
	if err != nil {
		return nil, err
	}

	s := result.(*promModel.Scalar)
	if s == nil {
		return nil, nil
	}

	floatVal := float64(s.Value)

	return &floatVal, nil
}
