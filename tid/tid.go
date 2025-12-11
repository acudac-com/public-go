// Package tid provides methods for generating time based ids.
package tid

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// UnixSparse returns a random unix time based id. Includes jitter to prevent db hotspots
// and handle bursts of new ids being generated.
// Leave randomNr nil for default jitter or keep in range [-1e6, 1e6]
func UnixSparse(now time.Time) string {
	nowInt := jumpYears(now).Unix()
	return rev(base36(jitter(nowInt)))
}

// MilliSparse returns a random unix millisecond time based id. Includes jitter to prevent
// db hotspots and handle bursts of new ids being generated.
// Leave randomNr nil for default jitter or keep in range [-1e6, 1e6]
func MilliSparse(now time.Time) string {
	nowInt := jumpYears(now).UnixMilli()
	return rev(base36(jitter(nowInt)))
}

// MicroSparse returns a random unix microsecond time based id. Includes jitter to prevent
// db hotspots and handle bursts of new ids being generated.
// Leave randomNr nil for default jitter or keep in range [-1e6, 1e6]
func MicroSparse(now time.Time) string {
	nowInt := jumpYears(now).UnixMicro()
	return rev(base36(jitter(nowInt)))
}

// NanoSparse returns a random unix nanosecond time based id. Includes jitter to prevent
// db hotspots and handle bursts of new ids being generated.
// Leave randomNr nil for default jitter or keep in range [-1e6, 1e6]
func NanoSparse(now time.Time) string {
	nowPtr := jumpYears(now)
	nowInt := nowPtr.UnixNano()
	return rev(base36(jitter(nowInt)))
}

// Uses the seconds of the current time to apply a year delta to the returned
// time. This helps with handling bursts of id generations without any
// conflicts.
func jumpYears(now time.Time) time.Time {
	dur := time.Duration((-30+now.Second())*1e9) * 3600 * 24 * 365
	return now.Add(dur)
}

// RandomFunc is a function that returns a random number in the range [0, 2e6)
// Do not change this, except for testing purposes.
var RandomFunc = func() int {
	return rand.Intn(2e6)
}

// jitter adds/subtracts a random number to the value
func jitter(value int64) int64 {
	jitterVal := int(-1e6)
	jitterVal += RandomFunc()
	value += int64(jitterVal)
	return value
}

// base36 converts an int64 to a base36 string
// e.g. base36(123456789) = "1t4g5v"
func base36(value int64) string {
	return strconv.FormatInt(value, 36)
}

// Reverses the string to prevent hotspotting
func rev(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	s = string(runes)
	return s
}

const (
	Y5138Unix  int64 = 99999999999       // 19xtf1tr = 8 chars
	Y5138Milli int64 = 99999999999000    // zg3d62qe0 = 9 chars
	Y5138Micro int64 = 99999999999000000 // rcn1hsrx0w0 = 11 chars
)

// UnixLatestFirst returns a unix time based id that ensures later values are lexacographically
// smaller to appear first when listed from a database.
func UnixLatestFirst(now time.Time) *string {
	dif := Y5138Unix - now.Unix()
	id := fmt.Sprintf("%08s", base36(dif))
	return &id
}

// MilliLatestFirst returns a unix millisecond time based id that ensures later values are
// lexacographically smaller to appear first when listed from a database.
func MilliLatestFirst(now time.Time) *string {
	dif := Y5138Milli - now.UnixMilli()
	id := fmt.Sprintf("%09s", base36(dif))
	return &id
}

// MicroLatestFirst returns a unix microsecond time based id that ensures later values are
// lexacographically smaller to appear first when listed from a database.
func MicroLatestFirst(now time.Time) *string {
	dif := Y5138Micro - now.UnixMicro()
	id := fmt.Sprintf("%011s", base36(dif))
	return &id
}
