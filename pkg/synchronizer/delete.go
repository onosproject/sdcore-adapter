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
	"errors"
	"fmt"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

// GetEnterpriseObjectID returns the object id for a path.
// If the path is a top-level model, then return the id of the object. If
// the object is not a top-level model then return nil.
func (s *Synchronizer) GetEnterpriseObjectID(scope *AetherScope, modelName string, path *pb.Path, keyName string) (*AetherScope, *string, error) {
	// Example path: site[site_id=something]/slice[id=something]
	if len(path.Elem) > 2 {
		// It's for some portion of the model, not the root of the model.
		// We don't care.
		return nil, nil, nil
	}

	// Extract and lookup the site. If the first element is not site, then it's
	// not something we can delete.
	if path.Elem[0].Name != "site" {
		return nil, nil, fmt.Errorf("Path first element is not site, but is %s", path.Elem[0].Name)
	}
	siteID, okay := path.Elem[0].Key["site-id"]
	if !okay {
		return nil, nil, fmt.Errorf("Delete of %s does not have a site-id key", modelName)
	}
	site, err := s.GetSite(scope, &siteID)
	if err != nil {
		return nil, nil, fmt.Errorf("Delete of %s failed to find site %s", modelName, siteID)
	}
	scope.Site = site

	if modelName == "site" {
		// caller is asking for the site. We can stop now.
		return scope, &siteID, nil
	}

	if len(path.Elem) < 2 {
		// This should not be possible, as the caller will not have called us unless
		// it's looking at a path that is at least 2 elements long.
		return nil, nil, errors.New("Caller is asking for an object in path that is too short")
	}

	if path.Elem[1].Name != modelName {
		// It's hard to imagine what else this could be. Future-proof by ignoring it.
		return nil, nil, nil
	}
	id, okay := path.Elem[1].Key[keyName]
	if !okay {
		return nil, nil, fmt.Errorf("Delete of %s does not have an id key", modelName)
	}

	return scope, &id, nil
}

// deleteSlice deletes a Slice from the core
func (s *Synchronizer) deleteSliceByID(scope *AetherScope, id *string) error {
	log.Infof("Delete slice %s", *id)

	slice, err := s.GetSlice(scope, id)
	if err != nil {
		return err
	}

	// Fill in the CoreEndpoint
	err = s.updateScopeFromSlice(scope, slice)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v1/network-slice/%s", *scope.CoreEndpoint, *id)
	err = s.pusher.PushDelete(url)
	if err != nil {
		pushError, ok := err.(*PushError)
		if ok && pushError.StatusCode == 404 {
			// This may mean we already deleted it.
			log.Infof("Tried to delete slice %s but it does not exist", *id)
			// Fall through as success

		} else {
			return fmt.Errorf("Slice %s failed to push delete: %s", *id, err)
		}
	}

	// Remove slice from the cache
	s.CacheDelete(CacheModelSlice, *id)
	s.CacheDelete(CacheModelSliceUpf, *id)

	return nil
}

// deleteSlice deletes a Slice from the core, given a gNMI path
func (s *Synchronizer) deleteSliceByPath(scope *AetherScope, path *pb.Path) error {
	scope, id, err := s.GetEnterpriseObjectID(scope, "slice", path, "slice-id")
	if err != nil {
		return err
	}
	if id == nil {
		return nil
	}
	return s.deleteSliceByID(scope, id)
}

// deleteDeviceGroupByID deletes a devicegroup from the core, given an ID
func (s *Synchronizer) deleteDeviceGroupByID(scope *AetherScope, id *string) error {
	log.Infof("Delete device-group %s", *id)

	dg, err := s.GetDeviceGroup(scope, id)
	if err != nil {
		return err
	}

	// Fill in the CoreEndpoint
	err = s.updateScopeFromDeviceGroup(scope, dg)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v1/device-group/%s", *scope.CoreEndpoint, *id)
	err = s.pusher.PushDelete(url)
	if err != nil {
		pushError, ok := err.(*PushError)
		if ok && pushError.StatusCode == 404 {
			// This may mean we already deleted it.
			log.Infof("Tried to delete slice %s but it does not exist", *id)
			// Fall through as successful
		} else {
			return fmt.Errorf("Device-Group %s failed to push delete: %s", *id, err)
		}
	}

	// Remove device-group from the cache
	s.CacheDelete(CacheModelDeviceGroup, *id)

	return nil
}

// deleteDeviceGroupByPath deletes a devicegroup from the core, given a gNMI path
func (s *Synchronizer) deleteDeviceGroupByPath(scope *AetherScope, path *pb.Path) error {
	scope, id, err := s.GetEnterpriseObjectID(scope, "device-group", path, "dg-id")
	if err != nil {
		return err
	}
	if id == nil {
		return nil
	}
	return s.deleteDeviceGroupByID(scope, id)
}

// deleteSiteByPath deletes a site, the one that is part of the scope
func (s *Synchronizer) deleteSiteByScope(scope *AetherScope) error {
	log.Infof("Delete site %s", *scope.Site.SiteId)

	for dgID := range scope.Site.DeviceGroup {
		err := s.deleteDeviceGroupByID(scope, &dgID)
		if err != nil {
			return err
		}
	}

	for sliceID := range scope.Site.Slice {
		err := s.deleteSliceByID(scope, &sliceID)
		if err != nil {
			return err
		}
	}

	return nil
}

// deleteSiteByPath deletes a site from the core, given a gNMI path
func (s *Synchronizer) deleteSiteByPath(scope *AetherScope, path *pb.Path) error {
	scope, _, err := s.GetEnterpriseObjectID(scope, "site", path, "site-id")
	if err != nil {
		return err
	}

	return s.deleteSiteByScope(scope)
}

// HandleDelete synchronously performs a delete
func (s *Synchronizer) HandleDelete(config *gnmi.ConfigForest, path *pb.Path) error {
	if path == nil || len(path.Elem) == 0 {
		return errors.New("Delete of whole enterprise is not currently supported")
	}

	target := path.Target
	if target == "" {
		return errors.New("Refusing to handle delete without target specified")
	}

	rootDeviceInterface, okay := config.Configs[target]
	if !okay {
		log.Infof("Delete on target %s is for an empty tree", target)
		return nil
	}

	rootDevice := rootDeviceInterface.(*RootDevice)

	scope := &AetherScope{Enterprise: rootDevice}

	log.Infof("HandleDelete: %s", gnmi.PathToString(path))

	if len(path.Elem) < 1 {
		// To delete a site requires a 1-element path.
		// Less than 1 elements would be a delete of the whole tree. Refuse that.
		return fmt.Errorf("Refusing to delete path %s because it is too broad", gnmi.PathToString(path))
	}

	if path.Elem[0].Name != "site" {
		// It's not rooted at site. We don't care.
		return nil
	}

	if len(path.Elem) == 1 {
		// It must be the delete of an entire site
		return s.deleteSiteByPath(scope, path)
	}

	// At this point, length must be 4, it's some object inside of a site.

	switch path.Elem[1].Name {
	case "slice":
		return s.deleteSliceByPath(scope, path)
	case "device-group":
		return s.deleteDeviceGroupByPath(scope, path)
	}

	// It's for something else.
	// We don't care.

	return nil
}
