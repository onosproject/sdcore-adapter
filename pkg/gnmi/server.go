// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package gnmi

import (
	"encoding/json"
	"github.com/eapache/channels"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
)

// NewServer creates an instance of Server with given json config.
func NewServer(model *Model, callback ConfigCallback) (*Server, error) {
	s := &Server{
		model:    model,
		config:   ConfigForest{},
		callback: callback,
	}

	s.subscribed = make(map[string][]*streamClient)

	/* Create a RingChannel that can hold 100 items, and will discard
	 * the oldest if it becomes full.
	 *
	 * TODO: This needs further testing and design work. It's unclear
	 *    whether we have any subscribers now, and if so then simply
	 *    deleting the channel may suffice. If we do have (or expect to
	 *    have) subscribers then we may need a more robust solution.
	 *    The RingChannel will prevent blocking, but may lead to loss of
	 *    notifications.
	 */

	s.ConfigUpdate = channels.NewRingChannel(100)

	return s, nil
}

// Close - called on shutdown - shutdown gracefully
func (s *Server) Close() {
	log.Info("Shutting down gNMI server")
	for sub, streamClientList := range s.subscribed {
		log.Infof("Closing subscriptions to %s. %d", sub, len(streamClientList))
		for _, sc := range streamClientList {
			close(sc.UpdateChan)
		}
	}

	if s.ConfigUpdate != nil {
		log.Info("Closing Ring Buffer Channel")
		s.ConfigUpdate.Close()
	}
}

// ExecuteCallbacks executes the callbacks for the synchronizer
func (s *Server) ExecuteCallbacks(reason ConfigCallbackType, target string, path *pb.Path) error {
	if s.callback != nil {
		if err := s.callback(s.config, reason, target, path); err != nil {
			return err
		}
	}
	return nil
}

// GetJSON returns the JSON value of the config tree
func (s *Server) GetJSON(target string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	jsonTree, err := ygot.ConstructIETFJSON(s.config[target], &ygot.RFC7951JSONConfig{})
	if err != nil {
		return []byte{}, err
	}
	data, err := json.MarshalIndent(jsonTree, "", "  ")
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

// PutJSON sets the config tree from a json value
func (s *Server) PutJSON(target string, b []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rootStruct, err := s.model.NewConfigStruct(b)
	if err != nil {
		return err
	}
	s.config[target] = rootStruct
	return nil
}
