package time_util

import (
	"testing"
	"time"
)

var (
	test_time = time.Date(2014, 12, 22, 0, 0, 0, 0, time.Local)
)

func TestTimeToName(t *testing.T) {
	if string(TimeToName(test_time, "test")) != "2014-12-22T00:00:00-05:00#test" {
		t.Fail()
	}
}
