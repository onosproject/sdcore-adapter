// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

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
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
)

// GetEnterpriseObjectID return the object id for a path.
// If the path is a top-level model, then return the id of the object. If
// the object is not a top-level model then return nil.
func (s *Synchronizer) GetEnterpriseObjectID(scope *AetherScope, modelName string, path *pb.Path, keyName string) (*AetherScope, *string, error) {
	// Example path: /enterprises/enterprise[ent_id=sometthing]/site[site_id=something]/slice[id=something]
	if len(path.Elem) > 4 {
		// It's for some portion of the model, not the root of the model.
		// We don't care.
		return nil, nil, nil
	}

	// The first element better be "enterprises"
	if path.Elem[0].Name != "enterprises" {
		return nil, nil, nil
	}

	// Extract and lookup the enterprise
	if path.Elem[1].Name != "enterprise" {
		return nil, nil, nil
	}
	entID, okay := path.Elem[1].Key["ent-id"]
	if !okay {
		return nil, nil, fmt.Errorf("Delete of %s does not have an ent-id key", modelName)
	}
	enterprise, err := s.GetEnterprise(scope, &entID)
	if err != nil {
		return nil, nil, fmt.Errorf("Delete of %s failed to find enterprise %s", modelName, entID)
	}
	scope.Enterprise = enterprise

	// extract and lookup the site
	if path.Elem[2].Name != "site" {
		return nil, nil, nil
	}
	siteID, okay := path.Elem[2].Key["site-id"]
	if !okay {
		return nil, nil, fmt.Errorf("Delete of %s does not have a site-id key", modelName)
	}
	site, err := s.GetSite(scope, &siteID)
	if err != nil {
		return nil, nil, fmt.Errorf("Delete of %s failed to find site %s", modelName, siteID)
	}
	scope.Site = site

	if path.Elem[3].Name != modelName {
		// It's hard to imagine what else this could be. Future-proof by ignoring it.
		return nil, nil, nil
	}
	id, okay := path.Elem[3].Key[keyName]
	if !okay {
		return nil, nil, fmt.Errorf("Delete of %s does not have an id key", modelName)
	}

	return scope, &id, nil
}

// DeleteSlice deletes a Slice from the core
func (s *Synchronizer) DeleteSlice(scope *AetherScope, path *pb.Path) error {
	scope, id, err := s.GetEnterpriseObjectID(scope, "slice", path, "slice-id")
	if err != nil {
		return err
	}
	if id == nil {
		return nil
	}
	log.Infof("Delete slice %s", *id)

	_, err = s.GetSlice(scope, id)
	if err != nil {
		return err
	}

	// For each connectivity service, delete the VCS from the connectivity
	// service.
	csList, err := s.GetConnectivityServicesForEnterprise(scope)
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
				log.Infof("Tried to delete slice %s but it does not exist", *id)
				continue csLoop
			}
			return fmt.Errorf("Slice %s failed to push delete: %s", *id, err)
		}
	}

	// Remove slice from the cache
	s.CacheDelete(CacheModelSlice, *id)
	s.CacheDelete(CacheModelSliceUpf, *id)

	return nil
}

// DeleteDeviceGroup deletes a devicegroup from the core
func (s *Synchronizer) DeleteDeviceGroup(scope *AetherScope, path *pb.Path) error {
	scope, id, err := s.GetEnterpriseObjectID(scope, "device-group", path, "dg-id")
	if err != nil {
		return err
	}
	if id == nil {
		return nil
	}
	log.Infof("Delete device-group %s", *id)

	_, err = s.GetDeviceGroup(scope, id)
	if err != nil {
		return err
	}

	csList, err := s.GetConnectivityServicesForEnterprise(scope)
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
				log.Infof("Tried to delete slice %s but it does not exist", *id)
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
	rootDevice := config.(*RootDevice)

	scope := &AetherScope{RootDevice: rootDevice}

	if path == nil || len(path.Elem) == 0 {
		return nil
	}
	if len(path.Elem) < 4 {
		// Deleting a path of length < 4 could delete an entire class of objects (i.e. all Device-Groups)
		// at once. The user probably doesn't want to do that.
		return fmt.Errorf("Refusing to delete path %s because it is too broad", gnmi.PathToString(path))
	}
	switch path.Elem[3].Name {
	case "slice":
		return s.DeleteSlice(scope, path)
	case "device-group":
		return s.DeleteDeviceGroup(scope, path)
	}

	// It for something else.
	// We don't care.

	return nil
}
