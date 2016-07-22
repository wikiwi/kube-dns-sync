/*
 * Copyright (C) 2016 wikiwi.io
 *
 * This software may be modified and distributed under the terms
 * of the MIT license. See the LICENSE file for details.
 */

package kubernetes

// Package kubernetes contains tools for interfacing with Kubernetes.
import (
	"reflect"
	"sort"
	"strings"

	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
)

// NewKubeClient creates a new Unversioned Kubernetes Client using default loading rules.
func NewKubeClient() (*unversioned.Client, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	client, err := unversioned.New(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// EqualRRSList return true when both arguments are equal.
func EqualRRSList(a []dnsprovider.ResourceRecordSet, b []dnsprovider.ResourceRecordSet) bool {
	if len(a) != len(b) {
		return false
	}
	for _, x := range a {
		found := false
		for i, y := range b {
			if EqualRRS(x, y) {
				b = append(b[:i], b[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// EqualRRS return true when both arguments are equal.
func EqualRRS(a dnsprovider.ResourceRecordSet, b dnsprovider.ResourceRecordSet) bool {
	if a.Name() != b.Name() {
		return false
	}
	if a.Ttl() != b.Ttl() {
		return false
	}
	if a.Type() != b.Type() {
		return false
	}
	dataA := sort.StringSlice(a.Rrdatas())
	dataA.Sort()
	dataB := sort.StringSlice(b.Rrdatas())
	dataB.Sort()
	return reflect.DeepEqual(dataA, dataB)
}

// IsNodeReady checks for the NodeReady condition in the Node.
func IsNodeReady(node *api.Node) bool {
	for _, cond := range node.Status.Conditions {
		if cond.Type == api.NodeReady {
			return cond.Status == api.ConditionTrue
		}
	}
	return false
}

// StringToAddressType converts a string to a NodeAddressType, ignores case.
func StringToAddressType(s string) api.NodeAddressType {
	norm := strings.ToLower(s)
	switch norm {
	case "externalip":
		return api.NodeExternalIP
	case "internalip":
		return api.NodeInternalIP
	case "legacyhostip":
		return api.NodeLegacyHostIP
	}
	return ""
}
