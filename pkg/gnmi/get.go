// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package gnmi

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	pb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/gnmi/value"
	"github.com/openconfig/ygot/util"
	"github.com/openconfig/ygot/ygot"
	"github.com/openconfig/ygot/ytypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Get implements the Get RPC in gNMI spec.
func (s *Server) Get(req *pb.GetRequest) (*pb.GetResponse, error) {

	dataType := req.GetType()

	tStart := time.Now()
	gnmiRequestsTotal.WithLabelValues("GET").Inc()

	if err := s.checkEncodingAndModel(req.GetEncoding(), req.GetUseModels()); err != nil {
		gnmiRequestsFailedTotal.WithLabelValues("GET").Inc()
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	prefix := req.GetPrefix()
	paths := req.GetPath()
	notifications := make([]*pb.Notification, len(paths))

	s.mu.RLock()
	defer s.mu.RUnlock()

	if paths == nil && dataType.String() != "" {

		jsonType := "IETF"
		if req.GetEncoding() == pb.Encoding_JSON {
			jsonType = "Internal"
		}
		notifications := make([]*pb.Notification, 1)
		path := pb.Path{}
		// Gets the whole config data tree
		node, err := ytypes.GetNode(s.model.schemaTreeRoot, s.config, &path)
		if isNil(node) || err != nil {
			gnmiRequestsFailedTotal.WithLabelValues("GET").Inc()
			return nil, status.Errorf(codes.NotFound, "path %s not found", path.String())
		}

		nodeStruct, _ := node[0].Data.(ygot.GoStruct)
		jsonTree, _ := ygot.ConstructIETFJSON(nodeStruct, &ygot.RFC7951JSONConfig{AppendModuleName: true})

		jsonTree = pruneConfigData(jsonTree, strings.ToLower(dataType.String()), &path).(map[string]interface{})
		jsonDump, err := json.Marshal(jsonTree)

		if err != nil {
			msg := fmt.Sprintf("error in marshaling %s JSON tree to bytes: %v", jsonType, err)
			log.Error(msg)
			gnmiRequestsFailedTotal.WithLabelValues("GET").Inc()
			return nil, status.Error(codes.Internal, msg)
		}
		ts := time.Now().UnixNano()

		update := buildUpdate(jsonDump, &path, jsonType)
		notifications[0] = &pb.Notification{
			Timestamp: ts,
			Prefix:    prefix,
			Update:    []*pb.Update{update},
		}
		resp := &pb.GetResponse{Notification: notifications}
		return resp, nil
	}

	for i, path := range paths {
		// Get schema node for path from config struct.
		fullPath := path
		if prefix != nil {
			fullPath = gnmiFullPath(prefix, path)
		}

		if fullPath.GetElem() == nil && fullPath.GetElement() != nil { // nolint:staticcheck
			gnmiRequestsFailedTotal.WithLabelValues("GET").Inc()
			return nil, status.Error(codes.Unimplemented, "deprecated path element type is unsupported")
		}

		nodes, err := ytypes.GetNode(s.model.schemaTreeRoot, s.config, fullPath)
		if len(nodes) == 0 || err != nil || util.IsValueNil(nodes[0].Data) {
			gnmiRequestsFailedTotal.WithLabelValues("GET").Inc()
			return nil, status.Errorf(codes.NotFound, "path %v not found: %v", fullPath, err)
		}
		node := nodes[0].Data

		ts := time.Now().UnixNano()

		nodeStruct, ok := node.(ygot.GoStruct)
		dataTypeFlag := false
		// Return leaf node.
		if !ok {
			elements := fullPath.GetElem()
			dataTypeString := strings.ToLower(dataType.String())
			if strings.Compare(dataTypeString, "all") == 0 {
				dataTypeFlag = true
			} else {
				for _, elem := range elements {
					if strings.Compare(dataTypeString, elem.GetName()) == 0 {
						dataTypeFlag = true
						break
					}

				}
			}
			if !dataTypeFlag {
				gnmiRequestsFailedTotal.WithLabelValues("GET").Inc()
				return nil, status.Error(codes.Internal, "The requested dataType is not valid")
			}
			var val *pb.TypedValue
			switch kind := reflect.ValueOf(node).Kind(); kind {
			case reflect.Ptr, reflect.Interface:
				var err error
				val, err = value.FromScalar(reflect.ValueOf(node).Elem().Interface())
				if err != nil {
					msg := fmt.Sprintf("leaf node %v does not contain a scalar type value: %v", path, err)
					log.Error(msg)
					gnmiRequestsFailedTotal.WithLabelValues("GET").Inc()
					return nil, status.Error(codes.Internal, msg)
				}
			case reflect.Int64:
				enumMap, ok := s.model.enumData[reflect.TypeOf(node).Name()]
				if !ok {
					gnmiRequestsFailedTotal.WithLabelValues("GET").Inc()
					return nil, status.Error(codes.Internal, "not a GoStruct enumeration type")
				}
				val = &pb.TypedValue{
					Value: &pb.TypedValue_StringVal{
						StringVal: enumMap[reflect.ValueOf(node).Int()].Name,
					},
				}
			case reflect.Slice:
				var err error
				switch kind := reflect.ValueOf(node).Kind(); kind {
				case reflect.Int64:
					//fmt.Println(reflect.TypeOf(node[0].Data).Elem())
					enumMap, ok := s.model.enumData[reflect.TypeOf(node).Name()]
					if !ok {
						gnmiRequestsFailedTotal.WithLabelValues("GET").Inc()
						return nil, status.Error(codes.Internal, "not a GoStruct enumeration type")
					}
					val = &pb.TypedValue{
						Value: &pb.TypedValue_StringVal{
							StringVal: enumMap[reflect.ValueOf(node).Int()].Name,
						},
					}
				default:
					val, err = value.FromScalar(reflect.ValueOf(node).Elem().Interface())
					if err != nil {
						msg := fmt.Sprintf("leaf node %v does not contain a scalar type value: %v", path, err)
						log.Error(msg)
						gnmiRequestsFailedTotal.WithLabelValues("GET").Inc()
						return nil, status.Error(codes.Internal, msg)
					}
				}
			default:
				gnmiRequestsFailedTotal.WithLabelValues("GET").Inc()
				return nil, status.Errorf(codes.Internal, "unexpected kind of leaf node type: %v %v", node, kind)
			}

			update := &pb.Update{Path: path, Val: val}
			notifications[i] = &pb.Notification{
				Timestamp: ts,
				Prefix:    prefix,
				Update:    []*pb.Update{update},
			}
			continue
		}
		dataTypeString := strings.ToLower(dataType.String())

		if req.GetUseModels() != nil {
			gnmiRequestsFailedTotal.WithLabelValues("GET").Inc()
			return nil, status.Errorf(codes.Unimplemented, "filtering Get using use_models is unsupported, got: %v", req.GetUseModels())
		}

		jsonType := "IETF"

		if req.GetEncoding() == pb.Encoding_JSON {
			jsonType = "Internal"
		}

		var jsonTree map[string]interface{}
		if reflect.ValueOf(nodeStruct).Pointer() == 0 {
			gnmiRequestsFailedTotal.WithLabelValues("GET").Inc()
			return nil, status.Error(codes.NotFound, "value is 0")

		}
		jsonTree, err = jsonEncoder(jsonType, nodeStruct)
		jsonTree = pruneConfigData(jsonTree, strings.ToLower(dataTypeString), fullPath).(map[string]interface{})
		if err != nil {
			msg := fmt.Sprintf("error in constructing %s JSON tree from requested node: %v", jsonType, err)
			log.Error(msg)
			gnmiRequestsFailedTotal.WithLabelValues("GET").Inc()
			return nil, status.Error(codes.Internal, msg)
		}

		jsonDump, err := json.Marshal(jsonTree)
		if err != nil {
			msg := fmt.Sprintf("error in marshaling %s JSON tree to bytes: %v", jsonType, err)
			log.Error(msg)
			gnmiRequestsFailedTotal.WithLabelValues("GET").Inc()
			return nil, status.Error(codes.Internal, msg)
		}

		update := buildUpdate(jsonDump, path, jsonType)
		notifications[i] = &pb.Notification{
			Timestamp: ts,
			Prefix:    prefix,
			Update:    []*pb.Update{update},
		}
	}
	resp := &pb.GetResponse{Notification: notifications}

	gnmiRequestDuration.WithLabelValues("GET").Observe(time.Since(tStart).Seconds())

	return resp, nil
}
