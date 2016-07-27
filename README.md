# kube-dns-sync
kube-dns-sync is a Kubernetes Controller that syncs Kubernetes Node IPs to a DNS service.

[![Build Status Widget]][Build Status]
[![Coverage Status Widget]][Coverage Status]
[![Code Climate Widget]][Code Climate]
[![MicroBadger Version Widget]][MicroBadger Version]

[Build Status]: https://travis-ci.org/wikiwi/kube-dns-sync
[Build Status Widget]: https://travis-ci.org/wikiwi/kube-dns-sync.svg?branch=master
[Coverage Status]: https://coveralls.io/github/wikiwi/kube-dns-sync?branch=master
[Coverage Status Widget]: https://coveralls.io/repos/github/wikiwi/kube-dns-sync/badge.svg?branch=master
[Code Climate]: https://codeclimate.com/github/wikiwi/kube-dns-sync
[Code Climate Widget]: https://codeclimate.com/github/wikiwi/kube-dns-sync/badges/gpa.svg
[MicroBadger Version]: http://microbadger.com/#/images/wikiwi/kube-dns-sync
[MicroBadger Version Widget]: https://images.microbadger.com/badges/version/wikiwi/kube-dns-sync.svg

## Supported DNS service
kube-dns-sync uses the DNS module of Kubernetes Federation and therefore supports the same DNS services. At the time of writing the supported services are 'google-clouddns' and 'aws-route53'.

## Authorization
The authorization mechanics are the same as for Kubernetes Federation. A link will be put here as soon as Kubernetes releases an official documentation for its Federation Service. 

*note:* google-clouddns requires the scope `https://www.googleapis.com/auth/ndev.clouddns.readwrite`.

## Flags and Environment Variables
    Usage:
      kube-dns-sync [OPTIONS]
    
    Application Options:
          --dns-provider=[aws-route53|google-clouddns]             DNS provider [$KDS_PROVIDER]
          --dns-provider-config=                                   Path to config file for configuring DNS provider [$KDS_PROVIDER_CONFIG]
          --zone-name=                                             Zone name, like example.com [$KDS_ZONE_NAME]
          --sync-interval=                                         Interval for syncing with the DNS Provider (default: 60s) [$KDS_INTERVAL]
          --ttl=                                                   TTL value of DNS Records (default: 60) [$KDS_TTL]
          --address-types=                                         Comma list of address types to sync [externalip|internalip|legacyhostip] [$KDS_ADDRESS_TYPES]
          --apex-address-type=[externalip|internalip|legacyhostip] Address type that is synced to the Apex Zone [$KDS_APEX_ADDRESS_TYPE]
          --selector=                                              Node selector e.g. 'cloud.google.com/gke-nodepool=default-pool' [$KDS_SELECTOR]
          --verbose                                                Turn on verbose logging
      -v, --version                                                Show version number
    
    Help Options:
      -h, --help                                                   Show this help message
