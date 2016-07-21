/*
 * Copyright (C) 2016 wikiwi.io
 *
 * This software may be modified and distributed under the terms
 * of the MIT license. See the LICENSE file for details.
 */

package integration

import (
	"k8s.io/kubernetes/federation/pkg/dnsprovider/rrstype"
	"k8s.io/kubernetes/pkg/api"

	"github.com/wikiwi/kube-dns-sync/pkg/util/kubernetes/dnsproviderfake"
)

// k8sFixture is added as the initial Kubernetes resources for the integration tests.
var k8sFixture = []api.Node{
	{
		ObjectMeta: api.ObjectMeta{Name: "node1"},
		Status: api.NodeStatus{
			Addresses: []api.NodeAddress{
				api.NodeAddress{Type: api.NodeExternalIP, Address: "1.1.1.1"},
				api.NodeAddress{Type: api.NodeInternalIP, Address: "127.0.0.1"},
			},
			Conditions: []api.NodeCondition{api.NodeCondition{
				Type:   api.NodeReady,
				Status: api.ConditionTrue,
			}},
		},
	},
	{
		ObjectMeta: api.ObjectMeta{Name: "node2"},
		Status: api.NodeStatus{
			Addresses: []api.NodeAddress{
				api.NodeAddress{Type: api.NodeLegacyHostIP, Address: "2.2.2.2"},
			},
			Conditions: []api.NodeCondition{api.NodeCondition{
				Type:   api.NodeReady,
				Status: api.ConditionTrue,
			}},
		},
	},
	{
		ObjectMeta: api.ObjectMeta{Name: "node3"},
		Status: api.NodeStatus{
			Addresses: []api.NodeAddress{
				api.NodeAddress{Type: api.NodeExternalIP, Address: "3.3.3.3"},
				api.NodeAddress{Type: api.NodeInternalIP, Address: "127.0.0.3"},
			},
			Conditions: []api.NodeCondition{api.NodeCondition{
				Type:   api.NodeReady,
				Status: api.ConditionFalse,
			}},
		},
	},
	{
		ObjectMeta: api.ObjectMeta{Name: "node4"},
		Status: api.NodeStatus{
			Addresses: []api.NodeAddress{
				api.NodeAddress{Type: api.NodeExternalIP, Address: "4.4.4.4"},
				api.NodeAddress{Type: api.NodeInternalIP, Address: "127.0.0.4"},
			},
			Conditions: []api.NodeCondition{api.NodeCondition{
				Type:   api.NodeReady,
				Status: api.ConditionTrue,
			}},
		},
	},
}

// dnsFixture is added as the initial DNS resources for the integration tests.
var dnsFixture = []dnsproviderfake.ResourceRecordSetFake{
	dnsproviderfake.ResourceRecordSetFake{
		RRSName: "garbage",
		RRSType: rrstype.A,
	},
}
