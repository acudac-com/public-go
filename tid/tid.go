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
func UnixSparse(now time.Time) string {
	return rev(base36(jitter(timeWithYearJumps(now).Unix())))
}

// MilliSparse returns a random unix millisecond time based id. Includes jitter to prevent
// db hotspots and handle bursts of new ids being generated.
func MilliSparse(now time.Time) string {
	return rev(base36(jitter(timeWithYearJumps(now).UnixMilli())))
}

// MicroSparse returns a random unix microsecond time based id. Includes jitter to prevent
// db hotspots and handle bursts of new ids being generated.
func MicroSparse(now time.Time) string {
	return rev(base36(jitter(timeWithYearJumps(now).UnixMicro())))
}

// NanoSparse returns a random unix microsecond time based id. Includes jitter to prevent
// db hotspots and handle bursts of new ids being generated.
func NanoSparse(now time.Time) string {
	return rev(base36(jitter(timeWithYearJumps(now).UnixNano())))
}

// Uses the seconds of the current time to apply a year delta to the returned
// time. This helps with handling bursts of id generations without any
// conflicts.
func timeWithYearJumps(now time.Time) time.Time {
	dur := time.Duration((-30+now.Second())*1e9) * 3600 * 24 * 365
	return now.Add(dur)
}

func jitter(value int64) int64 {
	dif := -1e6 + rand.Intn(2e6)
	value = value + int64(dif)
	return value
}

func base36(value int64) string {
	return strconv.FormatInt(value, 36)
}

// Reverses the string to prevent hotspotting
func rev(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

const (
	Y5138Unix  int64 = 99999999999       // 19xtf1tr = 8 chars
	Y5138Milli int64 = 99999999999000    // zg3d62qe0 = 9 chars
	Y5138Micro int64 = 99999999999000000 // rcn1hsrx0w0 = 11 chars
)

// UnixLatestFirst returns a unix time based id that ensures later values are lexacographically
// smaller to appear first when listed from a database.
func UnixLatestFirst(now time.Time) string {
	dif := Y5138Unix - now.Unix()
	return fmt.Sprintf("%08s", base36(dif))
}

// MilliLatestFirst returns a unix millisecond time based id that ensures later values are
// lexacographically smaller to appear first when listed from a database.
func MilliLatestFirst(now time.Time) string {
	dif := Y5138Milli - now.UnixMilli()
	return fmt.Sprintf("%09s", base36(dif))
}

// MicroLatestFirst returns a unix microsecond time based id that ensures later values are
// lexacographically smaller to appear first when listed from a database.
func MicroLatestFirst(now time.Time) string {
	dif := Y5138Micro - now.UnixMicro()
	return fmt.Sprintf("%011s", base36(dif))
}
