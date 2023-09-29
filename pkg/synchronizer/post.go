// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// RESTPusher implements a pusher that pushes to a REST API endpoint.

package synchronizer

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

// PushError is an error class that is returned for failed POSTs and DELETEs. It
// makes it easier to detect a nonfatal error, such as a 404.
type PushError struct {
	Endpoint   string
	StatusCode int
	Status     string
	Operation  string
}

func (e *PushError) Error() string {
	return fmt.Sprintf("Push Error op=%s endpoint=%s code=%d status=%s", e.Operation, e.Endpoint, e.StatusCode, e.Status)
}

// RESTPusher implements a pusher that pushes to a rest endpoint.
type RESTPusher struct {
}

// PushUpdate pushes an update to the REST endpoint.
func (p *RESTPusher) PushUpdate(endpoint string, data []byte) error {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	log.Infof("Push Update endpoint=%s data=%s", endpoint, string(data))


	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(data))
	if err != nil {
		log.Infof("NewRequest failed and returned status %s", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Infof("do failed and returned status %s", err)
		return err
	}

	defer resp.Body.Close()

	log.Infof("Put returned status %s", resp.Status)

	if (resp.StatusCode < 200) || (resp.StatusCode >= 300) {
		return &PushError{Operation: "POST", Endpoint: endpoint, StatusCode: resp.StatusCode, Status: resp.Status}
	}

	return nil
}

// PushDelete pushes a delete to the REST endpoint
func (p *RESTPusher) PushDelete(endpoint string) error {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	log.Infof("Push Delete endpoint=%s", endpoint)

	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	log.Infof("Delete returned status %s", resp.Status)

	if (resp.StatusCode < 200) || (resp.StatusCode >= 300) {
		return &PushError{Operation: "DELETE", Endpoint: endpoint, StatusCode: resp.StatusCode, Status: resp.Status}
	}

	return nil
}
