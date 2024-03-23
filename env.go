package factory

import (
	"os"
)

func init() {
	NamedSingleton[map[string]string]("env").SetInitFunc(func() *map[string]string { return envToMap(os.Environ()) })
}
