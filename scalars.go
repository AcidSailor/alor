package alor

import (
	"bytes"
	"fmt"
	"time"
)

// alorTimeLayouts are tried in order by Time.UnmarshalJSON. Alor emits naive,
// zone-less timestamps (e.g. "2026-06-18T00:00:00.0000000"); time.Parse accepts
// optional trailing fractional seconds against the zone-less layout, so one
// naive layout covers both whole-second and fractional forms. RFC3339Nano is
// tried first for the occasional zoned value.
var alorTimeLayouts = []string{
	time.RFC3339Nano,
	"2006-01-02T15:04:05",
}

// Time wraps time.Time with an UnmarshalJSON lenient enough for Alor's wire
// shapes; time.Time's methods promote, so a Time is usable wherever a time.Time
// is. Response-only fields generate as *Time: nil is absent, a parsed value is
// real, and NeverExpires reports the sentinel.
type Time struct{ time.Time }

// UnmarshalJSON accepts JSON null/empty (no-op, leaving the zero value), an
// RFC3339 timestamp, or Alor's naive zone-less layout (interpreted as UTC). It
// errors only when a non-empty, non-null value matches no known layout.
func (t *Time) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		return nil
	}
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	if len(b) == 0 {
		return nil
	}
	s := string(b)
	// Wrap the last layout's error: the naive zone-less layout is tried last
	// and is what Alor actually emits, so its diagnostic is the useful one
	// (RFC3339Nano, tried first, is the occasional zoned case).
	var lastErr error
	for _, layout := range alorTimeLayouts {
		parsed, err := time.Parse(layout, s)
		if err == nil {
			t.Time = parsed
			return nil
		}
		lastErr = err
	}
	return fmt.Errorf("alor: cannot parse time %q: %w", s, lastErr)
}

// MarshalJSON emits RFC3339Nano (or null for the zero value). These fields are
// response-only and never sent to Alor; this exists for round-trip symmetry.
func (t Time) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	b := make([]byte, 0, len(time.RFC3339Nano)+2)
	b = append(b, '"')
	b = t.AppendFormat(b, time.RFC3339Nano)
	b = append(b, '"')
	return b, nil
}

// text renders the value in UTC as RFC3339Nano — the form Alor accepts for its
// date/date-time query inputs (documented UTC). Total: time.Format never fails,
// so setTime (params.go), which has no error channel, can encode through it
// directly.
func (t Time) text() string { return t.UTC().Format(time.RFC3339Nano) }

// MarshalText implements encoding.TextMarshaler over text, overriding the
// promoted time.Time.MarshalText so an outgoing value is always normalised to
// UTC regardless of its location. Encode-only, mirroring MarshalJSON.
func (t Time) MarshalText() ([]byte, error) { return []byte(t.text()), nil }

// NeverExpires reports whether the value is Alor's "never expires" sentinel:
// .NET's DateTime.MaxValue (9999-12-31T23:59:59.9999999), emitted for
// instruments without an expiry (perpetual securities, etc.). It matches on the
// 9999-12-31 UTC date, not exact nanoseconds, so it stays correct at any
// sub-second precision — year 9999 is the sentinel and means nothing else.
func (t Time) NeverExpires() bool {
	y, m, d := t.UTC().Date()
	return y == 9999 && m == time.December && d == 31
}
