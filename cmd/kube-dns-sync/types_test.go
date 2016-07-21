/*
 * Copyright (C) 2016 wikiwi.io
 *
 * This software may be modified and distributed under the terms
 * of the MIT license. See the LICENSE file for details.
 */

package main

import (
	"reflect"
	"testing"

	"github.com/kr/pretty"

	"k8s.io/kubernetes/pkg/api"
)

func TestAddressTypes(t *testing.T) {
	testScenarios := []struct {
		input  string
		expect []api.NodeAddressType
		err    bool
	}{
		{input: "externalip", expect: []api.NodeAddressType{api.NodeExternalIP}},
		{input: "externalip,internalip", expect: []api.NodeAddressType{api.NodeExternalIP, api.NodeInternalIP}},
		{input: "", expect: []api.NodeAddressType{}},
		{input: "invalid", expect: []api.NodeAddressType{}, err: true},
	}
	for _, x := range testScenarios {
		t.Log(pretty.Sprint(x))
		a := &addressTypes{}
		err := a.UnmarshalFlag(x.input)
		if x.err && err == nil {
			t.Errorf("error unmarshalling: %q", err)
		}
		if !reflect.DeepEqual(x.expect, a.Types) {
			t.Errorf("%v", pretty.Diff(x.expect, a.Types))
		}
		if err != nil {
			continue
		}
		marshalled, err := a.MarshalFlag()
		if err != nil {
			t.Errorf("error marshalling: %q", err)
			continue
		}
		if marshalled != x.input {
			t.Errorf("%q != %q", marshalled, x.input)
		}
	}
}
