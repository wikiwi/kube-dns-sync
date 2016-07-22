/*
 * Copyright (C) 2016 wikiwi.io
 *
 * This software may be modified and distributed under the terms
 * of the MIT license. See the LICENSE file for details.
 */

package controller

import (
	"fmt"
	"strings"

	"github.com/kr/pretty"

	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"k8s.io/kubernetes/federation/pkg/dnsprovider/rrstype"
	"k8s.io/kubernetes/pkg/api"

	k8sutil "github.com/wikiwi/kube-dns-sync/pkg/util/kubernetes"
)

// sync starts the syncing process.
func (c *Controller) sync() error {
	c.log.Infof("Perform sync now")
	var zone dnsprovider.Zone

	c.log.Infof("Looking for Zone %q", c.zoneName)
	zones, supported := c.dns.Zones()
	if !supported {
		return fmt.Errorf("DNS Provider %q doesn't support Zones", c.dnsProvider)
	}

	zoneList, err := zones.List()
	if err != nil {
		return err
	}
	for _, x := range zoneList {
		if x.Name() == c.zoneName {
			zone = x
			break
		}
	}
	if zone == nil {
		return fmt.Errorf("Zone %q not found, waiting until one is created", c.zoneName)
	}

	rrs, supported := zone.ResourceRecordSets()
	if !supported {
		return fmt.Errorf("Zone %q doesn't support ResourceRecordSets", c.zoneName)
	}

	managedRecords := c.managedResourceRecordSets(rrs)
	return c.syncRecordSets(managedRecords, rrs)
}

// syncRecordSets will sync given list of RecordSets to the DNS Provider.
func (c *Controller) syncRecordSets(managedRecords []dnsprovider.ResourceRecordSet, rrs dnsprovider.ResourceRecordSets) error {
	c.log.Infof("Sync A Records")
	recordList, err := rrs.List()
	if err != nil {
		return err
	}
	for _, record := range managedRecords {
		create := true
		for _, x := range recordList {
			if x.Type() != rrstype.A {
				continue
			}
			if x.Name() == record.Name() {
				if !k8sutil.EqualRRS(x, record) {
					c.log.Infof("Remove diverged Record %q", record.Name())
					pretty.Pdiff(c.log, x, record)
					err := rrs.Remove(x)
					if err != nil {
						return err
					}
				} else {
					create = false
				}
			}
		}
		if create {
			c.log.Infof("Adding A Record %q", record.Name())
			_, err := rrs.Add(record)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// managedResourceRecordSets returns a list of managed ResourceRecordSets.
func (c *Controller) managedResourceRecordSets(rrs dnsprovider.ResourceRecordSets) []dnsprovider.ResourceRecordSet {
	var nodes []*api.Node
	for _, x := range c.cache.List() {
		nodes = append(nodes, x.(*api.Node))
	}

	addressTypes := c.addressTypes
	apexInGroup := false
	if c.apexAddressType != "" {
		for _, x := range c.addressTypes {
			if c.apexAddressType == x {
				apexInGroup = true
				break
			}
		}
		if !apexInGroup {
			addressTypes = append(addressTypes, c.apexAddressType)
		}
	}

	sets := []dnsprovider.ResourceRecordSet{}
	for _, addressType := range addressTypes {
		typeString := strings.ToLower(string(addressType))
		groupAddresses := []string{}
		for _, node := range nodes {
			if !k8sutil.IsNodeReady(node) {
				continue
			}
			addresses := []string{}
			for _, x := range node.Status.Addresses {
				if x.Type == addressType {
					addresses = append(addresses, x.Address)
				}
			}
			if len(addresses) == 0 {
				continue
			}
			groupAddresses = append(groupAddresses, addresses...)
		}
		if len(groupAddresses) == 0 {
			continue
		}
		if addressType == c.apexAddressType {
			record := rrs.New(c.zoneName, groupAddresses, c.ttl, rrstype.A)
			sets = append(sets, record)
		}
		if addressType != c.apexAddressType || apexInGroup {
			name := typeString + "." + c.zoneName
			record := rrs.New(name, groupAddresses, c.ttl, rrstype.A)
			sets = append(sets, record)
		}
	}
	return sets
}
