/*
 * Copyright (C) 2016 wikiwi.io
 *
 * This software may be modified and distributed under the terms
 * of the MIT license. See the LICENSE file for details.
 */

package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/cvle/kube-dns-sync/pkg/version"
)

func TestVersion(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		os.Args = []string{"kube-dns-sync", "--version"}
		main()
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestVersion")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	raw, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("process ran with err %v", err)
	}
	firstline := strings.TrimSpace(strings.Split(string(raw), "\n")[0])
	if firstline != version.Version {
		t.Fatalf("%q != %q", firstline, version.Version)
	}
}
