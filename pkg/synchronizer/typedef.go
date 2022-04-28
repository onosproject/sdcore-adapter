// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizer

import (
	models "github.com/onosproject/aether-models/models/aether-2.1.x/api"
)

const ConnectivityService4G = models.OnfSite_Site_Slice_ConnectivityService_4g //nolint
const ConnectivityService5G = models.OnfSite_Site_Slice_ConnectivityService_5g //nolint

// Various typedefs to make modeling types more convenient throughout the synchronizer.

type Application = models.OnfApplication_Application                         //nolint
type ApplicationEndpoint = models.OnfApplication_Application_Endpoint        //nolint
type ApplicationEndpointMbr = models.OnfApplication_Application_Endpoint_Mbr //nolint
type RootDevice = models.Device                                              //nolint
type ConnectivityService = models.OnfSite_Site_ConnectivityService           //nolint
type Core4G = models.OnfSite_Site_ConnectivityService_Core_4G                //nolint
type Core5G = models.OnfSite_Site_ConnectivityService_Core_5G                //nolint
type Device = models.OnfSite_Site_Device                                     //nolint
type DeviceState = models.OnfSite_Site_Device_State                          //nolint
type DeviceGroup = models.OnfSite_Site_DeviceGroup                           //nolint
type DeviceGroupMbr = models.OnfSite_Site_DeviceGroup_Mbr                    //nolint
type DeviceGroupDevice = models.OnfSite_Site_DeviceGroup_Device              //nolint
type ImsiDefinition = models.OnfSite_Site_ImsiDefinition                     //nolint
type IpDomain = models.OnfSite_Site_IpDomain                                 //nolint
type SimCard = models.OnfSite_Site_SimCard                                   //nolint
type SmallCell = models.OnfSite_Site_SmallCell                               //nolint
type Site = models.OnfSite_Site                                              //nolint
type Template = models.OnfTemplate_Template                                  //nolint
type TrafficClass = models.OnfTrafficClass_TrafficClass                      //nolint
type Upf = models.OnfSite_Site_Upf                                           //nolint
type Slice = models.OnfSite_Site_Slice                                       //nolint
type SliceDeviceGroup = models.OnfSite_Site_Slice_DeviceGroup                //nolint
type SliceFilter = models.OnfSite_Site_Slice_Filter                          //nolint
type SliceMbr = models.OnfSite_Site_Slice_Mbr                                //nolint
