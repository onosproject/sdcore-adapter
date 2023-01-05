// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package promkafka implements a prometheus-to-kafka gateway
package promkafka

import (
	"github.com/onosproject/analytics/pkg/messages"
	"time"
)

// IPAddressEvent is an event that assigns an IP Address to a UE
type IPAddressEvent struct {
	Type      string    `json:"type"`
	Imsi      string    `json:"imsi"`
	Connected bool      `json:"connected"`
	IPAddress string    `json:"ipaddress"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

/*
MessageType is required to satisfy the Message interface definition
*/
func (event IPAddressEvent) MessageType() messages.MessageType {
	return messages.EVENT
}
