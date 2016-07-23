/*
 * Copyright (C) 2016 wikiwi.io
 *
 * This software may be modified and distributed under the terms
 * of the MIT license. See the LICENSE file for details.
 */

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"

	"github.com/wikiwi/kube-dns-sync/pkg/version"
)

var opts struct {
	DNSProvider       string         `long:"dns-provider" env:"KDS_PROVIDER" description:"DNS provider" required:"yes"`
	DNSProviderConfig flags.Filename `long:"dns-provider-config" env:"KDS_PROVIDER_CONFIG" description:"Path to config file for configuring DNS provider"`
	ZoneName          string         `long:"zone-name" env:"KDS_ZONE_NAME" description:"Zone name, like example.com" required:"yes"`
	SyncInterval      time.Duration  `long:"sync-interval" default:"60s" env:"KDS_INTERVAL" description:"Interval for syncing with the DNS Provider"`
	TTL               int64          `long:"ttl" default:"60" env:"KDS_TTL" description:"TTL value of DNS Records"`
	AddressTypes      addressTypes   `long:"address-types" env:"KDS_ADDRESS_TYPES" description:"Comma list of address types to sync [externalip|internalip|legacyhostip]"`
	ApexAddressType   addressType    `long:"apex-address-type" env:"KDS_APEX_ADDRESS_TYPE" description:"Address type that is synced to the Apex Zone" choice:"externalip" choice:"internalip" choice:"legacyhostip"`
	SelectorType      selectorType   `long:"selector" env:"KDS_SELECTOR" description:"Node selector e.g. 'cloud.google.com/gke-nodepool=default-pool'"`
	Verbose           func()         `yaml:"-" long:"verbose"  description:"Turn on verbose logging"`
	Version           func()         `yaml:"-" long:"version" short:"v" description:"Show version number"`
}

func init() {
	opts.Version = func() {
		fmt.Println(version.Version)
		os.Exit(0)
	}
	opts.Verbose = func() {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
