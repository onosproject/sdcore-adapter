// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package subproxy

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/test/mocks"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var sp = subscriberProxy{
	AetherConfigAddress:   "onos-config.micro-onos.svc.cluster.local:5150",
	BaseWebConsoleURL:     "http://webui.omec.svc.cluster.local:5000",
	AetherConfigTarget:    "connectivity-service-v3",
	gnmiClient:            nil,
	PostTimeout:           0,
	retryInterval:         0,
	busy:                  0,
	synchronizeDeviceFunc: nil,
}

func TestMain(m *testing.M) {
	log := logging.GetLogger("subscriber-proxy")
	log.SetLevel(logging.DebugLevel)
	clientHTTP = &mocks.MockHTTPClient{}
	os.Exit(m.Run())
}

func TestSubscriberProxy_addSubscriberByID(t *testing.T) {

	dataJSON, err := ioutil.ReadFile("./testdata/testData.json")
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	gnmiMockClient := mocks.NewMockGnmiInterface(ctrl)
	sp.gnmiClient = gnmiMockClient

	gnmiMockClient.EXPECT().GetPath(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, path string, target string, addr string) (*gpb.TypedValue, error) {
			return &gpb.TypedValue{
				Value: &gpb.TypedValue_JsonVal{JsonVal: dataJSON},
			}, nil
		}).AnyTimes()

	var updSetRequests []*gpb.SetRequest
	gnmiMockClient.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, prefix *gpb.Path, target string, addr string, updates []*gpb.Update) error {
			updSetRequests = append(updSetRequests, &gpb.SetRequest{
				Update: updates,
			})
			return nil
		}).AnyTimes()

	respMock := ioutil.NopCloser(bytes.NewReader([]byte(`{}`)))

	httpMockClient := mocks.NewMockHTTPClient(ctrl)

	httpMockClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 201,
			Body:       respMock,
		}, nil
	}).AnyTimes()

	w := httptest.NewRecorder()
	router := gin.New()
	router.Use(getlogger(), gin.Recovery())
	router.POST("/api/subscriber/:ueId", sp.addSubscriberByID)
	payload := strings.NewReader(`{` + "" + `"plmnID": "26512",` + "" + `"ueId": "imsi-111222333444555",` + "" + `
	"OPc": "8e27b6af0e692e750f32667a3b14605d",` + "" + `"key": "8baf473f2f8fd09487cccbd7097c6862",` + "" + `
	"sequenceNumber": "16f3b3f70fc2",` + "" + `"DNN": "internet "` + "" + `}`)
	req, err := http.NewRequest("POST", "/api/subscriber/imsi-111222333444555", payload)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		assert.Equal(t, 201, w.Code)
	}

	resp, err := ioutil.ReadAll(w.Body)
	if err != nil {
		assert.NoError(t, err)
	}
	assert.Equal(t, "{\"status\":\"success\"}", string(resp))

}

func TestSubscriberProxy_updateImsiDeviceGroup(t *testing.T) {

	dataJSON, err := ioutil.ReadFile("./testdata/testData.json")
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	gnmiMockClient := mocks.NewMockGnmiInterface(ctrl)
	sp.gnmiClient = gnmiMockClient

	gnmiMockClient.EXPECT().GetPath(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, path string, target string, addr string) (*gpb.TypedValue, error) {
			return &gpb.TypedValue{
				Value: &gpb.TypedValue_JsonVal{JsonVal: dataJSON},
			}, nil
		}).AnyTimes()

	var updSetRequests []*gpb.SetRequest
	gnmiMockClient.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, prefix *gpb.Path, target string, addr string, updates []*gpb.Update) error {
			updSetRequests = append(updSetRequests, &gpb.SetRequest{
				Update: updates,
			})
			return nil
		}).AnyTimes()

	// IMSI will be added to default device group under default site
	imsiValue := uint64(111222333444555)
	err = sp.updateImsiDeviceGroup(&imsiValue)
	assert.NoError(t, err)
	assert.NotNil(t, updSetRequests)
	assert.Len(t, updSetRequests, 1)

	//IMSI already exist in device group under default site
	imsiValue = uint64(21032002000010)
	err = sp.updateImsiDeviceGroup(&imsiValue)
	assert.NoError(t, err)
	assert.NotNil(t, updSetRequests)
	assert.Len(t, updSetRequests, 1)

	// IMSI will be added to device group under default site
	imsiValue = uint64(265122002000035)
	err = sp.updateImsiDeviceGroup(&imsiValue)
	assert.NoError(t, err)
	assert.NotNil(t, updSetRequests)
	assert.Len(t, updSetRequests, 2)

	//IMSI exist in device group under site
	imsiValue = uint64(21032002000040)
	err = sp.updateImsiDeviceGroup(&imsiValue)
	assert.NoError(t, err)
	assert.NotNil(t, updSetRequests)
	assert.Len(t, updSetRequests, 2)

}
