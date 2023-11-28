package config

import (
	"fmt"
	"strings"
)

type Environment int

const (
	EnvLocal Environment = iota // 0
	EnvTest                     // 1
	EnvCI                       // 2
	EnvDEV1                     // 3
	EnvEU1                      // 4
)

var environments = []string{
	"local", // 0
	"test",  // 1
	"ci",    // 2
	"dev1",  // 4
	"eu1",   // 8
}

func (e Environment) IsLocal() bool {
	return e == EnvLocal
}

// IsDevlabs reports whether the current environment is on aws devlabs.
func (e Environment) IsDevlabs() bool {
	switch e {
	case EnvDEV1, EnvDEV2, EnvSIT1:
		return true
	default:
		return false
	}
}

// IsProduction reports whether the current environment is production.
func (e Environment) IsProduction() bool {
	switch e {
	case EnvNA1, EnvNA2, EnvEU1, EnvEU2, EnvAP2, EnvAP3:
		return true
	default:
		return false
	}
}

// IsNotProduction reports whether the current environment is not production.
func (e Environment) IsNotProduction() bool {
	return !e.IsProduction()
}

// String returns the string value of the environment.
func (e Environment) String() string {
	return environments[e]
}

// MarshalText satisfies TextMarshaler
func (e Environment) MarshalText() ([]byte, error) {
	return []byte(e.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (e *Environment) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(environments); i++ {
		if enum == environments[i] {
			*e = Environment(i)
			return nil
		}
	}
	return fmt.Errorf("unknown environment:%q\n  supported: %s", enum, strings.Join(environments, ", "))
}
