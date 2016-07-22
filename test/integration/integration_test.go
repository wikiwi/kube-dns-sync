/*
 * Copyright (C) 2016 wikiwi.io
 *
 * This software may be modified and distributed under the terms
 * of the MIT license. See the LICENSE file for details.
 */

package integration

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kr/pretty"

	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"k8s.io/kubernetes/federation/pkg/dnsprovider/rrstype"
	"k8s.io/kubernetes/pkg/api"

	"github.com/wikiwi/kube-dns-sync/pkg/controller"
	k8sutil "github.com/wikiwi/kube-dns-sync/pkg/util/kubernetes"
	"github.com/wikiwi/kube-dns-sync/pkg/util/kubernetes/dnsproviderfake"
)

type Test struct {
	ControllerOptions controller.Options
	Modify            func(c *controller.Controller)
	Expected          []dnsprovider.ResourceRecordSet
}

func (t Test) Run(rrs dnsprovider.ResourceRecordSets) {
	c, err := controller.New(&t.ControllerOptions)
	Expect(err).To(BeNil())
	report := make(chan struct{})
	go runAndReportExit(c, report)
	time.Sleep(1 * time.Second)
	if t.Modify != nil {
		t.Modify(c)
	}
	ls, err := rrs.List()
	Expect(err).To(BeNil())
	if !k8sutil.EqualRRSList(ls, t.Expected) {
		pretty.Fprintf(GinkgoWriter, "# Expected Value:\n%# v\n\n", t.Expected)
		pretty.Fprintf(GinkgoWriter, "# Received Value:\n%# v\n", ls)
		Fail("Unexpected DNS Records")
	}
	c.Stop()
	waitForReport(report)
}

