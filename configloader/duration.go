package configloader

import "time"

var DurationHandler = WithTypeHandler(time.ParseDuration)
