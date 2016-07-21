/*
 * Copyright (C) 2016 wikiwi.io
 *
 * This software may be modified and distributed under the terms
 * of the MIT license. See the LICENSE file for details.
 */

package kubernetes

import (
	"testing"

	"github.com/kr/pretty"

	"k8s.io/kubernetes/federation/pkg/dnsprovider/rrstype"
	"k8s.io/kubernetes/pkg/api"

	"github.com/wikiwi/kube-dns-sync/pkg/util/kubernetes/dnsproviderfake"
)

func TestEqualRRS(t *testing.T) {
	testScenarios := []struct {
		Aname    string
		Attl     int64
		ArrsType rrstype.RrsType
		Adata    []string
		Bname    string
		Bttl     int64
		BrrsType rrstype.RrsType
		Bdata    []string
		equal    bool
	}{
		{
			Aname: "name", Attl: 300, ArrsType: rrstype.A, Adata: []string{"data"},
			Bname: "name", Bttl: 300, BrrsType: rrstype.A, Bdata: []string{"data"},
			equal: true,
		},
		{
			Aname: "name", Attl: 300, ArrsType: rrstype.A, Adata: []string{"data"},
			Bname: "different", Bttl: 300, BrrsType: rrstype.A, Bdata: []string{"data"},
			equal: false,
		},
		{
			Aname: "name", Attl: 300, ArrsType: rrstype.A, Adata: []string{"data"},
			Bname: "name", Bttl: 400, BrrsType: rrstype.A, Bdata: []string{"data"},
			equal: false,
		},
		{
			Aname: "name", Attl: 300, ArrsType: rrstype.A, Adata: []string{"data"},
			Bname: "name", Bttl: 300, BrrsType: rrstype.CNAME, Bdata: []string{"data"},
			equal: false,
		},
		{
			Aname: "name", Attl: 300, ArrsType: rrstype.A, Adata: []string{"data"},
			Bname: "name", Bttl: 300, BrrsType: rrstype.A, Bdata: []string{"data", "more"},
			equal: false,
		},
		{
			Aname: "name", Attl: 300, ArrsType: rrstype.A, Adata: []string{"a", "b"},
			Bname: "name", Bttl: 300, BrrsType: rrstype.A, Bdata: []string{"b", "a"},
			equal: true,
		},
	}
	for _, x := range testScenarios {
		t.Log(pretty.Sprint(x))
		a := &dnsproviderfake.ResourceRecordSetFake{
			RRSName: x.Aname, RRSType: x.ArrsType, RRSTTL: x.Attl, RRSDatas: x.Adata,
		}
		b := &dnsproviderfake.ResourceRecordSetFake{
			RRSName: x.Bname, RRSType: x.BrrsType, RRSTTL: x.Bttl, RRSDatas: x.Bdata,
		}
		if EqualRRS(a, b) != x.equal {
			if x.equal {
				t.Errorf("expected equality but was %v", pretty.Diff(a, b))
			} else {
				t.Errorf("expected inequality")
			}
		}
	}
}

func TestIsNodeReady(t *testing.T) {
	testScenarios := []struct {
		conditions []api.NodeCondition
		ready      bool
	}{
		{
			conditions: []api.NodeCondition{api.NodeCondition{
				Type:   api.NodeReady,
				Status: api.ConditionTrue,
			}},
			ready: true,
		},
		{
			conditions: []api.NodeCondition{api.NodeCondition{
				Type:   api.NodeReady,
				Status: api.ConditionFalse,
			}},
			ready: false,
		},
		{
			conditions: []api.NodeCondition{},
			ready:      false,
		},
	}
	for _, x := range testScenarios {
		t.Log(pretty.Sprint(x))
		node := &api.Node{}
		node.Status.Conditions = x.conditions
		if ready := IsNodeReady(node); ready != x.ready {
			t.Errorf("expect readiness of %v, but was %v", x.ready, ready)
		}
	}
}
