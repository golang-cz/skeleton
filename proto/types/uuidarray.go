package types

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"
)

type UUIDArray []uuid.UUID

// Value converts the UUIDArray to a driver.Value.
func (a UUIDArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		uuidBytes := make([][]byte, n)
		for i, v := range a {
			uuidBytes[i] = v.Bytes()
		}
		return uuidBytes, nil
	}

	return "{}", nil
}

func (a *UUIDArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	strValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid type for UUIDArray: value is %v", value)
	}

	if strValue == "{}" {
		*a = []uuid.UUID{}
		return nil
	}

	// Remove the curly braces '{' and '}' and split the string by commas.
	strValue = strValue[1 : len(strValue)-1]
	uuidStrings := strings.Split(strValue, ",")

	result := make([]uuid.UUID, len(uuidStrings))
	for i, uuidStr := range uuidStrings {
		// Parse the UUID from the hexadecimal string.
		parsedUUID, err := uuid.FromString(uuidStr)
		if err != nil {
			return fmt.Errorf("to parse %s: %w", uuidStr, err)
		}
		result[i] = parsedUUID
	}

	*a = result
	return nil
}

func HaveSameElements(firstArray, secondArray UUIDArray) bool {
	if len(firstArray) != len(secondArray) {
		return false
	}

	count := make(map[uuid.UUID]int)
	for _, item := range firstArray {
		count[item]++
	}

	for _, item := range secondArray {
		count[item]--
		if count[item] < 0 {
			return false
		}
	}

	return true
}
