// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

/*
 * The main entry point for the migration engine. Provides functions to create new migrators
 * and steps, and to execute migration.
 */

package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"github.com/onosproject/sdcore-adapter/pkg/gnmiclient"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"io/ioutil"
	"strconv"
)

var log = logging.GetLogger("migration")

// AddMigrationStep adds a migration step to the list
func (m *Migrator) AddMigrationStep(fromVersion string, fromModels *gnmi.Model, toVersion string, toModels *gnmi.Model, migrationFunc MigrationFunction) {
	step := MigrationStep{
		FromVersion:   fromVersion,
		FromModels:    fromModels,
		ToVersion:     toVersion,
		ToModels:      toModels,
		MigrationFunc: migrationFunc,
		Migrator:      m,
	}
	m.steps = append(m.steps, &step)
}

// buildStepList builds a list of migration steps that starts with "fromVersion" and ends with "toVersion".
// Steps must form a contiguous list of migrations -- each step must yield the migration that is used
// by the following migration.
func (m *Migrator) buildStepList(fromVersion string, toVersion string) ([]*MigrationStep, error) {
	steps := []*MigrationStep{}
	currentVersion := fromVersion

	if currentVersion == toVersion {
		//  there's nothing to do!
		return steps, nil
	}

	for _, step := range m.steps {
		if step.FromVersion == currentVersion {
			steps = append(steps, step)
			currentVersion = step.ToVersion
		}
		if currentVersion == toVersion {
			break
		}
	}

	if len(steps) == 0 {
		return nil, fmt.Errorf("Unable to find a step that started with version %s", fromVersion)
	}

	if currentVersion != toVersion {
		return nil, fmt.Errorf("Unable to find a step that ended with version %s", toVersion)
	}

	return steps, nil
}

// runStep runs a migration step
func (m *Migrator) runStep(step *MigrationStep, fromTarget string, toTarget string) ([]*MigrationActions, error) {
	// fetch the old models
	srcVal, err := m.Gnmi.GetPath(context.Background(), "", fromTarget, m.Gnmi.Address())
	if err != nil {
		return nil, err
	}
	// TODO: handle srcVal == nil

	// fetch the new models
	destVal, err := m.Gnmi.GetPath(context.Background(), "", toTarget, m.Gnmi.Address())
	if err != nil {
		return nil, err
	}
	// TODO: handle destVal == nil

	// execute the function to migrate items from old to new
	actions, err := step.MigrationFunc(step, fromTarget, toTarget, srcVal, destVal)
	if err != nil {
		return nil, err
	}

	return actions, nil
}

