package sanitize

import (
	"net/url"
)

// FilterPIIParams deletes all params that contain user PII, including:
// state,token,jwt,email and hash params
func FilterPIIParams(values url.Values) url.Values {
	for _, param := range []string{
		"state",
		"token",
		"jwt",
		"email",
		"hash",
	} {
		if values.Get(param) != "" {
			values.Set(param, "XXX")
		}
	}
	return values
}
