/*
 * Copyright (C) 2016 wikiwi.io
 *
 * This software may be modified and distributed under the terms
 * of the MIT license. See the LICENSE file for details.
 */

package dnsproviderfake

import (
	"fmt"

	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"k8s.io/kubernetes/federation/pkg/dnsprovider/rrstype"
)

var _ dnsprovider.Interface = new(Fake)
var _ dnsprovider.Zones = new(ZonesFake)
var _ dnsprovider.Zone = new(ZoneFake)
var _ dnsprovider.ResourceRecordSets = new(ResourceRecordSetsFake)
var _ dnsprovider.ResourceRecordSet = new(ResourceRecordSetFake)

// Fake is a fake dns provider.
type Fake struct {
	ZonesFake ZonesFake
}

// Zones returns ZonesFake.
func (f *Fake) Zones() (dnsprovider.Zones, bool) {
	return &f.ZonesFake, true
}

// ZonesFake is a fake of Zones.
type ZonesFake struct {
	ZoneList []dnsprovider.Zone
}

// List of added zones.
func (f *ZonesFake) List() ([]dnsprovider.Zone, error) {
	return f.ZoneList, nil
}

// Add zone to list.
func (f *ZonesFake) Add(z dnsprovider.Zone) (dnsprovider.Zone, error) {
	f.ZoneList = append(f.ZoneList, z)
	return z, nil
}

// Remove zone from list.
func (f *ZonesFake) Remove(z dnsprovider.Zone) error {
	for i, x := range f.ZoneList {
		if z.Name() == x.Name() {
			f.ZoneList = append(f.ZoneList[:i], f.ZoneList[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("zone %q not found", z.Name())
}

// New creates a new ZoneFake.
func (f *ZonesFake) New(name string) (dnsprovider.Zone, error) {
	return &ZoneFake{ZoneName: name, RRS: new(ResourceRecordSetsFake)}, nil
}

// ZoneFake is a fake implementation of Zone.
type ZoneFake struct {
	ZoneName string
	RRS      *ResourceRecordSetsFake
}

// Name returns name of zone.
func (f *ZoneFake) Name() string {
	return f.ZoneName
}

// ResourceRecordSets returns ResourceRecordSetsFake.
func (f *ZoneFake) ResourceRecordSets() (dnsprovider.ResourceRecordSets, bool) {
	return f.RRS, true
}

// ResourceRecordSetsFake fake implementation of ResourceRecordSets.
type ResourceRecordSetsFake struct {
	RRSList []dnsprovider.ResourceRecordSet
}

// List returns list of Resource Record Sets.
func (f *ResourceRecordSetsFake) List() ([]dnsprovider.ResourceRecordSet, error) {
	return f.RRSList, nil
}

// Add Resource Record Set to list.
func (f *ResourceRecordSetsFake) Add(rrs dnsprovider.ResourceRecordSet) (dnsprovider.ResourceRecordSet, error) {
	f.RRSList = append(f.RRSList, rrs)
	return rrs, nil
}

// Remove Resource Record Set from list.
func (f *ResourceRecordSetsFake) Remove(rrs dnsprovider.ResourceRecordSet) error {
	for i, x := range f.RRSList {
		if rrs.Name() == x.Name() {
			f.RRSList = append(f.RRSList[:i], f.RRSList[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Resource Record Set %q not found", rrs.Name())
}

// New creates instance of ResourceRecordSetFake.
func (f *ResourceRecordSetsFake) New(name string, rrdatas []string, ttl int64, rrstype rrstype.RrsType) dnsprovider.ResourceRecordSet {
	return &ResourceRecordSetFake{
		RRSName: name, RRSDatas: rrdatas, RRSTTL: ttl, RRSType: rrstype,
	}
}

// ResourceRecordSetFake is a fake implementation of ResourceRecordSet.
type ResourceRecordSetFake struct {
	RRSName  string
	RRSDatas []string
	RRSTTL   int64
	RRSType  rrstype.RrsType
}

// Name returns name of Resource Record Set.
func (f *ResourceRecordSetFake) Name() string {
	return f.RRSName
}

// Rrdatas returns datas of Resource Record Set.
func (f *ResourceRecordSetFake) Rrdatas() []string {
	return f.RRSDatas
}

// Ttl returns TTL of Resource Record Set.
func (f *ResourceRecordSetFake) Ttl() int64 {
	return f.RRSTTL
}

// Type returns type of Resource Record Set.
func (f *ResourceRecordSetFake) Type() rrstype.RrsType {
	return f.RRSType
}
