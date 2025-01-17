// Copyright 2021 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Provides integration tests when implicit_dir flag is set.
package implicitdir_test

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/googlecloudplatform/gcsfuse/tools/util"
)

var testBucket = flag.String("testbucket", "", "The GCS bucket used for the test.")

var (
	testDir string
	binFile string
	logFile string
	mntDir  string
)

func setUpTestDir() error {
	var err error
	testDir, err = ioutil.TempDir("", "gcsfuse_readwrite_test_")
	if err != nil {
		return fmt.Errorf("TempDir: %w\n", err)
	}

	err = util.BuildGcsfuse(testDir)
	if err != nil {
		return fmt.Errorf("BuildGcsfuse(%q): %w\n", testDir, err)
	}

	binFile = path.Join(testDir, "bin/gcsfuse")
	logFile = path.Join(testDir, "gcsfuse.log")
	mntDir = path.Join(testDir, "mnt")

	err = os.Mkdir(mntDir, 0755)
	if err != nil {
		return fmt.Errorf("Mkdir(%q): %v\n", mntDir, err)
	}
	return nil
}

func mountGcsfuse() error {
	mountCmd := exec.Command(
		binFile,
		"--implicit-dirs",
		"--debug_gcs",
		"--debug_fs",
		"--debug_fuse",
		"--log-file="+logFile,
		"--log-format=text",
		*testBucket,
		mntDir,
	)
	output, err := mountCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cannot mount gcsfuse: %w\n", err)
	}
	if lines := bytes.Count(output, []byte{'\n'}); lines > 1 {
		return fmt.Errorf("mount output: %q\n", output)
	}
	return nil
}

func unMount() error {
	fusermount, err := exec.LookPath("fusermount")
	if err != nil {
		return fmt.Errorf("cannot find fusermount: %w", err)
	}
	cmd := exec.Command(fusermount, "-uz", mntDir)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("fusermount error: %w", err)
	}
	return nil
}

func clearKernelCache() error {
	if _, err := os.Stat("/proc/sys/vm/drop_caches"); err != nil {
		log.Printf("Kernel cache file not found: %v", err)
		// No need to stop the test execution if cache file is not found. Further
		// reads will be served from kernel cache.
		return nil
	}

	// sudo permission is required to clear kernel page cache.
	cmd := exec.Command("sudo", "sh", "-c", "echo 3 > /proc/sys/vm/drop_caches")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clear kernel cache failed with error: %w", err)
	}
	return nil
}

func TestMain(m *testing.M) {
	flag.Parse()

	if *testBucket == "" {
		log.Printf("--testbucket must be specified")
		os.Exit(0)
	}

	if err := setUpTestDir(); err != nil {
		log.Printf("setUpTestDir: %v\n", err)
		os.Exit(1)
	}

	if err := mountGcsfuse(); err != nil {
		log.Printf("mountGcsfuse: %v\n", err)
		os.Exit(1)
	}

	log.Printf("Test log: %s\n", logFile)
	ret := m.Run()

	// Delete all files from mntDir to delete files from gcs bucket.
	os.RemoveAll(mntDir)
	unMount()

	os.Exit(ret)
}
