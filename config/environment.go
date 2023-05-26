package config

import (
	"fmt"
	"strings"
)

type Environment int

const (
	EnvLocal Environment = iota
	EnvTest
	EnvCI
	EnvProd
)

var environments = []string{
	"local",      // 0
	"test",       // 1
	"ci",         // 2
	"production", // 3
}

func (e Environment) IsLocal() bool {
	return e == EnvLocal
}

// IsProduction reports whether the current environment is production.
func (e Environment) IsProduction() bool {
	switch e {
	case EnvProd:
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
	return fmt.Errorf(
		"unknown environment:%q\n  supported: %s",
		enum,
		strings.Join(environments, ", "),
	)
}
