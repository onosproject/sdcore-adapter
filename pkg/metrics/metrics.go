// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package metrics

import (
	"context"
	"fmt"
	"time"

	promApi "github.com/prometheus/client_golang/api"
	promApiV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promModel "github.com/prometheus/common/model"
)

// MetricsFetcher is a wrapper around prometheus.

type MetricsFetcher struct {
	Address string
	client  promApi.Client
	v1api   promApiV1.API
}

// Create a new MetricsFetcher at the given prometheus address.
func NewMetricsFetcher(address string) (*MetricsFetcher, error) {
	mf := &MetricsFetcher{Address: address}
	return mf, mf.Connect()
}

// Connect to the MetricsFetcher.
func (m *MetricsFetcher) Connect() error {
	var err error

	m.client, err = promApi.NewClient(promApi.Config{
		Address: m.Address,
	})
	if err != nil {
		return fmt.Errorf("Error creating client: %v\n", err)
	}

	m.v1api = promApiV1.NewAPI(m.client)

	return nil
}

// Execute a query.
func (m *MetricsFetcher) GetMetrics(query string) (promModel.Value, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := m.v1api.Query(ctx, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("Error querying Prometheus: %v\n", err)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}

	// result is a Value, which is an interface to ValueType and a String() method
	// Can cast to:
	//    Matrix, Vector, *Scalar, *String

	return result, nil
}

// Execute a query and return the result as a vector.
func (m *MetricsFetcher) GetVector(query string) (promModel.Vector, error) {
	result, err := m.GetMetrics(query)
	if err != nil {
		return nil, err
	}

	v := result.(promModel.Vector)

	return v, nil
}

// Execute a query, assume the result is a vector of size one, and return it.
func (m *MetricsFetcher) GetSingleVector(query string) (*float64, error) {
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

// Execute a query and return the result as a scalar.
func (m *MetricsFetcher) GetScalar(query string) (*float64, error) {
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
