// Package types berisi tipe kolom khusus yang dipakai lintas modul.
package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

const layoutDate = "2006-01-02"

// Date memetakan kolom `date` PostgreSQL. Di JSON tampil sebagai "YYYY-MM-DD"
// (bentuk yang diharapkan frontend), bukan timestamp.
type Date struct{ time.Time }

func ParseDate(s string) (Date, error) {
	t, err := time.Parse(layoutDate, s)
	return Date{t}, err
}

func (Date) GormDataType() string { return "date" }

func (d Date) String() string { return d.Time.Format(layoutDate) }

func (d Date) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}
	return d.Time.Format(layoutDate), nil
}

func (d *Date) Scan(v any) error {
	switch x := v.(type) {
	case nil:
		*d = Date{}
	case time.Time:
		*d = Date{time.Date(x.Year(), x.Month(), x.Day(), 0, 0, 0, 0, time.UTC)}
	case string:
		t, err := time.Parse(layoutDate, x[:min(len(x), 10)])
		if err != nil {
			return err
		}
		*d = Date{t}
	case []byte:
		return d.Scan(string(x))
	default:
		return fmt.Errorf("types.Date: tipe tidak didukung %T", v)
	}
	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.Time.Format(layoutDate))
}

func (d *Date) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s == "" {
		*d = Date{}
		return nil
	}
	return d.Scan(s)
}
