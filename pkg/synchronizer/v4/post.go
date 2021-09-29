// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// RESTPusher implements a pusher that pushes to a REST API endpoint.

package synchronizerv4

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

// RESTPusher implements a pusher that pushes to a rest endpoint.
type RESTPusher struct {
}

// PushUpdate pushes an update to the REST endpoint.
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
