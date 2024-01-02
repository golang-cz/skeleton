package types

import (
	"database/sql/driver"

	"github.com/upper/db/v4/adapter/postgresql"
)

type JSONBMap map[string]string

func (m JSONBMap) Value() (driver.Value, error) {
	return postgresql.JSONBValue(m)
}

func (m *JSONBMap) Scan(src interface{}) error {
	*m = map[string]string(nil)
	return postgresql.ScanJSONB(m, src)
}
