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
		{input: "", expect: nil},
		{input: "invalid", expect: nil, err: true},
	}
	for _, x := range testScenarios {
		t.Log(pretty.Sprint(x))
		var a addressTypes
		err := a.UnmarshalFlag(x.input)
		if x.err && err == nil {
			t.Errorf("error unmarshalling: %q", err)
		}

		types := []api.NodeAddressType(a)
		if !reflect.DeepEqual(x.expect, types) {
			t.Errorf("%v", pretty.Diff(x.expect, types))
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

func TestAddressType(t *testing.T) {
	testScenarios := []struct {
		input  string
		expect api.NodeAddressType
		err    bool
	}{
		{input: "externalip", expect: api.NodeExternalIP},
		{input: "internalip", expect: api.NodeInternalIP},
		{input: "legacyhostip", expect: api.NodeLegacyHostIP},
		{input: "", expect: ""},
		{input: "invalid", expect: "", err: true},
	}
	for _, x := range testScenarios {
		t.Log(pretty.Sprint(x))
		var a addressType
		err := a.UnmarshalFlag(x.input)
		if x.err && err == nil {
			t.Errorf("error unmarshalling: %q", err)
		}

		converted := api.NodeAddressType(a)
		if converted != x.expect {
			t.Errorf("%v != %v", x.expect, converted)
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

func TestSelectorType(t *testing.T) {
	testScenarios := []struct {
		input string
		err   bool
	}{
		{input: "environment in (production,qa)"},
		{input: "foo=bar"},
		{input: ""},
	}
	for _, x := range testScenarios {
		t.Log(pretty.Sprint(x))
		var a selectorType
		err := a.UnmarshalFlag(x.input)
		if x.err && err == nil {
			t.Errorf("error unmarshalling: %q", err)
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
