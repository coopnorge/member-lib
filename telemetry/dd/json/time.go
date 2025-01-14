package json

import (
	"fmt"
	"time"
)

type LogTime time.Time

func (t LogTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`%q`, time.Time(t).UTC().Format(time.RFC3339))), nil
}