// executeActions executes a list of actions
func (m *Migrator) executeActions(actions []*MigrationActions, fromTarget string, toTarget string) error {
	// do the updates in forward order
	for _, action := range actions {
		if len(action.Updates) == 0 {
			if len(action.Deletes) == 0 {
				return errors.NewInternal("action has neither Updates nor Deletes",
					action.UpdatePrefix.String(), action.DeletePrefix.String())
			}
			continue
		}
		err := m.Gnmi.Update(context.Background(), action.UpdatePrefix, toTarget, m.Gnmi.Address(), action.Updates)
		if err != nil {
			return err
		}
	}

	// now do the deletes in reverse order
	for i := len(actions) - 1; i >= 0; i-- {
		action := actions[i]
		if len(action.Deletes) == 0 {
			continue
		}
		err := m.Gnmi.Delete(context.Background(), action.DeletePrefix, fromTarget, m.Gnmi.Address(), action.Deletes)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Migrator) outputActions(actions []*MigrationActions,
	fromTarget string, toTarget string, fromVersion string, toVersion string) ([]byte, error) {

	topLevel := make(map[string]interface{})
	topLevel["default-target"] = toTarget
	updateModel := make(map[string]interface{})
	topLevel["Updates"] = updateModel
	deleteModel := make(map[string]interface{})
	topLevel["Deletes"] = deleteModel
	topLevel["Extensions"] = map[string]string{
		"model-version-101": "3.0.0",
		"model-type-102":    "Aether",
	}

	for _, action := range actions {
		updateModelIf := interface{}(updateModel)
		updatePrefix, err := addObject(&updateModelIf, action.UpdatePrefix.GetElem(), true, nil, "", toVersion)
		if err != nil {
			return nil, err
		}
		for _, u := range action.Updates {
			if _, err = addObject(updatePrefix, u.GetPath().GetElem(), false, u.GetVal(), "",
				keepSuffix(len(action.UpdatePrefix.GetElem()), toVersion)); err != nil {
				return nil, err
			}
		}
		deleteModelIf := interface{}(deleteModel)
		deletePrefix, err := addObject(&deleteModelIf, action.DeletePrefix.GetElem(), true, nil, fromTarget, fromVersion)
		if err != nil {
			return nil, err
		}
		for _, u := range action.Deletes {
			if _, err = addObject(deletePrefix, u.GetElem(), true, nil, fromTarget,
				keepSuffix(len(action.DeletePrefix.GetElem()), fromVersion)); err != nil {
				return nil, err
			}
		}
	}

	return json.MarshalIndent(topLevel, "", "  ")
}

// addObject - recursive function to build up struct which can be marshalled to JSON
func addObject(nodeif *interface{}, elems []*gpb.PathElem, isPrefix bool,
	val *gpb.TypedValue, target string, suffix string) (*interface{}, error) {

	if len(elems) == 0 {
		return nodeif, nil
	}
	log.Debugf("Handling %s %s %v", target, suffix, elems)
	// Convert to its real type
	nodemap, ok := (*nodeif).(map[string]interface{})
	if !ok {
		// Might be a slice
		nodeslice, ok := (*nodeif).([]interface{})
		if !ok {
			return nil, errors.NewInternal("could not convert nodeif %v for %v", *nodeif, elems)
		}
		if nodemap, ok = nodeslice[len(nodeslice)-1].(map[string]interface{}); !ok {
			return nil, errors.NewInternal("error decoding - expecting map", nodeslice, elems)
		}
	}
	name := elems[0].Name
	if suffix != "" {
		name = fmt.Sprintf("%s-%s", elems[0].Name, suffix)
	}

	if len(elems) == 1 && !isPrefix { // Last item
		fieldDescriptor := val.ProtoReflect().WhichOneof(val.ProtoReflect().Descriptor().Oneofs().ByName("value"))

		switch fieldDescriptor.Kind().String() {
		case "uint8", "uint16", "uint32", "uint64":
			nodemap[name] = val.GetUintVal()
		case "bool":
			nodemap[name] = val.GetBoolVal()
		case "string":
			nodemap[name] = val.GetStringVal()
		case "float32", "float64":
			nodemap[name] = val.GetFloatVal()
		default:
			log.Fatal("unhandled type", fieldDescriptor.Kind().String())
		}
		return nil, nil
	}

	child, ok := nodemap[name]
	if elems[0].Key == nil {
		var childMap map[string]interface{}
		if ok {
			if childMap, ok = (child).(map[string]interface{}); !ok {
				return nil, errors.NewInternal("unknown type %v", child)
			}
		} else {
			childMap = make(map[string]interface{})
			nodemap[name] = childMap
		}
		if target != "" {
			childMap["additionalProperties"] = map[string]string{
				"target": target,
			}
		}
		childMapIf := interface{}(childMap)
		if len(elems) > 1 {
			return addObject(&childMapIf, elems[1:], isPrefix, val, "", "")
		}
		return &childMapIf, nil
	}
	var childInstance map[string]interface{}
	var childSlice []interface{}
	if ok {
		if childSlice, ok = (child).([]interface{}); !ok {
			return nil, errors.NewInternal("unknown type %v", child)
		}
		// Does the particular instance we want exit on the childSlice?
		for _, childInst := range childSlice {
			childMap, ok := childInst.(map[string]interface{})
			if !ok {
				return nil, errors.NewInternal("unexpected type for child %v", childInst)
			}
			matchingKeys := 0
			for k, v := range elems[0].Key {
				// It might be a numeric identifier
				var value interface{}
				vInt, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					value = v
				} else {
					value = vInt
				}
				if childKey, ok := childMap[k]; ok {
					if childKey == value {
						matchingKeys++
					}
				}
			}
			if matchingKeys == len(elems[0].Key) {
				childInstance = childMap
			}
		}
	} else {
		childSlice = make([]interface{}, 0, 1)
		nodemap[name] = childSlice
	}
	if childInstance == nil {
		childInstance = make(map[string]interface{})
		childSlice = append(childSlice, childInstance)
		nodemap[name] = childSlice // Reattach new slice
		for k, v := range elems[0].Key {
			// It might be a numeric identifier
			vInt, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				childInstance[k] = v
			} else {
				childInstance[k] = vInt
			}
		}
	}
	if len(elems) > 1 {
		childInstanceIf := interface{}(childInstance)
		return addObject(&childInstanceIf, elems[1:], isPrefix, val, "", "")
	}
	childSliceIf := interface{}(childSlice)
	return &childSliceIf, nil
}

func keepSuffix(elemsLen int, suffix string) string {
	if elemsLen > 0 {
		return ""
	}
	return suffix
}

// Migrate performs migration from one version to another
func (m *Migrator) Migrate(fromTarget string, fromVersion string, toTarget string, toVersion string, outToGnmi *bool, output *string) error {
	steps, err := m.buildStepList(fromVersion, toVersion)
	if err != nil {
		return err
	}

	for _, step := range steps {
		actions, errStep := m.runStep(step, fromTarget, toTarget)
		if errStep != nil {
			return errStep
		}

		if outToGnmi != nil && *outToGnmi {
			errStep = m.executeActions(actions, fromTarget, toTarget)
		} else {
			var json []byte
			json, errStep = m.outputActions(actions, fromTarget, toTarget, fromVersion, toVersion)
			if output != nil && *output != "" {
				err = ioutil.WriteFile(*output, json, 0644)
				if err != nil {
					log.Fatalf("error writing generated code to file: %s\n", err)
				}
				log.Infof("Output written to %s", *output)
			} else {
				fmt.Println(string(json))
			}
		}
		if errStep != nil {
			return errStep
		}
	}

	return nil
}

// NewMigrator creates a new migrator
func NewMigrator(gnmiClient gnmiclient.GnmiInterface) *Migrator {
	m := &Migrator{
		AetherConfigAddr: "",
		Gnmi:             gnmiClient,
	}

	return m
}

// SupportedVersions - list the versions supported
func (m *Migrator) SupportedVersions() []string {
	stepsSupported := make([]string, 0)
	for _, s := range m.steps {
		stepsSupported = append(stepsSupported, fmt.Sprintf("%s to %s", s.FromVersion, s.ToVersion))
	}
	return stepsSupported
}
