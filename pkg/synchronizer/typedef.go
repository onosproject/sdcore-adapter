// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizer

import (
	models "github.com/onosproject/config-models/modelplugin/aether-2.0.0/aether_2_0_0"
)

// Various typedefs to make modeling types more convenient throughout the synchronizer.

type Application = models.OnfEnterprise_Enterprises_Enterprise_Application                           //nolint
type ApplicationEndpoint = models.OnfEnterprise_Enterprises_Enterprise_Application_Endpoint          //nolint
type ApplicationEndpointMbr = models.OnfEnterprise_Enterprises_Enterprise_Application_Endpoint_Mbr   //nolint
type ConnectivityService = models.OnfConnectivityService_ConnectivityServices_ConnectivityService    //nolint
type RootDevice = models.Device                                                                      //nolint
type Device = models.OnfEnterprise_Enterprises_Enterprise_Site_Device                                //nolint
type DeviceGroup = models.OnfEnterprise_Enterprises_Enterprise_Site_DeviceGroup                      //nolint
type DeviceGroupMbr = models.OnfEnterprise_Enterprises_Enterprise_Site_DeviceGroup_Mbr               //nolint
type DeviceGroupDevice = models.OnfEnterprise_Enterprises_Enterprise_Site_DeviceGroup_Device         //nolint
type Enterprise = models.OnfEnterprise_Enterprises_Enterprise                                        //nolint
type EnterpriseConnectivityService = models.OnfEnterprise_Enterprises_Enterprise_ConnectivityService //nolint
type ImsiDefinition = models.OnfEnterprise_Enterprises_Enterprise_Site_ImsiDefinition                //nolint
type IpDomain = models.OnfEnterprise_Enterprises_Enterprise_Site_IpDomain                            //nolint
type SimCard = models.OnfEnterprise_Enterprises_Enterprise_Site_SimCard                              //nolint
type SmallCell = models.OnfEnterprise_Enterprises_Enterprise_Site_SmallCell                          //nolint
type Site = models.OnfEnterprise_Enterprises_Enterprise_Site                                         //nolint
type Template = models.OnfEnterprise_Enterprises_Enterprise_Template                                 //nolint
type TrafficClass = models.OnfEnterprise_Enterprises_Enterprise_TrafficClass                         //nolint
type Upf = models.OnfEnterprise_Enterprises_Enterprise_Site_Upf                                      //nolint
type Slice = models.OnfEnterprise_Enterprises_Enterprise_Site_Slice                                  //nolint
type SliceDeviceGroup = models.OnfEnterprise_Enterprises_Enterprise_Site_Slice_DeviceGroup           //nolint
type SliceFilter = models.OnfEnterprise_Enterprises_Enterprise_Site_Slice_Filter                     //nolint
type SliceMbr = models.OnfEnterprise_Enterprises_Enterprise_Site_Slice_Mbr                           //nolint
