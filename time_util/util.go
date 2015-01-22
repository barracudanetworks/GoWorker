package time_util

import (
	"errors"
	"fmt"
	"time"
)

var (
	TIME_FORMAT = time.RFC3339
	BAD_FORMAT  = errors.New("Date time in wrong format")
)

// TimeToName given a time, return the file name representing that time
// format: RFC3339
// suffix can be any string. This will be stripped when converting from
// file name to time later. The suffix is there merely to diferentiate
// between files at similar times
func TimeToName(t time.Time, suffix string) []byte {
	return []byte(fmt.Sprintf("%s#%s", t.Format(TIME_FORMAT), suffix))
}
