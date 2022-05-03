// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
package gnmi

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto" //nolint: staticcheck
	gnmiproto "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/gnmi/value"
	"github.com/openconfig/ygot/ygot"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/openconfig/gnmi/proto/gnmi"

	// NOTE: This test case needs to have some models and modeldata in order
	// to run tests, so it gets these by using the sd-core synchronizers models.
	// TODO: It might be better to eventually switch to a service-independent
	// set of test models, so that this test code can remain independent of
	// any particular service.
	models "github.com/onosproject/aether-models/models/aether-2.1.x/api"
)

var ModelData = []*gnmiproto.ModelData{
	{Name: "access-profile", Organization: "Open Networking Foundation", Version: "2020-10-22"},
}

var (
	// model is the model for test config server.
	model = &Model{
		modelData:       ModelData,
		structRootType:  reflect.TypeOf((*models.Device)(nil)),
		schemaTreeRoot:  models.SchemaTree["Device"],
		jsonUnmarshaler: models.Unmarshal,
		enumData:        map[string]map[int64]ygot.EnumDefinition{},
	}
)

func TestCapabilities(t *testing.T) {
	s, err := NewServer(model, nil)
	if err != nil {
		t.Fatalf("error in creating server: %v", err)
	}
	resp, err := s.Capabilities(context.TODO(), &pb.CapabilityRequest{})
	if err != nil {
		t.Fatalf("got error %v, want nil", err)
	}
	if !reflect.DeepEqual(resp.GetSupportedModels(), model.modelData) {
		t.Errorf("got supported models %v\nare not the same as\nmodel supported by the server %v", resp.GetSupportedModels(), model.modelData)
	}
	if !reflect.DeepEqual(resp.GetSupportedEncodings(), supportedEncodings) {
		t.Errorf("got supported encodings %v\nare not the same as\nencodings supported by the server %v", resp.GetSupportedEncodings(), supportedEncodings)
	}
}

func TestGet(t *testing.T) {
	jsonConfigRoot, err := ioutil.ReadFile("./testdata/sample-config-root.json")
	assert.NoError(t, err)
	s, err := NewServer(model, nil)
	if err != nil {
		t.Fatalf("error in creating server: %v", err)
	}

	err = s.PutJSON("acme", jsonConfigRoot)
	assert.NoError(t, err)

	tds := []struct {
		desc        string
		textPbPath  string
		modelData   []*pb.ModelData
		wantRetCode codes.Code
		wantRespVal interface{}
	}{{
		desc: "get ip-domain",
		textPbPath: `
		  target: "acme"
			elem: <name: "site"
						key: <
							key:'site-id',
							value:'acme-site'
							  >
					   >						
			elem: <name: "ip-domain" 
						 key: <
							 key:'ip-domain-id',
							 value:'acme-chicago-ip'
							 >
						>
			elem: <name: "subnet">
		`,
		wantRetCode: codes.OK,
		wantRespVal: "163.25.44.0/31",
	},
		{
			desc: "get ip-domain with empty target",
			textPbPath: `
			elem: <name: "site"
						key: <
							key:'site-id',
							value:'acme-site'
							  >
					   >						
			elem: <name: "ip-domain" 
						 key: <
							 key:'ip-domain-id',
							 value:'acme-chicago-ip'
							 >
						>
			elem: <name: "subnet">
		`,
			wantRetCode: codes.InvalidArgument,
		}}

	for _, td := range tds {
		td := td // Linter: make shadow copy of range variable
		t.Run(td.desc, func(t *testing.T) {
			runTestGet(t, s, td.textPbPath, td.wantRetCode, td.wantRespVal, td.modelData)
		})
	}
}

// runTestGet requests a path from the server by Get grpc call, and compares if
// the return code and response value are expected.
func runTestGet(t *testing.T, s *Server, textPbPath string, wantRetCode codes.Code, wantRespVal interface{}, useModels []*pb.ModelData) {
	// Send request
	var pbPath pb.Path
	if err := proto.UnmarshalText(textPbPath, &pbPath); err != nil {
		t.Fatalf("error in unmarshaling path: %v", err)
	}
	req := &pb.GetRequest{
		Path:      []*pb.Path{&pbPath},
		Encoding:  pb.Encoding_JSON_IETF,
		UseModels: useModels,
	}
	resp, err := s.Get(req)

	// Check return code
	gotRetStatus, ok := status.FromError(err)
	if !ok {
		t.Fatal("got a non-grpc error from grpc call")
	}
	if gotRetStatus.Code() != wantRetCode {
		t.Fatalf("got return code %v, want %v", gotRetStatus.Code(), wantRetCode)
	}

	// Check response value
	var gotVal interface{}
	if resp != nil {
		notifs := resp.GetNotification()
		if len(notifs) != 1 {
			t.Fatalf("got %d notifications, want 1", len(notifs))
		}
		updates := notifs[0].GetUpdate()
		if len(updates) != 1 {
			t.Fatalf("got %d updates in the notification, want 1", len(updates))
		}
		val := updates[0].GetVal()
		if val.GetJsonIetfVal() == nil {
			gotVal, err = value.ToScalar(val)
			if err != nil {
				t.Errorf("got: %v, want a scalar value", gotVal)
			}
		} else {
			// Unmarshal json data to gotVal container for comparison
			if err := json.Unmarshal(val.GetJsonIetfVal(), &gotVal); err != nil {
				t.Fatalf("error in unmarshaling IETF JSON data to json container: %v", err)
			}
			var wantJSONStruct interface{}
			if err := json.Unmarshal([]byte(wantRespVal.(string)), &wantJSONStruct); err != nil {
				t.Fatalf("error in unmarshaling IETF JSON data to json container: %v", err)
			}
			wantRespVal = wantJSONStruct
		}
	}

	if !reflect.DeepEqual(gotVal, wantRespVal) {
		t.Errorf("got: %v (%T),\nwant %v (%T)", gotVal, gotVal, wantRespVal, wantRespVal)
	}
}

