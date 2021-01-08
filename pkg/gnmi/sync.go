// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package gnmi

// doDelete deletes the path from the json tree if the path exists. If success,
// it calls the callback function to apply the change to the device hardware.
func (s *Server) Synchronize() {
	if s.synchronizer == nil {
		return
	}
	err := s.synchronizer.Synchronize(s.config)
	if err != nil {
		// Report the error, but do not send the error upstream.
		log.Warnf("Error during synchronize: %v", err)
	}
}
