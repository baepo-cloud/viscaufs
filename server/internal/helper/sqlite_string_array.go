package helper

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type SQLiteStringArray []string

func (sa *SQLiteStringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = SQLiteStringArray{}
		return nil
	}

	str, ok := value.(string)
	if !ok {
		bytes, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("failed to scan SQLiteStringArray: %T is not a string or []byte", value)
		}
		str = string(bytes)
	}

	if str == "" {
		*sa = SQLiteStringArray{}
	} else {
		*sa = strings.Split(str, ",")
	}

	return nil
}

func (sa SQLiteStringArray) Value() (driver.Value, error) {
	if len(sa) == 0 {
		return "", nil
	}
	return strings.Join(sa, ","), nil
}
