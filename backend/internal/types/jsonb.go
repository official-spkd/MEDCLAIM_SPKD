package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONB memetakan kolom jsonb; isinya diteruskan apa adanya ke JSON respons.
type JSONB json.RawMessage

func (JSONB) GormDataType() string { return "jsonb" }

func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

func (j *JSONB) Scan(v any) error {
	switch x := v.(type) {
	case nil:
		*j = nil
	case []byte:
		*j = append((*j)[:0], x...)
	case string:
		*j = JSONB(x)
	default:
		return fmt.Errorf("types.JSONB: tipe tidak didukung %T", v)
	}
	return nil
}

func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

func (j *JSONB) UnmarshalJSON(b []byte) error {
	*j = append((*j)[:0], b...)
	return nil
}
