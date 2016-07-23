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
	"k8s.io/kubernetes/pkg/labels"
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
	nodeList          api.NodeList
	initialNodes      []api.Node
	lock              sync.Mutex
	fakeWatch         *watch.FakeWatcher
	watchRestrictions testclient.WatchRestrictions
}

func (f *kubeFake) init(nodes []api.Node) {
	f.fakeWatch = watch.NewFake()
	fakeClient := &testclient.Fake{}
	fakeClient.AddReactor("list", "nodes", f.reactor)
	fakeClient.AddWatchReactor("nodes", f.reactorWatch)
	f.Fake = fakeClient
	f.initialNodes = nodes
}

func (f *kubeFake) reactor(action testclient.Action) (handled bool, ret runtime.Object, err error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	listAction := action.(testclient.ListAction)
	var nodeList api.NodeList
	for _, x := range f.nodeList.Items {
		fmt.Println(listAction.GetListRestrictions().Labels.String())
		if listAction.GetListRestrictions().Labels.Matches(labels.Set(x.Labels)) {
			nodeList.Items = append(nodeList.Items, x)
		}
	}
	return true, &nodeList, nil
}

func (f *kubeFake) reactorWatch(action testclient.Action) (bool, watch.Interface, error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	watchAction := action.(testclient.WatchAction)
	f.watchRestrictions = watchAction.GetWatchRestrictions()
	go func() {
		for _, x := range f.initialNodes {
			f.AddNode(x)
		}
	}()
	return true, f.fakeWatch, nil
}

func (f *kubeFake) shouldNotify(node api.Node) bool {
	if f.watchRestrictions.Labels.Matches(labels.Set(node.Labels)) {
		return true
	}
	return false
}

func (f *kubeFake) AddNode(node api.Node) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.nodeList.Items = append(f.nodeList.Items, node)
	if f.shouldNotify(node) {
		go func(node *api.Node) {
			f.fakeWatch.Add(node)
		}(&node)
	}
}

func (f *kubeFake) DeleteNode(name string) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	for i, node := range f.nodeList.Items {
		if node.Name == name {
			f.nodeList.Items = append(f.nodeList.Items[:i], f.nodeList.Items[i+1:]...)
			if f.shouldNotify(node) {
				go func(node *api.Node) {
					f.fakeWatch.Delete(node)
				}(&node)
			}
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
			if f.shouldNotify(node) {
				go func(node *api.Node) {
					f.fakeWatch.Modify(node)
				}(&node)
			}
			return nil
		}
	}
	return fmt.Errorf("Node %q not found", node.ObjectMeta.Name)
}
