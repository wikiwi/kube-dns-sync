/*
 * Copyright (C) 2016 wikiwi.io
 *
 * This software may be modified and distributed under the terms
 * of the MIT license. See the LICENSE file for details.
 */

package integration

import (
	"fmt"
	"sync"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/unversioned/testclient"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/watch"
)

func newKubeFake(nodes ...api.Node) *kubeFake {
	f := new(kubeFake)
	f.init(nodes)
	return f
}

// kubeFake implements a fake Kubernetes Client. It only deals with the Nodes Service and the verbs 'list', and 'watch'.
// This implementation is Thread-Safe.
type kubeFake struct {
	*testclient.Fake
	nodeList  api.NodeList
	lock      sync.Mutex
	fakeWatch *watch.FakeWatcher
}

func (f *kubeFake) init(nodes []api.Node) {
	f.fakeWatch = watch.NewFake()
	fakeClient := &testclient.Fake{}
	fakeClient.AddReactor("list", "nodes", f.reactor)
	fakeClient.AddWatchReactor("nodes", testclient.DefaultWatchReactor(f.fakeWatch, nil))
	f.Fake = fakeClient
	for _, x := range nodes {
		f.AddNode(x)
	}
}

func (f *kubeFake) reactor(action testclient.Action) (handled bool, ret runtime.Object, err error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	return true, &f.nodeList, nil
}

func (f *kubeFake) AddNode(node api.Node) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.nodeList.Items = append(f.nodeList.Items, node)
	go func(node *api.Node) {
		f.fakeWatch.Add(node)
	}(&node)
}

func (f *kubeFake) DeleteNode(name string) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	for i, x := range f.nodeList.Items {
		if x.Name == name {
			f.nodeList.Items = append(f.nodeList.Items[:i], f.nodeList.Items[i+1:]...)
			go func(node *api.Node) {
				f.fakeWatch.Delete(node)
			}(&x)
			return nil
		}
	}
	return fmt.Errorf("Node %q not found", name)
}

func (f *kubeFake) ModifyNode(node api.Node) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	for i, x := range f.nodeList.Items {
		if x.Name == node.ObjectMeta.Name {
			f.nodeList.Items = append(append(f.nodeList.Items[:i], node), f.nodeList.Items[i+1:]...)
			go func(node *api.Node) {
				f.fakeWatch.Modify(node)
			}(&node)
			return nil
		}
	}
	return fmt.Errorf("Node %q not found", node.ObjectMeta.Name)
}
