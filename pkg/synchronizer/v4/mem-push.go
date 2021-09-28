// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// MemPusher implements a pusher that pushes to an in-memory map, which can easily be
// retrieved by a unit test.

package synchronizerv4

// MemPusher pushes sychronizations to an in-memory map, for ease of unit testing.
type MemPusher struct {
	Pushes map[string]string
}

// NewMemPusher creates a new MemPusher
func NewMemPusher() *MemPusher {
	return &MemPusher{
		Pushes: map[string]string{},
	}
}

// PushUpdate pushes an update, storing it in memory
func (p *MemPusher) PushUpdate(endpoint string, data []byte) error {
	p.Pushes[endpoint] = string(data)
	return nil
}
