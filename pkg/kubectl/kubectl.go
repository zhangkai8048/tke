/*
 * Tencent is pleased to support the open source community by making TKEStack
 * available.
 *
 * Copyright (C) 2012-2020 Tencent. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the “License”); you may not use
 * this file except in compliance with the License. You may obtain a copy of the
 * License at
 *
 * https://opensource.org/licenses/Apache-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an “AS IS” BASIS, WITHOUT
 * WARRANTIES OF ANY KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations under the License.
 */

package kubectl

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"tkestack.io/tke/pkg/util/env"
)

const (
	envKey       = "KUBECTL"
	defaultValue = "kubectl"
)

func cmd() *exec.Cmd {
	return exec.Command(bin())
}

func bin() string {
	return env.GetEnvAsStringOrFallback(envKey, defaultValue)
}

// Validate validates a k8s object.
// func Validate(object []byte) ([]byte, error) {
// 	cmd := cmd()
// 	cmd.Args = append(cmd.Args, "apply", "--dry-run=client", "-f", "-")
// 	cmd.Stdin = bytes.NewBuffer(object)

// 	return cmd.CombinedOutput()
// }

// Validate validates a k8s object.
func Validate(object []byte) ([]byte, error) {
	cmd := cmd()
	opt := os.Getenv("OPT")

	switch opt {
	case "validate":
		cmd.Args = append(cmd.Args, "apply", "--dry-run", "-f", "-")
	case "delete":
		cmd.Args = append(cmd.Args, opt, "-f", "-")
	case "apply":
		cmd.Args = append(cmd.Args, opt, "-f", "-")
	default:
		cmd.Args = append(cmd.Args, "apply", "--dry-run", "-f", "-")
	}

	fmt.Println("opt", opt, cmd.Args)
	cmd.Stdin = bytes.NewBuffer(object)

	return cmd.CombinedOutput()

}
