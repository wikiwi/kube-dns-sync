/*
 * Copyright (C) 2016 wikiwi.io
 *
 * This software may be modified and distributed under the terms
 * of the MIT license. See the LICENSE file for details.
 */

package controller

import (
	"reflect"
	"time"

	"github.com/kr/pretty"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/controller/framework"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/watch"

	k8sutil "github.com/cvle/kube-dns-sync/pkg/util/kubernetes"
)

// watch watches the Kubernetes API and requests a sync when a change was detected.
func (c *Controller) watch() {
	c.log.Infof("Start kubernetes watcher")

	resyncPeriod := time.Second * 60
	nodeEventHandler := framework.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node := obj.(*api.Node)
			c.log.Infof("CREATE %s/%s", node.Namespace, node.Name)
			c.requestSync()
		},
		DeleteFunc: func(obj interface{}) {
			node := obj.(*api.Node)
			c.log.Infof("DELETE %s/%s", node.Namespace, node.Name)
			c.requestSync()
		},
		UpdateFunc: func(oldI, curI interface{}) {
			cur := curI.(*api.Node)
			old := oldI.(*api.Node)
			if k8sutil.IsNodeReady(old) != k8sutil.IsNodeReady(cur) || !reflect.DeepEqual(old.Status.Addresses, cur.Status.Addresses) {
				c.log.Infof("UPDATE %s/%s", cur.Namespace, cur.Name)
				pretty.Pdiff(c.log, old.Status.Addresses, cur.Status.Addresses)
				c.requestSync()
			}
		},
	}

	store, controller := framework.NewInformer(
		&cache.ListWatch{
			ListFunc:  func(opts api.ListOptions) (runtime.Object, error) { return c.client.Nodes().List(opts) },
			WatchFunc: func(opts api.ListOptions) (watch.Interface, error) { return c.client.Nodes().Watch(opts) },
		},
		&api.Node{},
		resyncPeriod,
		nodeEventHandler,
	)

	c.cache = store

	go controller.Run(c.stopCh)
}
