/*
 * Copyright (C) 2016 wikiwi.io
 *
 * This software may be modified and distributed under the terms
 * of the MIT license. See the LICENSE file for details.
 */

package main

import (
	"fmt"
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/labels"
)

type addressTypes []api.NodeAddressType

func (a addressTypes) MarshalFlag() (string, error) {
	var s string
	for _, x := range a {
		if s != "" {
			s += ","
		}
		tmp, err := (addressType(x)).MarshalFlag()
		if err != nil {
			return "", err
		}
		s += tmp
	}
	return s, nil
}

func (a *addressTypes) UnmarshalFlag(value string) error {
	parts := strings.Split(value, ",")
	for _, x := range parts {
		var add addressType
		err := add.UnmarshalFlag(x)
		if err != nil {
			return err
		}
		*a = append(*a, api.NodeAddressType(add))
	}
	return nil
}

type addressType api.NodeAddressType

func (a addressType) MarshalFlag() (string, error) {
	return strings.ToLower(string(a)), nil
}

func (a *addressType) UnmarshalFlag(value string) error {
	for _, t := range []api.NodeAddressType{api.NodeInternalIP, api.NodeExternalIP, api.NodeLegacyHostIP} {
		stringified := strings.ToLower(string(t))
		if value == stringified {
			*a = addressType(t)
			return nil
		}
	}
	return fmt.Errorf("Invalid value %q", value)
}

type selectorType struct {
	labels.Selector
}

func (s selectorType) MarshalFlag() (string, error) {
	if s.Selector == nil {
		return "", nil
	}
	return s.String(), nil
}

func (s *selectorType) UnmarshalFlag(value string) error {
	sel, err := labels.Parse(value)
	if err != nil {
		return err
	}
	s.Selector = sel
	return nil
}
