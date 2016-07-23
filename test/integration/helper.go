/*
 * Copyright (C) 2016 wikiwi.io
 *
 * This software may be modified and distributed under the terms
 * of the MIT license. See the LICENSE file for details.
 */

package integration

import (
	"github.com/onsi/gomega"

	"github.com/wikiwi/kube-dns-sync/pkg/controller"
	"k8s.io/kubernetes/pkg/labels"
)

// runAndReportExit runs given Controller, expects err=nil, and notifies channel report.
func runAndReportExit(c *controller.Controller, report chan struct{}) {
	err := c.Run()
	gomega.Expect(err).To(gomega.BeNil())
	report <- struct{}{}
}

// waitForReport wait for report to be triggered, timeout when it takes longer than 1s.
func waitForReport(report chan struct{}) {
	gomega.Eventually(func() interface{} { return <-report })
}

// parseSelectorOrDie will parse s as Selector or panics on error.
func parseSelectorOrDie(s string) labels.Selector {
	sel, err := labels.Parse(s)
	if err != nil {
		panic(err)
	}
	return sel
}
