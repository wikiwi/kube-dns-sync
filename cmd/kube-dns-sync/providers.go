/*
 * Copyright (C) 2016 wikiwi.io
 *
 * This software may be modified and distributed under the terms
 * of the MIT license. See the LICENSE file for details.
 */

package main

import (
	// Blank import to force loading of the Kubernetes DNS Provider Plugins.
	_ "k8s.io/kubernetes/federation/pkg/dnsprovider/providers/aws/route53"
	_ "k8s.io/kubernetes/federation/pkg/dnsprovider/providers/google/clouddns"
)
