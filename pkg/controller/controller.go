/*
 * Copyright (C) 2016 wikiwi.io
 *
 * This software may be modified and distributed under the terms
 * of the MIT license. See the LICENSE file for details.
 */

package controller

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"

	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/client/unversioned"

	k8sutil "github.com/cvle/kube-dns-sync/pkg/util/kubernetes"
)

// Options for creating a new Controller.
type Options struct {
	// DNSProvider is the provider for dns services, required.
	DNSProvider dnsprovider.Interface

	// ZoneName, like "example.com.", required.
	ZoneName string

	// TTL value of Records, defaults to 300
	TTL int64

	// SyncInterval is the interval for syncing with the DNS Provider, defaults to 60 seconds.
	SyncInterval time.Duration

	// Client is the Kubernetes Client to use or use default when nil.
	Client unversioned.Interface

	// AddressTypes are the address types to sync to DNS, defaults to NodeExternalIP.
	AddressTypes []api.NodeAddressType
}

// New creates a new Controller.
func New(opts *Options) (*Controller, error) {
	c := &Controller{}
	if opts.DNSProvider == nil {
		return nil, fmt.Errorf("please provide a DNS Provider")
	}
	if opts.ZoneName == "" {
		return nil, fmt.Errorf("please provide a zone name")
	}
	c.dns = opts.DNSProvider
	c.ttl = opts.TTL
	c.zoneName = opts.ZoneName
	c.client = opts.Client
	c.addressTypes = opts.AddressTypes
	c.syncInterval = opts.SyncInterval
	c.stopCh = make(chan struct{})
	c.syncCh = make(chan struct{})
	c.log = logrus.StandardLogger()
	if c.ttl == 0 {
		c.ttl = 300
	}
	if c.client == nil {
		client, err := k8sutil.NewKubeClient()
		if err != nil {
			return nil, err
		}
		c.client = client
	}
	if c.syncInterval == 0 {
		c.syncInterval = time.Second * 60
	}
	if c.addressTypes == nil {
		c.addressTypes = []api.NodeAddressType{api.NodeExternalIP}
	}
	return c, nil
}

// Controller syncs Kubernetes Node IPs to a DNS service.
type Controller struct {
	dns          dnsprovider.Interface
	dnsProvider  string
	zoneName     string
	ttl          int64
	syncInterval time.Duration
	log          *logrus.Logger
	stopCh       chan struct{}
	syncCh       chan struct{}
	client       unversioned.Interface
	addressTypes []api.NodeAddressType
	cache        cache.Store
}

// Run starts the Controller Controller in an endless loop.
func (c *Controller) Run() error {
	c.watch()
	c.loop()
	return nil
}

// Stop will unblock Run(). Only call this once.
func (c *Controller) Stop() {
	close(c.stopCh)
}

// loop blocks and run sync when it is request through
// syncCh or when syncInterval has passed.
func (c *Controller) loop() {
	timer := time.NewTimer(c.syncInterval)
	sync := func() {
		err := c.sync()
		if err != nil {
			c.log.Error(err)
		}
		timer.Reset(c.syncInterval)
	}
L:
	for {
		select {
		case <-c.stopCh:
			timer.Stop()
			break L
		case <-timer.C:
			sync()
		case <-c.syncCh:
			sync()
		}
	}
}

// requestSync will trigger a sync in the next loop iteration.
func (c *Controller) requestSync() {
	select {
	case c.syncCh <- struct{}{}:
	default:
	}
}
