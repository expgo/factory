//go:build !windows

package factory

import (
	"reflect"
	"testing"
)

// copied from github.com/caarlos0/env
func TestUnix(t *testing.T) {
	envVars := []string{":=/test/unix", "PATH=:/test_val1:/test_val2", "VAR=REGULARVAR", "FOO=", "BAR"}
	result := envToMap(envVars)

	if !reflect.DeepEqual(map[string]string{
		":":    "/test/unix",
		"PATH": ":/test_val1:/test_val2",
		"VAR":  "REGULARVAR",
		"FOO":  "",
	}, *result) {
		t.Error("envToMap test error")
	}
}
