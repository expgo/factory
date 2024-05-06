package factory

import (
	"os"
)

func init() {
	NamedSingleton[map[string]string]("env").SetInitFunc(func() any { return envToMap(os.Environ()) })
}
