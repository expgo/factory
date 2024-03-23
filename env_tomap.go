//go:build !windows

package factory

import "strings"

// copied from github.com/caarlos0/env
func envToMap(envs []string) *map[string]string {
	r := map[string]string{}
	for _, e := range envs {
		p := strings.SplitN(e, "=", 2)
		if len(p) == 2 {
			r[p[0]] = p[1]
		}
	}
	return &r
}
