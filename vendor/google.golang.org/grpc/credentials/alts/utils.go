/*
 *
 * Copyright 2018 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package alts

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const (
	linuxProductNameFile     = "/sys/class/dmi/id/product_name"
	windowsCheckCommand      = "powershell.exe"
	windowsCheckCommandArgs  = "Get-WmiObject -Class Win32_BIOS"
	powershellOutputFilter   = "Manufacturer"
	windowsManufacturerRegex = ":(.*)"
	windowsCheckTimeout      = 30 * time.Second
)

type platformError string

func (k platformError) Error() string {
	return fmt.Sprintf("%v is not supported", k)
}

var (
	// The following two variables will be reassigned in tests.
	runningOS          = runtime.GOOS
	manufacturerReader = func() (io.Reader, error) {
		switch runningOS {
		case "linux":
			return os.Open(linuxProductNameFile)
		case "windows":
			cmd := exec.Command(windowsCheckCommand, windowsCheckCommandArgs)
			out, err := cmd.Output()
			if err != nil {
				return nil, err
			}

			for _, line := range strings.Split(strings.TrimSuffix(string(out), "\n"), "\n") {
				if strings.HasPrefix(line, powershellOutputFilter) {
					re := regexp.MustCompile(windowsManufacturerRegex)
					name := re.FindString(line)
					name = strings.TrimLeft(name, ":")
					return strings.NewReader(name), nil
				}
			}

			return nil, errors.New("cannot determine the machine's manufacturer")
		default:
			return nil, platformError(runningOS)
		}
	}
	vmOnGCP bool
)

// isRunningOnGCP checks whether the local system, without doing a network request is
// running on GCP.
func isRunningOnGCP() bool {
	manufacturer, err := readManufacturer()
	if err != nil {
		log.Fatalf("failure to read manufacturer information: %v", err)
	}
	name := string(manufacturer)
	switch runningOS {
	case "linux":
		name = strings.TrimSpace(name)
		return name == "Google" || name == "Google Compute Engine"
	case "windows":
		name = strings.Replace(name, " ", "", -1)
		name = strings.Replace(name, "\n", "", -1)
		name = strings.Replace(name, "\r", "", -1)
		return name == "Google"
	default:
		log.Fatal(platformError(runningOS))
	}
	return false
}

func readManufacturer() ([]byte, error) {
	reader, err := manufacturerReader()
	if err != nil {
		return nil, err
	}
	if reader == nil {
		return nil, errors.New("got nil reader")
	}
	manufacturer, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed reading %v: %v", linuxProductNameFile, err)
	}
	return manufacturer, nil
}