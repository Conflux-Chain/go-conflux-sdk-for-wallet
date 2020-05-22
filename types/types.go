package richtypes

import (
	"fmt"
	"time"
)

// JSONTime is created for json marshal
type JSONTime int64

// MarshalJSON implements interface Marshaler
func (t JSONTime) MarshalJSON() ([]byte, error) {
	_t := time.Unix(int64(t), 0)
	loc, _ := time.LoadLocation("Greenwich")
	stamp := fmt.Sprintf("\"%s\"", _t.In(loc).Format("2006-01-02T15:04:05"))
	return []byte(stamp), nil
}
