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

var _ = Describe("Controller", func() {

	var client *kubeFake
	var dns *dnsproviderfake.Fake
	var rrs dnsprovider.ResourceRecordSets
	var report chan struct{}

	BeforeEach(func() {
		report = make(chan struct{})
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
		for _, x := range dnsFixture {
			rrs.Add(&x)
		}
	})

	It("should remove unknown records", func() {
		c, err := controller.New(&controller.Options{
			DNSProvider:  dns,
			ZoneName:     "test.com.",
			Client:       client,
			AddressTypes: []api.NodeAddressType{api.NodeExternalIP},
		})
		Expect(err).To(BeNil())
		go runAndReportExit(c, report)
		time.Sleep(1 * time.Second)
		ls, err := rrs.List()
		Expect(err).To(BeNil())
		for _, x := range ls {
			Expect(x.Name()).NotTo(Equal("garbage"))
		}
		c.Stop()
		waitForReport(report)
	})

	It("should sync external IPs", func() {
		expected := []dnsprovider.ResourceRecordSet{
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node4.externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"4.4.4.4"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node1.externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"1.1.1.1"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"1.1.1.1", "4.4.4.4"}, RRSType: rrstype.A},
		}
		c, err := controller.New(&controller.Options{
			DNSProvider:  dns,
			ZoneName:     "test.com.",
			Client:       client,
			TTL:          300,
			AddressTypes: []api.NodeAddressType{api.NodeExternalIP},
		})
		Expect(err).To(BeNil())
		go runAndReportExit(c, report)
		time.Sleep(1 * time.Second)
		ls, err := rrs.List()
		Expect(err).To(BeNil())
		if !k8sutil.EqualRRSList(ls, expected) {
			pretty.Fprintf(GinkgoWriter, "# Expected Value:\n%# v\n\n", expected)
			pretty.Fprintf(GinkgoWriter, "# Received Value:\n%# v\n", ls)
			Fail("Unexpected DNS Records")
		}
		c.Stop()
		waitForReport(report)
	})

	It("should sync internal IPs", func() {
		expected := []dnsprovider.ResourceRecordSet{
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node4.internalip.test.com.", RRSTTL: 300, RRSDatas: []string{"127.0.0.4"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node1.internalip.test.com.", RRSTTL: 300, RRSDatas: []string{"127.0.0.1"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "internalip.test.com.", RRSTTL: 300, RRSDatas: []string{"127.0.0.1", "127.0.0.4"}, RRSType: rrstype.A},
		}
		c, err := controller.New(&controller.Options{
			DNSProvider:  dns,
			ZoneName:     "test.com.",
			Client:       client,
			TTL:          300,
			AddressTypes: []api.NodeAddressType{api.NodeInternalIP},
		})
		Expect(err).To(BeNil())
		go runAndReportExit(c, report)
		time.Sleep(1 * time.Second)
		ls, err := rrs.List()
		Expect(err).To(BeNil())
		if !k8sutil.EqualRRSList(ls, expected) {
			pretty.Fprintf(GinkgoWriter, "# Expected Value:\n%# v\n\n", expected)
			pretty.Fprintf(GinkgoWriter, "# Received Value:\n%# v\n", ls)
			Fail("Unexpected DNS Records")
		}
		c.Stop()
		waitForReport(report)
	})

	It("should sync legacy IPs", func() {
		expected := []dnsprovider.ResourceRecordSet{
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node2.legacyhostip.test.com.", RRSTTL: 300, RRSDatas: []string{"2.2.2.2"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "legacyhostip.test.com.", RRSTTL: 300, RRSDatas: []string{"2.2.2.2"}, RRSType: rrstype.A},
		}
		c, err := controller.New(&controller.Options{
			DNSProvider:  dns,
			ZoneName:     "test.com.",
			Client:       client,
			TTL:          300,
			AddressTypes: []api.NodeAddressType{api.NodeLegacyHostIP},
		})
		Expect(err).To(BeNil())
		go runAndReportExit(c, report)
		time.Sleep(1 * time.Second)
		ls, err := rrs.List()
		Expect(err).To(BeNil())
		if !k8sutil.EqualRRSList(ls, expected) {
			pretty.Fprintf(GinkgoWriter, "# Expected Value:\n%# v\n\n", expected)
			pretty.Fprintf(GinkgoWriter, "# Received Value:\n%# v\n", ls)
			Fail("Unexpected DNS Records")
		}
		c.Stop()
		waitForReport(report)
	})

	It("should sync accept different TTL values", func() {
		expected := []dnsprovider.ResourceRecordSet{
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node4.externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"4.4.4.4"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node1.externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"1.1.1.1"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"1.1.1.1", "4.4.4.4"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node4.internalip.test.com.", RRSTTL: 300, RRSDatas: []string{"127.0.0.4"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node1.internalip.test.com.", RRSTTL: 300, RRSDatas: []string{"127.0.0.1"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "internalip.test.com.", RRSTTL: 300, RRSDatas: []string{"127.0.0.1", "127.0.0.4"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node2.legacyhostip.test.com.", RRSTTL: 300, RRSDatas: []string{"2.2.2.2"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "legacyhostip.test.com.", RRSTTL: 300, RRSDatas: []string{"2.2.2.2"}, RRSType: rrstype.A},
		}
		c, err := controller.New(&controller.Options{
			DNSProvider:  dns,
			ZoneName:     "test.com.",
			Client:       client,
			TTL:          300,
			AddressTypes: []api.NodeAddressType{api.NodeExternalIP, api.NodeInternalIP, api.NodeLegacyHostIP},
		})
		Expect(err).To(BeNil())
		go runAndReportExit(c, report)
		time.Sleep(1 * time.Second)
		ls, err := rrs.List()
		Expect(err).To(BeNil())
		if !k8sutil.EqualRRSList(ls, expected) {
			pretty.Fprintf(GinkgoWriter, "# Expected Value:\n%# v\n\n", expected)
			pretty.Fprintf(GinkgoWriter, "# Received Value:\n%# v\n", ls)
			Fail("Unexpected DNS Records")
		}
		c.Stop()
		waitForReport(report)
	})

	It("should sync different kind of addresses", func() {
		expected := []dnsprovider.ResourceRecordSet{
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node2.legacyhostip.test.com.", RRSTTL: 200, RRSDatas: []string{"2.2.2.2"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "legacyhostip.test.com.", RRSTTL: 200, RRSDatas: []string{"2.2.2.2"}, RRSType: rrstype.A},
		}
		c, err := controller.New(&controller.Options{
			DNSProvider:  dns,
			ZoneName:     "test.com.",
			Client:       client,
			TTL:          200,
			AddressTypes: []api.NodeAddressType{api.NodeLegacyHostIP},
		})
		Expect(err).To(BeNil())
		go runAndReportExit(c, report)
		time.Sleep(1 * time.Second)
		ls, err := rrs.List()
		Expect(err).To(BeNil())
		if !k8sutil.EqualRRSList(ls, expected) {
			pretty.Fprintf(GinkgoWriter, "# Expected Value:\n%# v\n\n", expected)
			pretty.Fprintf(GinkgoWriter, "# Received Value:\n%# v\n", ls)
			Fail("Unexpected DNS Records")
		}
		c.Stop()
		waitForReport(report)
	})

	It("should resync when DNS record changed out of band", func() {
		expected := []dnsprovider.ResourceRecordSet{
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node4.externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"4.4.4.4"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node1.externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"1.1.1.1"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"1.1.1.1", "4.4.4.4"}, RRSType: rrstype.A},
		}
		c, err := controller.New(&controller.Options{
			DNSProvider:  dns,
			ZoneName:     "test.com.",
			Client:       client,
			TTL:          300,
			AddressTypes: []api.NodeAddressType{api.NodeExternalIP},
			SyncInterval: 500 * time.Millisecond,
		})
		Expect(err).To(BeNil())
		go runAndReportExit(c, report)
		time.Sleep(1 * time.Second)
		rrs.Remove(&dnsproviderfake.ResourceRecordSetFake{RRSName: "node1.externalip.test.com."})
		rrs.Add(&dnsproviderfake.ResourceRecordSetFake{RRSName: "garbage.test.com.", RRSType: rrstype.A})
		time.Sleep(1 * time.Second)
		ls, err := rrs.List()
		Expect(err).To(BeNil())
		if !k8sutil.EqualRRSList(ls, expected) {
			pretty.Fprintf(GinkgoWriter, "# Expected Value:\n%# v\n\n", expected)
			pretty.Fprintf(GinkgoWriter, "# Received Value:\n%# v\n", ls)
			Fail("Unexpected DNS Records")
		}
		c.Stop()
		waitForReport(report)
	})

	It("should remove Node when it is becomes not ready", func() {
		expected := []dnsprovider.ResourceRecordSet{
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node4.externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"4.4.4.4"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"4.4.4.4"}, RRSType: rrstype.A},
		}
		c, err := controller.New(&controller.Options{
			DNSProvider:  dns,
			ZoneName:     "test.com.",
			Client:       client,
			TTL:          300,
			AddressTypes: []api.NodeAddressType{api.NodeExternalIP},
		})
		Expect(err).To(BeNil())
		go runAndReportExit(c, report)
		time.Sleep(1 * time.Second)
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
		ls, err := rrs.List()
		Expect(err).To(BeNil())
		if !k8sutil.EqualRRSList(ls, expected) {
			pretty.Fprintf(GinkgoWriter, "# Expected Value:\n%# v\n\n", expected)
			pretty.Fprintf(GinkgoWriter, "# Received Value:\n%# v\n", ls)
			Fail("Unexpected DNS Records")
		}
		c.Stop()
		waitForReport(report)
	})

	It("should remove Node when it gets deleted", func() {
		expected := []dnsprovider.ResourceRecordSet{
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node4.externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"4.4.4.4"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"4.4.4.4"}, RRSType: rrstype.A},
		}
		c, err := controller.New(&controller.Options{
			DNSProvider:  dns,
			ZoneName:     "test.com.",
			Client:       client,
			TTL:          300,
			AddressTypes: []api.NodeAddressType{api.NodeExternalIP},
		})
		Expect(err).To(BeNil())
		go runAndReportExit(c, report)
		time.Sleep(1 * time.Second)
		client.DeleteNode("node1")
		time.Sleep(500 * time.Millisecond)
		ls, err := rrs.List()
		Expect(err).To(BeNil())
		if !k8sutil.EqualRRSList(ls, expected) {
			pretty.Fprintf(GinkgoWriter, "# Expected Value:\n%# v\n\n", expected)
			pretty.Fprintf(GinkgoWriter, "# Received Value:\n%# v\n", ls)
			Fail("Unexpected DNS Records")
		}
		c.Stop()
		waitForReport(report)
	})

	It("should add Node when one gets created", func() {
		expected := []dnsprovider.ResourceRecordSet{
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node4.externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"4.4.4.4"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node5.externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"5.5.5.5"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node1.externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"1.1.1.1"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"1.1.1.1", "4.4.4.4", "5.5.5.5"}, RRSType: rrstype.A},
		}
		c, err := controller.New(&controller.Options{
			DNSProvider:  dns,
			ZoneName:     "test.com.",
			Client:       client,
			TTL:          300,
			AddressTypes: []api.NodeAddressType{api.NodeExternalIP},
		})
		Expect(err).To(BeNil())
		go runAndReportExit(c, report)
		time.Sleep(1 * time.Second)
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
		ls, err := rrs.List()
		Expect(err).To(BeNil())
		if !k8sutil.EqualRRSList(ls, expected) {
			pretty.Fprintf(GinkgoWriter, "# Expected Value:\n%# v\n\n", expected)
			pretty.Fprintf(GinkgoWriter, "# Received Value:\n%# v\n", ls)
			Fail("Unexpected DNS Records")
		}
		c.Stop()
		waitForReport(report)
	})

	It("should update IP when it changes", func() {
		expected := []dnsprovider.ResourceRecordSet{
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node4.externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"4.4.4.4"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node1.externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"6.6.6.6"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"6.6.6.6", "4.4.4.4"}, RRSType: rrstype.A},
		}
		c, err := controller.New(&controller.Options{
			DNSProvider:  dns,
			ZoneName:     "test.com.",
			Client:       client,
			TTL:          300,
			AddressTypes: []api.NodeAddressType{api.NodeExternalIP},
			SyncInterval: 500 * time.Millisecond,
		})
		Expect(err).To(BeNil())
		go runAndReportExit(c, report)
		time.Sleep(1 * time.Second)
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
		ls, err := rrs.List()
		Expect(err).To(BeNil())
		if !k8sutil.EqualRRSList(ls, expected) {
			pretty.Fprintf(GinkgoWriter, "# Expected Value:\n%# v\n\n", expected)
			pretty.Fprintf(GinkgoWriter, "# Received Value:\n%# v\n", ls)
			Fail("Unexpected DNS Records")
		}
		c.Stop()
		waitForReport(report)
	})

	It("should not touch Resources of other Types", func() {
		expected := []dnsprovider.ResourceRecordSet{
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node4.externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"4.4.4.4"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node1.externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"1.1.1.1"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"1.1.1.1", "4.4.4.4"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "keepit.test.com.", RRSType: rrstype.CNAME},
		}
		c, err := controller.New(&controller.Options{
			DNSProvider:  dns,
			ZoneName:     "test.com.",
			Client:       client,
			TTL:          300,
			AddressTypes: []api.NodeAddressType{api.NodeExternalIP},
			SyncInterval: 500 * time.Millisecond,
		})
		Expect(err).To(BeNil())
		go runAndReportExit(c, report)
		time.Sleep(1 * time.Second)
		rrs.Add(&dnsproviderfake.ResourceRecordSetFake{RRSName: "keepit.test.com.", RRSType: rrstype.CNAME})
		time.Sleep(500 * time.Millisecond)
		ls, err := rrs.List()
		Expect(err).To(BeNil())
		if !k8sutil.EqualRRSList(ls, expected) {
			pretty.Fprintf(GinkgoWriter, "# Expected Value:\n%# v\n\n", expected)
			pretty.Fprintf(GinkgoWriter, "# Received Value:\n%# v\n", ls)
			Fail("Unexpected DNS Records")
		}
		c.Stop()
		waitForReport(report)
	})

	It("should deal with sequence of changes", func() {
		expected := []dnsprovider.ResourceRecordSet{
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node4.externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"4.4.4.4"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "node6.externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"6.6.6.6"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "externalip.test.com.", RRSTTL: 300, RRSDatas: []string{"6.6.6.6", "4.4.4.4"}, RRSType: rrstype.A},
			&dnsproviderfake.ResourceRecordSetFake{RRSName: "keepit.test.com.", RRSType: rrstype.CNAME},
		}
		c, err := controller.New(&controller.Options{
			DNSProvider:  dns,
			ZoneName:     "test.com.",
			Client:       client,
			TTL:          300,
			AddressTypes: []api.NodeAddressType{api.NodeExternalIP},
			SyncInterval: 500 * time.Millisecond,
		})
		Expect(err).To(BeNil())
		go runAndReportExit(c, report)
		time.Sleep(1 * time.Second)
		client.DeleteNode("node1")
		rrs.Add(&dnsproviderfake.ResourceRecordSetFake{RRSName: "keepit.test.com.", RRSType: rrstype.CNAME})
		rrs.Remove(&dnsproviderfake.ResourceRecordSetFake{RRSName: "node4.externalip.test.com."})
		time.Sleep(500 * time.Millisecond)
		rrs.Add(&dnsproviderfake.ResourceRecordSetFake{RRSName: "garbage.test.com.", RRSType: rrstype.A})
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
		ls, err := rrs.List()
		Expect(err).To(BeNil())
		if !k8sutil.EqualRRSList(ls, expected) {
			pretty.Fprintf(GinkgoWriter, "# Expected Value:\n%# v\n\n", expected)
			pretty.Fprintf(GinkgoWriter, "# Received Value:\n%# v\n", ls)
			Fail("Unexpected DNS Records")
		}
		c.Stop()
		waitForReport(report)
	})

})
