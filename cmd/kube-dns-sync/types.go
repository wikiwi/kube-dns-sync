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
)

type addressTypes struct {
	Types []api.NodeAddressType
}

func (a addressTypes) MarshalFlag() (string, error) {
	var s string
	for _, x := range a.Types {
		if s != "" {
			s += ","
		}
		s += strings.ToLower(string(x))
	}
	return s, nil
}

func (a *addressTypes) UnmarshalFlag(value string) error {
	a.Types = []api.NodeAddressType{}
	parts := strings.Split(value, ",")
	for _, x := range parts {
		found := false
		for _, t := range []api.NodeAddressType{api.NodeInternalIP, api.NodeExternalIP, api.NodeLegacyHostIP} {
			if x == strings.ToLower(string(t)) {
				a.Types = append(a.Types, t)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Invalid value %q", x)
		}
	}
	return nil
}
