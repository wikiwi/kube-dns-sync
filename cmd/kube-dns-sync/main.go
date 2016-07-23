/*
 * Copyright (C) 2016 wikiwi.io
 *
 * This software may be modified and distributed under the terms
 * of the MIT license. See the LICENSE file for details.
 */

// kube-dns-sync implements an executable running the Controller.
package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/jessevdk/go-flags"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"

	"github.com/wikiwi/kube-dns-sync/pkg/controller"
	"k8s.io/kubernetes/pkg/api"
)

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	parser.FindOptionByLongName("dns-provider").Choices = dnsprovider.RegisteredDnsProviders()
	parser.Name = "kube-dns-sync"
	_, err := parser.Parse()
	if err != nil {
		if e2, ok := err.(*flags.Error); ok && e2.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	if opts.ApexAddressType == "" && len(opts.AddressTypes) == 0 {
		fmt.Println("neither --address-types nor --apex-address-type is specified")
		os.Exit(1)
	}

	dump, err := yaml.Marshal(opts)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Starting with following configuration\n%s", string(dump))
	dnsProvider, err := dnsprovider.InitDnsProvider(opts.DNSProvider, string(opts.DNSProviderConfig))
	if err != nil {
		panic(err)
	}
	c, err := controller.New(&controller.Options{
		DNSProvider:     dnsProvider,
		TTL:             opts.TTL,
		ZoneName:        opts.ZoneName,
		SyncInterval:    opts.SyncInterval,
		AddressTypes:    opts.AddressTypes,
		ApexAddressType: api.NodeAddressType(opts.ApexAddressType),
	})
	if err != nil {
		panic(err)
	}
	panic(c.Run())
}
