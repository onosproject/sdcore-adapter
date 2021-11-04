// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Cache implements a cache of data that is pushed to the core.

package synchronizerv4

import (
	"fmt"
	"reflect"
)

// CacheCheck returns true if (modelName, modelId) exists in the cache and the contents have not
// changed.
func (s *Synchronizer) CacheCheck(modelName string, modelID string, contents interface{}) bool {
	key := fmt.Sprintf("%s-%s", modelName, modelID)
	entry, okay := s.cache[key]
	if !okay {
		return false
	}

	return reflect.DeepEqual(entry, contents) // (entry == contents)
}

// CacheUpdate updates the contents of (modelName, modelID) in the cache with new contents
func (s *Synchronizer) CacheUpdate(modelName string, modelID string, contents interface{}) {
	// the unit tests don't initialize the map... Fix them and remove this.
	if s.cache == nil {
		s.cache = map[string]interface{}{}
	}

	key := fmt.Sprintf("%s-%s", modelName, modelID)
	s.cache[key] = contents
}

// CacheInvalidate removes all entries in the cache
func (s *Synchronizer) CacheInvalidate() {
	s.cache = map[string]interface{}{}
}
