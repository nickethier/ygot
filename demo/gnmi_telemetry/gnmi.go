// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Binary gnmi_telemetry provides an example application demonstrating the
// use of the ygot package to create gNMI telemetry notifications for use
// in the Subscribe RPC from a populated set of structs generated by ygen.
//
// The functionality in ygot supports both the pre-0.4.0 Path format for
// gNMI whereby the path consists of a slice of strings, as well as the
// PathElem format whereby the path is a set of structured elements.
package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/openconfig/ygot/ygot"

	log "github.com/golang/glog"
	oc "github.com/openconfig/ygot/exampleoc"

	gnmipb "github.com/openconfig/gnmi/proto/gnmi"
)

func main() {
	flag.Parse()
	d, err := CreateAFTInstance()
	if err != nil {
		log.Exitf("Error creating device instance: %v", err)
	}

	for _, e := range []bool{true, false} {
		g, err := renderToGNMINotifications(d, time.Now().Unix(), e)
		if err != nil {
			log.Exitf("Error creating notifications: %v", err)
		}

		if len(g) != 1 {
			log.Exitf("Unexpected number of notifications returned %s", len(g))
		}
		fmt.Printf("%v\n", proto.MarshalTextString(g[0]))
	}
}

// renderToGNMINotifications takes an input GoStruct and renders it to gNMI notifications. The
// timestamp is set to the ts argument. If usePathElem is set to true, the gNMI 0.4.0 path
// format is used.
func renderToGNMINotifications(s ygot.GoStruct, ts int64, usePathElem bool) ([]*gnmipb.Notification, error) {
	return ygot.TogNMINotifications(s, ts, ygot.GNMINotificationsConfig{UsePathElem: usePathElem})
}

// CreateAFTInstance creates an instance of the AFT model within a
// network instance and populates it with some example entries.
func CreateAFTInstance() (*oc.Device, error) {
	d := &oc.Device{}
	ni, err := d.NewNetworkInstance("DEFAULT")
	if err != nil {
		return nil, err
	}
	ni.Type = oc.OpenconfigNetworkInstanceTypes_NETWORK_INSTANCE_TYPE_DEFAULT_INSTANCE

	// Initialise the containers within the network instance model.
	ygot.BuildEmptyTree(ni)

	ip4, err := ni.Afts.NewIpv4Entry("192.0.2.1/32")
	if err != nil {
		return nil, err
	}

	nh4, err := ip4.NewNextHop(42)
	if err != nil {
		return nil, err
	}
	nh4.IpAddress = ygot.String("10.1.1.1")

	// The key to the MPLS list is a union, so we use of the generated
	// types for the interface that implements the union within NewLabelEntry.
	// Since these types have a single fied, then we can use the anonymous
	// initialiser.
	mpls, err := ni.Afts.NewLabelEntry(&oc.NetworkInstance_Afts_LabelEntry_Label_Union_Uint32{128})
	if err != nil {
		return nil, err
	}

	nh, err := mpls.NewNextHop(0)
	if err != nil {
		return nil, err
	}
	nh.IpAddress = ygot.String("192.0.2.1")

	// Each union has a method that is named To_X where the X is the union type, associated
	// with the struct that the union is within. This attempts to return the right type
	// based on the input interface.
	expNull, err := nh.To_NetworkInstance_Afts_LabelEntry_NextHop_PushedMplsLabelStack_Union(oc.OpenconfigAft_NextHop_PoppedMplsLabelStack_IPV4_EXPLICIT_NULL)
	if err != nil {
		return nil, err
	}
	nh.PushedMplsLabelStack = []oc.NetworkInstance_Afts_LabelEntry_NextHop_PushedMplsLabelStack_Union{
		&oc.NetworkInstance_Afts_LabelEntry_NextHop_PushedMplsLabelStack_Union_Uint32{42},
		&oc.NetworkInstance_Afts_LabelEntry_NextHop_PushedMplsLabelStack_Union_Uint32{84},
		expNull,
	}

	return d, nil
}
