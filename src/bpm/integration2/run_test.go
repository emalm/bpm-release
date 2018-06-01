// Copyright (C) 2018-Present CloudFoundry.org Foundation, Inc. All rights reserved.
//
// This program and the accompanying materials are made available under
// the terms of the under the Apache License, Version 2.0 (the "License”);
// you may not use this file except in compliance with the License.
//
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  See the
// License for the specific language governing permissions and limitations
// under the License.

package integration2_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/onsi/gomega/gexec"
)

func TestRun(t *testing.T) {
	s, err := NewSandbox()
	if err != nil {
		t.Fatalf("sandbox setup failed: %v", err)
	}
	defer s.Cleanup()

	if err := s.Fixture("blah", "testdata/blah.yml"); err != nil {
		t.Fatalf("couldn't load fixture: %v", err)
	}

	cmd := s.BPMCmd("run", "blah")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run bpm: %s", output)
	}
}

type Sandbox struct {
	bpmExe  string
	runcExe string

	root string
}

const runcExe = "/var/vcap/packages/bpm/bin/runc"

var bpmExe string

func init() {
	var err error
	bpmExe, err = gexec.Build("bpm/cmd/bpm")
	if err != nil {
		panic(err)
	}
}

func NewSandbox() (*Sandbox, error) {
	root, err := ioutil.TempDir("", "bpm_sandbox")
	if err != nil {
		return nil, fmt.Errorf("could not create sandbox root directory: %v", err)
	}

	paths := []string{
		filepath.Join(root, "packages", "bpm", "bin"),
		filepath.Join(root, "data", "packages"),
	}

	for _, path := range paths {
		if err := os.MkdirAll(path, 0777); err != nil {
			return nil, fmt.Errorf("could not create sandbox directory structure: %v", err)
		}
	}

	runcSandboxPath := filepath.Join(root, "packages", "bpm", "bin", "runc")
	if err := os.Symlink(runcExe, runcSandboxPath); err != nil {
		return nil, fmt.Errorf("could not link runc executable into sandbox: %v", err)
	}

	return &Sandbox{
		bpmExe:  bpmExe,
		runcExe: runcExe,
		root:    root,
	}, nil
}

func (s *Sandbox) BPMCmd(args ...string) *exec.Cmd {
	cmd := exec.Command(s.bpmExe, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("BPM_BOSH_ROOT=%s", s.root))
	return cmd
}

func (s *Sandbox) Fixture(job, path string) error {
	configPath := filepath.Join(s.root, "jobs", job, "config", "bpm.yml")

	if err := os.MkdirAll(filepath.Dir(configPath), 0777); err != nil {
		return err
	}

	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}

func (s *Sandbox) Cleanup() {
	_ = os.RemoveAll(s.root)
}
