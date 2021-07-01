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

type UEMetrics struct {
	Active   int32
	Inactive int32
	Idle     int32
}

type MetricsFetcher struct {
	Address string
	client  promApi.Client
	v1api   promApiV1.API
}

func NewMetricsFetcher(address string) (*MetricsFetcher, error) {
	mf := &MetricsFetcher{Address: address}
	return mf, mf.Connect()
}

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

func (m *MetricsFetcher) GetVector(query string) (promModel.Vector, error) {
	result, err := m.GetMetrics(query)
	if err != nil {
		return nil, err
	}

	v := result.(promModel.Vector)

	return v, nil
}

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

func (m *MetricsFetcher) GetSliceUEMetrics(sliceName string) (*UEMetrics, error) {
	query := fmt.Sprintf("sum by (state) (smf_pdu_session_profile{slice=\"%s\"})", sliceName)
	result, err := m.GetMetrics(query)
	if err != nil {
		return nil, err
	}

	v := result.(promModel.Vector)

	if len(v) == 0 {
		return nil, nil
	}

	uem := UEMetrics{}

	for _, sample := range v {
		state, okay := sample.Metric["state"]
		if !okay {
			continue
		}
		switch state {
		case "active":
			uem.Active += int32(sample.Value)
		case "inactive":
			uem.Inactive += int32(sample.Value)
		case "idle":
			uem.Idle += int32(sample.Value)
		}
	}

	return &uem, nil
}
