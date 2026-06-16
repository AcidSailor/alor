package alor

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/quagmt/udecimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeUnmarshalJSON(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    time.Time
		wantErr bool
	}{
		{
			"naive zoneless",
			`"2026-06-18T00:00:00.0000000"`,
			time.Date(2026, 6, 18, 0, 0, 0, 0, time.UTC),
			false,
		},
		{
			"rfc3339",
			`"2021-10-13T09:00:00Z"`,
			time.Date(2021, 10, 13, 9, 0, 0, 0, time.UTC),
			false,
		},
		{
			"sentinel",
			`"9999-12-31T23:59:59.9999999"`,
			time.Date(9999, 12, 31, 23, 59, 59, 999999900, time.UTC),
			false,
		},
		{"null", `null`, time.Time{}, false},
		{"empty string", `""`, time.Time{}, false},
		{"garbage", `"not-a-time"`, time.Time{}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got Time
			err := json.Unmarshal([]byte(tc.in), &got)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.True(
				t,
				got.Equal(tc.want),
				"got %v want %v",
				got.Time,
				tc.want,
			)
		})
	}
}

func TestTimeNeverExpires(t *testing.T) {
	var sentinel Time
	require.NoError(
		t,
		json.Unmarshal([]byte(`"9999-12-31T23:59:59.9999999"`), &sentinel),
	)
	assert.True(t, sentinel.NeverExpires())

	var real Time
	require.NoError(
		t,
		json.Unmarshal([]byte(`"2026-06-18T00:00:00.0000000"`), &real),
	)
	assert.False(t, real.NeverExpires())

	// Sentinel matches on the 9999-12-31 date, not exact nanoseconds, so a
	// whole-second or differently-rounded sentinel still reports never-expires.
	var whole Time
	require.NoError(
		t,
		json.Unmarshal([]byte(`"9999-12-31T23:59:59"`), &whole),
	)
	assert.True(
		t,
		whole.NeverExpires(),
		"sentinel match tolerates precision drift",
	)
}

func TestTimeMarshalZero(t *testing.T) {
	var z Time
	b, err := json.Marshal(z)
	require.NoError(t, err)
	assert.Equal(t, "null", string(b))
}

func TestTimeRoundTrip(t *testing.T) {
	var got Time
	require.NoError(t, json.Unmarshal([]byte(`"2026-06-18T12:30:00Z"`), &got))
	b, err := json.Marshal(got)
	require.NoError(t, err)
	var back Time
	require.NoError(t, json.Unmarshal(b, &back))
	assert.True(t, got.Equal(back.Time))
}

// TestTimeMarshalTextNormalizesToUTC checks the encode side for outgoing date
// query params: MarshalText renders in UTC (Alor documents these inputs as UTC),
// overriding the promoted time.Time.MarshalText which would emit the value's own
// offset.
func TestTimeMarshalTextNormalizesToUTC(t *testing.T) {
	msk := time.FixedZone("MSK", 3*3600)
	tm := Time{time.Date(2021, 10, 13, 12, 0, 0, 0, msk)} // 09:00:00Z
	b, err := tm.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "2021-10-13T09:00:00Z", string(b))
}

func TestDecimalLosslessDecode(t *testing.T) {
	var v struct {
		P udecimal.Decimal `json:"p"`
	}
	require.NoError(t, json.Unmarshal([]byte(`{"p":0.0001}`), &v))
	assert.Equal(t, "0.0001", v.P.String())

	require.NoError(t, json.Unmarshal([]byte(`{"p":123456.789}`), &v))
	assert.Equal(t, "123456.789", v.P.String())
}

// TestGeneratedFileIsRuntimeFree guards the core design invariant: the generated
// models must never pull github.com/oapi-codegen/runtime.
func TestGeneratedFileIsRuntimeFree(t *testing.T) {
	b, err := os.ReadFile("models.gen.go")
	require.NoError(t, err)
	assert.False(t, strings.Contains(string(b), "oapi-codegen/runtime"),
		"models.gen.go must not import oapi-codegen/runtime")
}

func TestUUIDStrictDecode(t *testing.T) {
	var v struct {
		ID uuid.UUID `json:"id"`
	}
	// A valid GUID decodes.
	require.NoError(
		t,
		json.Unmarshal(
			[]byte(`{"id":"123e4567-e89b-12d3-a456-426614174000"}`),
			&v,
		),
	)
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", v.ID.String())

	// Strictness: an empty/garbage value errors the whole decode.
	err := json.Unmarshal([]byte(`{"id":""}`), &v)
	require.Error(t, err)
}
