package timex

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type CtxKey string

const TimeCtxKey CtxKey = "now"

func Now(ctx context.Context) (context.Context, time.Time) {
	nowV := ctx.Value(TimeCtxKey)
	if nowV != nil {
		if now, ok := nowV.(time.Time); ok {
			return ctx, now
		} else {
			panic(fmt.Errorf("timex.Now() expected time.Time, got %T", nowV))
		}
	} else {
		now := time.Now().UTC()
		return context.WithValue(ctx, TimeCtxKey, now), now
	}
}

// Returns a random unix millisecond time based id. Includes jitter to prevent
// db hotspots and handle bursts of new ids being generated.
func SparseId(ctx context.Context) string {
	return MilliSparse(ctx)
}

// Returns a unix microsecond time based id that ensures later values are
// lexacographically smaller to appear first when listed from a database.
func LatestFirstId(ctx context.Context) string {
	return MicroLatestFirst(ctx)
}

// Uses the seconds of the current time to apply a year delta to the returned
// time. This helps with handling bursts of id generations without any
// conflicts.
func timeWithYearJumps(ctx context.Context) time.Time {
	_, now := Now(ctx)
	dur := time.Duration((-30+now.Second())*1e9) * 3600 * 24 * 365
	return now.Add(dur)
}

// Returns a random unix time based id. Includes jitter to prevent db hotspots
// and handle bursts of new ids being generated.
func UnixSparse(ctx context.Context) string {
	return rev(base36(jitter(timeWithYearJumps(ctx).Unix())))
}

// Returns a random unix millisecond time based id. Includes jitter to prevent
// db hotspots and handle bursts of new ids being generated.
func MilliSparse(ctx context.Context) string {
	return rev(base36(jitter(timeWithYearJumps(ctx).UnixMilli())))
}

// Returns a random unix microsecond time based id. Includes jitter to prevent
// db hotspots and handle bursts of new ids being generated.
func MicroSparse(ctx context.Context) string {
	return rev(base36(jitter(timeWithYearJumps(ctx).UnixMicro())))
}

// Returns a random unix microsecond time based id. Includes jitter to prevent
// db hotspots and handle bursts of new ids being generated.
func NanoSparse(ctx context.Context) string {
	return rev(base36(jitter(timeWithYearJumps(ctx).UnixNano())))
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

// Returns a unix time based id that ensures later values are lexacographically
// smaller to appear first when listed from a database.
func UnixLatestFirst(ctx context.Context) string {
	_, now := Now(ctx)
	dif := Y5138Unix - now.Unix()
	return fmt.Sprintf("%08s", base36(dif))
}

// Returns a unix millisecond time based id that ensures later values are
// lexacographically smaller to appear first when listed from a database.
func MilliLatestFirst(ctx context.Context) string {
	_, now := Now(ctx)
	dif := Y5138Milli - now.UnixMilli()
	return fmt.Sprintf("%09s", base36(dif))
}

// Returns a unix microsecond time based id that ensures later values are
// lexacographically smaller to appear first when listed from a database.
func MicroLatestFirst(ctx context.Context) string {
	_, now := Now(ctx)
	dif := Y5138Micro - now.UnixMicro()
	return fmt.Sprintf("%011s", base36(dif))
}
