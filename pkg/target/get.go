// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package target

import (
	"github.com/google/gnxi/utils/credentials"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Get overrides the Get func of gnmi.Target to provide user auth.
func (s *target) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	//log.Infof("Incoming get request: %v", req)
	msg, ok := credentials.AuthorizeUser(ctx)
	if !ok {
		log.Infof("denied a Get request: %v", msg)
		return nil, status.Error(codes.PermissionDenied, msg)
	}

	log.Infof("allowed a Get request: %v", msg)
	return s.Server.Get(req)
}
