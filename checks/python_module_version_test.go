// Copyright 2018 Mathew Robinson <chasinglogic@gmail.com>. All rights reserved. Use of this source code is
// governed by the Apache-2.0 license that can be found in the LICENSE file.

package checks

import (
	"os/exec"
	"testing"
)

func TestPythonModuleVersion(t *testing.T) {
	// First make sure python is in the PATH
	// Don't run these tests unless you are on a system with python installed
	_, err := exec.LookPath("python")
	if err != nil {
		return
	}

	tests := checkerTests{
		{
			Name: "pyyaml should be installed",
			Args: Args{
				"module": "yaml",
			},
		},
		{
			Name: "pyyaml should be at least version 3.00",
			Args: Args{
				"module":  "yaml",
				"version": "3.00",
			},
		},
		{
			Name: "pyyaml should be version 3.13",
			Args: Args{
				"module":       "yaml",
				"relationship": "eq",
				"version":      "3.13",
			},
		},
		{
			Name: "pyyaml should not be version 4.00",
			Args: Args{
				"module":  "yaml",
				"version": "4.00",
			},
			ShouldError: true,
		},
		{
			Name: "NOT_A_PYTHON_MODULE should be installed",
			Args: Args{
				"module": "NOT_A_PYTHON_MODULE",
			},
			ShouldError: true,
		},
	}

	runCheckerTests(t, tests, availableChecks["pip-installed"])
}
