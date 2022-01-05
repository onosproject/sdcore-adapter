// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizer

/*
 * Deletes are always handled synchronously. The synchronizer stops and wait for a delete to be
 * completed before it continues. This is to handle the case where we fail while performing the
 * delete. It'll get marked as a FAILED transaction in onos-config.
 *
 * This is in contrasts to configuration updates, which are generally handled asynchronously.
 */

import (
	"fmt"
	models "github.com/onosproject/config-models/modelplugin/aether-2.0.0/aether_2_0_0"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
)

// GetObjectID return the object id for a path.
// If the path is a top-level model, then return the id of the object. If
// the object is not a top-level model then return nil.
func (s *Synchronizer) GetObjectID(modelName string, path *pb.Path) (*string, error) {
	// Example path: /vcs/vcs[id=something]
	if len(path.Elem) > 2 {
		// It's for some portion of the model, not the root of the model.
		// We don't care.
		return nil, nil
	}
	if path.Elem[1].Name != modelName {
		// It's hard to imagine what else this could be. Future-proof by ignoring it.
		return nil, nil
	}
	id, okay := path.Elem[1].Key["id"]
	if !okay {
		return nil, fmt.Errorf("Delete of %s does not have an id key", modelName)
	}

	return &id, nil
}

// DeleteVcs deletes a Vcs from the core
func (s *Synchronizer) DeleteVcs(device *models.Device, path *pb.Path) error {
	id, err := s.GetObjectID("vcs", path)
	if err != nil {
		return err
	}
	if id == nil {
		return nil
	}
	log.Infof("Delete vcs %s", *id)

	vcs, err := s.GetVcs(device, id)
	if err != nil {
		return err
	}

	// For each connectivity service, delete the VCS from the connectivity
	// service.
	csList, err := s.GetConnectivityServiceForSite(device, vcs.Site)
	if err != nil {
		return err
	}
csLoop:
	for _, cs := range csList {
		url := fmt.Sprintf("%s/v1/network-slice/%s", *cs.Core_5GEndpoint, *id)
		err = s.pusher.PushDelete(url)
		if err != nil {
			pushError, ok := err.(*PushError)
			if ok && pushError.StatusCode == 404 {
				// This may mean we already deleted it.
				log.Infof("Tried to delete vcs %s but it does not exist", *id)
				continue csLoop
			}
			return fmt.Errorf("Vcs %s failed to push delete: %s", *id, err)
		}
	}

	// Remove slice from the cache
	s.CacheDelete(CacheModelSlice, *id)
	s.CacheDelete(CacheModelSliceUpf, *id)

	return nil
}

// DeleteDeviceGroup deletes a devicegroup from the core
func (s *Synchronizer) DeleteDeviceGroup(device *models.Device, path *pb.Path) error {
	id, err := s.GetObjectID("device-group", path)
	if err != nil {
		return err
	}
	if id == nil {
		return nil
	}
	log.Infof("Delete device-group %s", *id)

	dg, err := s.GetDeviceGroup(device, id)
	if err != nil {
		return err
	}

	csList, err := s.GetConnectivityServiceForSite(device, dg.Site)
	if err != nil {
		return err
	}
csLoop:
	for _, cs := range csList {
		url := fmt.Sprintf("%s/v1/device-group/%s", *cs.Core_5GEndpoint, *id)
		err = s.pusher.PushDelete(url)
		if err != nil {
			pushError, ok := err.(*PushError)
			if ok && pushError.StatusCode == 404 {
				// This may mean we already deleted it.
				log.Infof("Tried to delete vcs %s but it does not exist", *id)
				continue csLoop
			}
			return fmt.Errorf("Device-Group %s failed to push delete: %s", *id, err)
		}
	}

	// Remove device-group from the cache
	s.CacheDelete(CacheModelDeviceGroup, *id)

	return nil
}

// HandleDelete synchronously performs a delete
func (s *Synchronizer) HandleDelete(config ygot.ValidatedGoStruct, path *pb.Path) error {
	device := config.(*models.Device)

	if path == nil || len(path.Elem) == 0 {
		return nil
	}
	if len(path.Elem) == 1 {
		// Deleting a path of length == 1 could delete an entire class of objects (i.e. all Device-Groups)
		// at once. The user probably doesn't want to do that.
		return fmt.Errorf("Refusing to delete path %s because it is too broad", gnmi.PathToString(path))
	}
	switch path.Elem[0].Name {
	case "vcs":
		return s.DeleteVcs(device, path)
	case "device-group":
		return s.DeleteDeviceGroup(device, path)
	}

	// It for something else.
	// We don't care.

	return nil
}
