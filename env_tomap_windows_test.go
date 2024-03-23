//go:build windows

package factory

import (
	"reflect"
	"testing"
)

// copied from github.com/caarlos0/env
// On Windows, environment variables can start with '='. This test verifies this behavior without relying on a Windows environment.
// See env_windows.go in the Go source: https://github.com/golang/go/blob/master/src/syscall/env_windows.go#L58
func TestToMapWindows(t *testing.T) {
	envVars := []string{"=::=::\\", "=C:=C:\\test", "VAR=REGULARVAR", "FOO=", "BAR"}
	result := envToMap(envVars)

	if !reflect.DeepEqual(map[string]string{
		"=::": "::\\",
		"=C:": "C:\\test",
		"VAR": "REGULARVAR",
		"FOO": "",
	}, *result) {
		t.Error("envToMap test error")
	}
}
