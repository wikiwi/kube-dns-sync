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

	"github.com/jessevdk/go-flags"

	"github.com/cvle/kube-dns-sync/pkg/version"
)

var opts struct {
	DNSProvider       string         `long:"dns-provider" env:"DNS_SYNC_PROVIDER" description:"DNS provider" required:"yes"`
	DNSProviderConfig flags.Filename `long:"dns-provider-config" env:"DNS_SYNC_PROVIDER_CONFIG" description:"Path to config file for configuring DNS provider"`
	ZoneName          string         `long:"zone-name" env:"DNS_SYNC_ZONE_NAME" description:"Zone name, like example.com" required:"yes"`
	SyncInterval      time.Duration  `long:"sync-interval" default:"60s" env:"DNS_SYNC_INTERVAL" description:"Interval for syncing with the DNS Provider"`
	TTL               int64          `long:"ttl" default:"300" env:"DNS_SYNC_TTL" description:"TTL value of DNS Records"`
	AddressTypes      addressTypes   `long:"address-types" default:"externalip" env:"DNS_SYNC_ADDRESS_TYPES" description:"Comma list of address types to export [externalip|internalip|legacyhostip]"`
	Version           func()         `long:"version" short:"v" description:"show version number"`
}

func init() {
	opts.Version = func() {
		fmt.Println(version.Version)
		os.Exit(0)
	}
}
