// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// RESTPusher implements a pusher that pushes to a REST API endpoint.

package synchronizerv3

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

type RESTPusher struct {
}

// http.client lacks a "put" operation
// NOTE: This will be needed when sdcore implements put operation.
/*
func httpPut(client *http.Client, endpoint string, mimetype string, data []byte) (*http.Response, error) {
	// set the HTTP method, url, and request body
	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", mimetype)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
*/

func (p *RESTPusher) PushUpdate(endpoint string, data []byte) error {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	log.Infof("Push Update endpoint=%s data=%s", endpoint, string(data))

	resp, err := client.Post(
		endpoint,
		"application/json",
		bytes.NewBuffer(data))

	/* In the future, PUT will be the correct operation
	resp, err := httpPut(client, endpoint, "application/json", data)
	*/

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	log.Infof("Put returned status %s", resp.Status)

	if resp.StatusCode != 200 {
		return fmt.Errorf("Put returned error %s", resp.Status)
	}

	return nil
}
