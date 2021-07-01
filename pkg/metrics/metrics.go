// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package metrics

import (
	"context"
	"flag"
	"fmt"
	"time"

	promApi "github.com/prometheus/client_golang/api"
	promApiV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promModel "github.com/prometheus/common/model"
)

var (
	metricAddr = flag.String("metric_address", "http://aether-roc-umbrella-prometheus-server:80/", "Prometheus metric endpoint bind to retrieve metrics from")
)

type UEMetrics struct {
	Active   int32
	Inactive int32
	Idle     int32
}

func GetMetrics(query string) (promModel.Value, error) {
	client, err := promApi.NewClient(promApi.Config{
		Address: *metricAddr,
	})
	if err != nil {
		return nil, fmt.Errorf("Error creating client: %v\n", err)
	}

	v1api := promApiV1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := v1api.Query(ctx, query, time.Now())
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

func GetVector(query string) (promModel.Vector, error) {
	result, err := GetMetrics(query)
	if err != nil {
		return nil, err
	}

	v := result.(promModel.Vector)

	return v, nil
}

func GetSingleVector(query string) (*float64, error) {
	v, err := GetVector(query)
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

func GetScalar(query string) (*float64, error) {
	result, err := GetMetrics(query)
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

func GetSliceUEMetrics(sliceName string) (*UEMetrics, error) {
	query := fmt.Sprintf("sum by (state) (smf_pdu_session_profile{slice=\"%s\"})", sliceName)
	result, err := GetMetrics(query)
	if err != nil {
		return nil, err
	}

	v := result.(promModel.Vector)

	if len(v) == 0 {
		return nil, nil
	}

	m := UEMetrics{}

	for _, sample := range v {
		state, okay := sample.Metric["state"]
		if !okay {
			continue
		}
		switch state {
		case "active":
			m.Active += int32(sample.Value)
		case "inactive":
			m.Inactive += int32(sample.Value)
		case "idle":
			m.Idle += int32(sample.Value)
		}
	}

	return &m, nil
}