var _ = Describe("Controller", func() {

	var client *kubeFake
	var dns *dnsproviderfake.Fake
	var rrs dnsprovider.ResourceRecordSets

	BeforeEach(func() {
		client = newKubeFake(k8sFixture...)
		dns = &dnsproviderfake.Fake{}
		zones, supported := dns.Zones()
		Expect(supported).To(BeTrue())
		zone, err := zones.New("test.com.")
		Expect(err).To(BeNil())
		_, err = zones.Add(zone)
		Expect(err).To(BeNil())
		rrs, supported = zone.ResourceRecordSets()
		Expect(supported).To(BeTrue())
	})

	It("should sync external IPs", func() {
		Test{
			Expected: []dnsprovider.ResourceRecordSet{
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 60, RRSDatas: []string{"1.1.1.1", "4.4.4.4"}, RRSType: rrstype.A},
			},
			ControllerOptions: controller.Options{
				DNSProvider:  dns,
				ZoneName:     "test.com.",
				Client:       client,
				TTL:          60,
				AddressTypes: []api.NodeAddressType{api.NodeExternalIP},
			},
		}.Run(rrs)
	})

	It("should sync internal IPs", func() {
		Test{
			Expected: []dnsprovider.ResourceRecordSet{
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "internalip.test.com.", RRSTTL: 60, RRSDatas: []string{"127.0.0.1", "127.0.0.4"}, RRSType: rrstype.A},
			},
			ControllerOptions: controller.Options{
				DNSProvider:  dns,
				ZoneName:     "test.com.",
				Client:       client,
				TTL:          60,
				AddressTypes: []api.NodeAddressType{api.NodeInternalIP},
			},
		}.Run(rrs)
	})

	It("should sync legacy IPs", func() {
		Test{
			Expected: []dnsprovider.ResourceRecordSet{
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "legacyhostip.test.com.", RRSTTL: 60, RRSDatas: []string{"2.2.2.2"}, RRSType: rrstype.A},
			},
			ControllerOptions: controller.Options{
				DNSProvider:  dns,
				ZoneName:     "test.com.",
				Client:       client,
				TTL:          60,
				AddressTypes: []api.NodeAddressType{api.NodeLegacyHostIP},
			},
		}.Run(rrs)
	})

	It("should sync different kind of addresses", func() {
		Test{
			Expected: []dnsprovider.ResourceRecordSet{
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 60, RRSDatas: []string{"1.1.1.1", "4.4.4.4"}, RRSType: rrstype.A},
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "internalip.test.com.", RRSTTL: 60, RRSDatas: []string{"127.0.0.1", "127.0.0.4"}, RRSType: rrstype.A},
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "legacyhostip.test.com.", RRSTTL: 60, RRSDatas: []string{"2.2.2.2"}, RRSType: rrstype.A},
			},
			ControllerOptions: controller.Options{
				DNSProvider:  dns,
				ZoneName:     "test.com.",
				Client:       client,
				TTL:          60,
				AddressTypes: []api.NodeAddressType{api.NodeExternalIP, api.NodeInternalIP, api.NodeLegacyHostIP},
			},
		}.Run(rrs)
	})

	It("should sync different kind TTL values", func() {
		Test{
			Expected: []dnsprovider.ResourceRecordSet{
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "legacyhostip.test.com.", RRSTTL: 200, RRSDatas: []string{"2.2.2.2"}, RRSType: rrstype.A},
			},
			ControllerOptions: controller.Options{
				DNSProvider:  dns,
				ZoneName:     "test.com.",
				Client:       client,
				TTL:          200,
				AddressTypes: []api.NodeAddressType{api.NodeLegacyHostIP},
			},
		}.Run(rrs)
	})

	It("should resync when DNS record changed out of band", func() {
		Test{
			Expected: []dnsprovider.ResourceRecordSet{
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 60, RRSDatas: []string{"1.1.1.1", "4.4.4.4"}, RRSType: rrstype.A},
			},
			ControllerOptions: controller.Options{
				DNSProvider:  dns,
				ZoneName:     "test.com.",
				Client:       client,
				TTL:          60,
				AddressTypes: []api.NodeAddressType{api.NodeExternalIP},
				SyncInterval: 500 * time.Millisecond,
			},
			Modify: func(c *controller.Controller) {
				rrs.Remove(&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSType: rrstype.A})
				rrs.Add(&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 200, RRSDatas: []string{"2.2.2.2", "4.4.4.4"}, RRSType: rrstype.A})
				time.Sleep(1 * time.Second)
			},
		}.Run(rrs)
	})

	It("should remove Node when it is becomes not ready", func() {
		Test{
			Expected: []dnsprovider.ResourceRecordSet{
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 60, RRSDatas: []string{"4.4.4.4"}, RRSType: rrstype.A},
			},
			ControllerOptions: controller.Options{
				DNSProvider:  dns,
				ZoneName:     "test.com.",
				Client:       client,
				TTL:          60,
				AddressTypes: []api.NodeAddressType{api.NodeExternalIP},
				SyncInterval: 500 * time.Millisecond,
			},
			Modify: func(c *controller.Controller) {
				client.ModifyNode(api.Node{
					ObjectMeta: api.ObjectMeta{Name: "node1"},
					Status: api.NodeStatus{
						Addresses: []api.NodeAddress{
							api.NodeAddress{Type: api.NodeExternalIP, Address: "1.1.1.1"},
							api.NodeAddress{Type: api.NodeInternalIP, Address: "127.0.0.1"},
						},
						Conditions: []api.NodeCondition{api.NodeCondition{
							Type:   api.NodeReady,
							Status: api.ConditionFalse,
						}},
					},
				})
				time.Sleep(500 * time.Millisecond)
			},
		}.Run(rrs)
	})

	It("should remove Node when it gets deleted", func() {
		Test{
			Expected: []dnsprovider.ResourceRecordSet{
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 60, RRSDatas: []string{"4.4.4.4"}, RRSType: rrstype.A},
			},
			ControllerOptions: controller.Options{
				DNSProvider:  dns,
				ZoneName:     "test.com.",
				Client:       client,
				TTL:          60,
				AddressTypes: []api.NodeAddressType{api.NodeExternalIP},
				SyncInterval: 500 * time.Millisecond,
			},
			Modify: func(c *controller.Controller) {
				client.DeleteNode("node1")
				time.Sleep(500 * time.Millisecond)
			},
		}.Run(rrs)
	})

	It("should add Node when one gets created", func() {
		Test{
			Expected: []dnsprovider.ResourceRecordSet{
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 60, RRSDatas: []string{"1.1.1.1", "4.4.4.4", "5.5.5.5"}, RRSType: rrstype.A},
			},
			ControllerOptions: controller.Options{
				DNSProvider:  dns,
				ZoneName:     "test.com.",
				Client:       client,
				TTL:          60,
				AddressTypes: []api.NodeAddressType{api.NodeExternalIP},
				SyncInterval: 500 * time.Millisecond,
			},
			Modify: func(c *controller.Controller) {
				client.AddNode(api.Node{
					ObjectMeta: api.ObjectMeta{Name: "node5"},
					Status: api.NodeStatus{
						Addresses: []api.NodeAddress{
							api.NodeAddress{Type: api.NodeExternalIP, Address: "5.5.5.5"},
						},
						Conditions: []api.NodeCondition{api.NodeCondition{
							Type:   api.NodeReady,
							Status: api.ConditionTrue,
						}},
					},
				})
				time.Sleep(500 * time.Millisecond)
			},
		}.Run(rrs)
	})

	It("should update IP when it changes", func() {
		Test{
			Expected: []dnsprovider.ResourceRecordSet{
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 60, RRSDatas: []string{"6.6.6.6", "4.4.4.4"}, RRSType: rrstype.A},
			},
			ControllerOptions: controller.Options{
				DNSProvider:  dns,
				ZoneName:     "test.com.",
				Client:       client,
				TTL:          60,
				AddressTypes: []api.NodeAddressType{api.NodeExternalIP},
				SyncInterval: 500 * time.Millisecond,
			},
			Modify: func(c *controller.Controller) {
				client.ModifyNode(api.Node{
					ObjectMeta: api.ObjectMeta{Name: "node1"},
					Status: api.NodeStatus{
						Addresses: []api.NodeAddress{
							api.NodeAddress{Type: api.NodeExternalIP, Address: "6.6.6.6"},
							api.NodeAddress{Type: api.NodeInternalIP, Address: "127.0.0.6"},
						},
						Conditions: []api.NodeCondition{api.NodeCondition{
							Type:   api.NodeReady,
							Status: api.ConditionTrue,
						}},
					},
				})
				time.Sleep(500 * time.Millisecond)
			},
		}.Run(rrs)
	})

	It("should sync to apex zone when configured", func() {
		Test{
			Expected: []dnsprovider.ResourceRecordSet{
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "test.com.", RRSTTL: 60, RRSDatas: []string{"1.1.1.1", "4.4.4.4"}, RRSType: rrstype.A},
			},
			ControllerOptions: controller.Options{
				DNSProvider:     dns,
				ZoneName:        "test.com.",
				Client:          client,
				TTL:             60,
				ApexAddressType: api.NodeExternalIP,
				SyncInterval:    500 * time.Millisecond,
			},
		}.Run(rrs)
	})

	It("should sync to apex zone and another address-type when configured", func() {
		Test{
			Expected: []dnsprovider.ResourceRecordSet{
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "internalip.test.com.", RRSTTL: 60, RRSDatas: []string{"127.0.0.1", "127.0.0.4"}, RRSType: rrstype.A},
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "test.com.", RRSTTL: 60, RRSDatas: []string{"1.1.1.1", "4.4.4.4"}, RRSType: rrstype.A},
			},
			ControllerOptions: controller.Options{
				DNSProvider:     dns,
				ZoneName:        "test.com.",
				Client:          client,
				TTL:             60,
				AddressTypes:    []api.NodeAddressType{api.NodeInternalIP},
				ApexAddressType: api.NodeExternalIP,
				SyncInterval:    500 * time.Millisecond,
			},
		}.Run(rrs)
	})

	It("should sync to apex zone and same address-type when configured", func() {
		Test{
			Expected: []dnsprovider.ResourceRecordSet{
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 60, RRSDatas: []string{"1.1.1.1", "4.4.4.4"}, RRSType: rrstype.A},
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "test.com.", RRSTTL: 60, RRSDatas: []string{"1.1.1.1", "4.4.4.4"}, RRSType: rrstype.A},
			},
			ControllerOptions: controller.Options{
				DNSProvider:     dns,
				ZoneName:        "test.com.",
				Client:          client,
				TTL:             60,
				AddressTypes:    []api.NodeAddressType{api.NodeExternalIP},
				ApexAddressType: api.NodeExternalIP,
				SyncInterval:    500 * time.Millisecond,
			},
		}.Run(rrs)
	})

	It("should not touch Resources of other Types", func() {
		Test{
			Expected: []dnsprovider.ResourceRecordSet{
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 60, RRSDatas: []string{"1.1.1.1", "4.4.4.4"}, RRSType: rrstype.A},
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "keepitCNAME.test.com.", RRSType: rrstype.CNAME},
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "keepitA.test.com.", RRSType: rrstype.A},
			},
			ControllerOptions: controller.Options{
				DNSProvider:  dns,
				ZoneName:     "test.com.",
				Client:       client,
				TTL:          60,
				AddressTypes: []api.NodeAddressType{api.NodeExternalIP},
				SyncInterval: 500 * time.Millisecond,
			},
			Modify: func(c *controller.Controller) {
				rrs.Add(&dnsproviderfake.ResourceRecordSetFake{RRSName: "keepitCNAME.test.com.", RRSType: rrstype.CNAME})
				rrs.Add(&dnsproviderfake.ResourceRecordSetFake{RRSName: "keepitA.test.com.", RRSType: rrstype.A})
				time.Sleep(600 * time.Millisecond)
			},
		}.Run(rrs)
	})

	It("should deal with sequence of changes", func() {
		Test{
			Expected: []dnsprovider.ResourceRecordSet{
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 60, RRSDatas: []string{"6.6.6.6", "4.4.4.4"}, RRSType: rrstype.A},
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "internalip.test.com.", RRSTTL: 60, RRSDatas: []string{"127.0.0.4", "127.0.0.6"}, RRSType: rrstype.A},
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "test.com.", RRSTTL: 60, RRSDatas: []string{"6.6.6.6", "4.4.4.4"}, RRSType: rrstype.A},
				&dnsproviderfake.ResourceRecordSetFake{RRSName: "keepit.test.com.", RRSType: rrstype.A},
			},
			ControllerOptions: controller.Options{
				DNSProvider:     dns,
				ZoneName:        "test.com.",
				Client:          client,
				TTL:             60,
				AddressTypes:    []api.NodeAddressType{api.NodeExternalIP, api.NodeInternalIP},
				SyncInterval:    500 * time.Millisecond,
				ApexAddressType: api.NodeExternalIP,
			},
			Modify: func(c *controller.Controller) {
				client.DeleteNode("node1")
				rrs.Add(&dnsproviderfake.ResourceRecordSetFake{RRSName: "keepit.test.com.", RRSType: rrstype.A})
				time.Sleep(500 * time.Millisecond)
				rrs.Remove(&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com."})
				time.Sleep(500 * time.Millisecond)
				client.AddNode(api.Node{
					ObjectMeta: api.ObjectMeta{Name: "node6"},
					Status: api.NodeStatus{
						Addresses: []api.NodeAddress{
							api.NodeAddress{Type: api.NodeExternalIP, Address: "6.6.6.6"},
							api.NodeAddress{Type: api.NodeInternalIP, Address: "127.0.0.6"},
						},
						Conditions: []api.NodeCondition{api.NodeCondition{
							Type:   api.NodeReady,
							Status: api.ConditionTrue,
						}},
					},
				})
				time.Sleep(500 * time.Millisecond)

			},
		}.Run(rrs)
	})
})