func TestSet(t *testing.T) {
	s, err := NewServer(model, nil)
	if err != nil {
		t.Fatalf("error in creating server: %v", err)
	}

	tds := []struct {
		desc         string
		textPbPrefix string
		textPbUpdate string
		modelData    []*pb.ModelData
		wantRetCode  codes.Code
		wantRespVal  interface{}
	}{{
		desc: "set ip-domain",
		textPbPrefix: `
		target: 'connectivity-service-v4'
		elem: <name: 'site'
				key: <
					key:'site-id',
					value:'acme-site'
		    >
		>			
		elem: <
			name: 'ip-domain'
			key:<
				key:'ip-domain-id'
				value:'ip-domain-demo-1'
			>
		>
		`,
		textPbUpdate: `
		path: <
			elem: <
				name: 'dns-primary'
			>
		>
		val: <
			string_val: '8.8.8.1'
		>
		`,
		wantRetCode: codes.OK,
	},
		{
			desc: "set ip-domain",
			textPbPrefix: `
			elem: <name: 'site'
					key: <
						key:'site-id',
						value:'acme-site'
					>
			>			
			elem: <
				name: 'ip-domain'
				key:<
					key:'ip-domain-id'
					value:'ip-domain-demo-1'
				>
			>
			`,
			textPbUpdate: `
			path: <
				elem: <
					name: 'dns-primary'
				>
			>
			val: <
				string_val: '8.8.8.1'
			>
		`,
			wantRetCode: codes.InvalidArgument,
		}}

	for _, td := range tds {
		td := td // Linter: make shadow copy of range variable
		t.Run(td.desc, func(t *testing.T) {
			runTestSet(t, s, td.textPbPrefix, td.textPbUpdate, td.wantRetCode, td.modelData)
		})
	}
}

// runTestGet requests a path from the server by Get grpc call, and compares if
// the return code and response value are expected.
func runTestSet(t *testing.T, s *Server, textPbPrefix string, textPbUpdate string, wantRetCode codes.Code, useModels []*pb.ModelData) {
	// Send request
	var pbPrefix pb.Path
	var pbUpdate pb.Update
	if err := proto.UnmarshalText(textPbPrefix, &pbPrefix); err != nil {
		t.Fatalf("error in unmarshaling path: %v", err)
	}
	if err := proto.UnmarshalText(textPbUpdate, &pbUpdate); err != nil {
		t.Fatalf("error in unmarshaling path: %v", err)
	}
	req := &pb.SetRequest{
		Prefix: &pbPrefix,
		Update: []*pb.Update{&pbUpdate},
	}
	_, err := s.Set(req)

	// Check return code
	gotRetStatus, ok := status.FromError(err)
	if !ok {
		t.Fatal("got a non-grpc error from grpc call")
	}
	if gotRetStatus.Code() != wantRetCode {
		t.Fatalf("got return code %v, want %v", gotRetStatus.Code(), wantRetCode)
	}
}

func TestServer_GetJSON(t *testing.T) {
	jsonConfigRoot, err := ioutil.ReadFile("./testdata/sample-config-root.json")
	assert.NoError(t, err)
	s, err := NewServer(model, nil)
	assert.NoError(t, err)
	err = s.PutJSON("acme", jsonConfigRoot)
	assert.NoError(t, err)
	jsonData, err := s.GetJSON("acme")
	assert.NoError(t, err)
	assert.NotNil(t, jsonData)
	require.JSONEq(t, string(jsonConfigRoot), string(jsonData))
}

func TestServer_PutJSON(t *testing.T) {
	s, err := NewServer(model, nil)
	assert.NoError(t, err)
	data := []byte("{}")
	err = s.PutJSON("acme", data)
	assert.NoError(t, err)
	acme, okay := s.config.Configs["acme"]
	assert.True(t, okay)
	assert.NotNil(t, acme)
}
