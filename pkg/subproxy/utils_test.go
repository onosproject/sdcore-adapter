// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0
package subproxy

import (
	"bytes"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/onosproject/sdcore-adapter/pkg/test/mocks"
	"github.com/stretchr/testify/assert"

	//"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	//"strings"
	"testing"
	"time"
)

func TestExtractBaseURL(t *testing.T) {
	type args struct {
		baseURL string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"test1", args{baseURL: "http://config4g:5000/config"}, "http://config4g:5000/"},
		{"test2", args{baseURL: "https://config4g:5000/config"}, "https://config4g:5000/"},
		{"test3", args{baseURL: "http://webui:5000/config"}, "http://webui:5000/"},
		{"test4", args{baseURL: "https://webui:5000/config"}, "https://webui:5000/"},
		{"test5", args{baseURL: "http://webui:5000/config/test"}, "http://webui:5000/"},
		{"test6", args{baseURL: "https://webui:5000/config/test"}, "https://webui:5000/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractBaseURL(tt.args.baseURL); got != tt.want {
				t.Errorf("ExtractBaseURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestForwardReqToEndpoint(t *testing.T) {

	type args struct {
		postURI     string
		payload     []byte
		postTimeout time.Duration
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"4gurl",
			args{
				postURI:     "http://config4g:5000/api/subscriber/imsi-208014567891201",
				payload:     []byte("{\"plmnID\":\"20893\",\"ueId\":\"imsi-208014567891201\",\"OPc\":\"8e27b6af0e692e750f32667a3b14605d\",\"key\":\"8baf473f2f8fd09487cccbd7097c6862\",\"sequenceNumber\":\"16f3b3f70fc2\",\"DNN\": \"internet\"}"),
				postTimeout: 1 * time.Second,
			}, 201,
		},
		{"5gurl",
			args{
				postURI:     "http://webui:5000/api/subscriber/imsi-208014567891201",
				payload:     []byte("{\"plmnID\":\"20893\",\"ueId\":\"imsi-208014567891201\",\"OPc\":\"8e27b6af0e692e750f32667a3b14605d\",\"key\":\"8baf473f2f8fd09487cccbd7097c6862\",\"sequenceNumber\":\"16f3b3f70fc2\",\"DNN\": \"internet\"}"),
				postTimeout: 1 * time.Second,
			}, 201,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			respMock := ioutil.NopCloser(bytes.NewReader([]byte(`{"status":"success"}`)))
			httpMockClient := mocks.NewMockHTTPClient(ctrl)
			clientHTTP = httpMockClient

			httpMockClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {

				log.Infof(" from Http mock client ...%v", req.URL)
				assert.Equal(t, tt.args.postURI, fmt.Sprintf("%v", req.URL))
				return &http.Response{
					StatusCode: 201,
					Body:       respMock,
					Header:     make(http.Header),
				}, nil
			}).AnyTimes()

			got, err := ForwardReqToEndpoint(tt.args.postURI, tt.args.payload, tt.args.postTimeout)
			if err != nil {
				t.Errorf("ForwardReqToEndpoint() error = %v", err)
				return
			}
			resp, err := ioutil.ReadAll(got.Body)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.want, got.StatusCode)
		})
	}
}
