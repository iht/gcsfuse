// Copyright 2015 Google Inc. All Rights Reserved.
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

package main

import (
	"fmt"
	"log"
	"os/exec"
	"path"
)

// Given a directory containing release binaries, create an appropriate .tar.gz
// file.
func packageTarball(
	binDir string,
	version string,
	osys string,
	arch string,
	outputDir string) (err error) {
	// Choose an output file name.
	outputFile := path.Join(
		outputDir,
		fmt.Sprintf("gcsfuse_v%s_%s_%s.tar.gz", version, osys, arch))

	log.Printf("Writing tarball to %s", outputFile)

	// Run tar.
	cmd := exec.Command(
		"tar",
		"zcvf",
		outputFile,
		".")

	cmd.Dir = binDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("tar: %v\nOutput:\n%s", err, output)
		return
	}

	return
}

// Given a directory containing release binaries, create an appropriate .deb
// file.
func packageDeb(
	binDir string,
	version string,
	osys string,
	arch string,
	outputDir string) (err error) {
	log.Println("Building a .deb file")

	// Call fpm.
	cmd := exec.Command(
		"fpm",
		"-s", "dir",
		"-t", "deb",
		"-n", "gcsfuse",
		"-C", binDir,
		"-v", version,
		"-d", "fuse (>= 2.9.2)",
		"--prefix", "/usr/bin",
		"--vendor", "",
		"--maintainer", "Aaron Jacobs <jacobsa@google.com>",
		"--url", "https://github.com/googlecloudplatform/gcsfuse",
		"--description", "A user-space file system for interacting with Google Cloud Storage.",
	)

	cmd.Dir = outputDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("fpm: %v\nOutput:\n%s", err, output)
		return
	}

	return
}
