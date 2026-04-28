package coremedia_test

import (
	"encoding/binary"
	"math"
	"testing"
	"time"

	"github.com/danielpaulus/quicktime_video_hack/screencapture/common"
	"github.com/danielpaulus/quicktime_video_hack/screencapture/coremedia"
	"github.com/stretchr/testify/assert"
)

// roundtrip serializes a single-entry StringKeyDict and parses it back.
func roundtrip(t *testing.T, key string, value interface{}) interface{} {
	t.Helper()
	dict := coremedia.StringKeyDict{Entries: []coremedia.StringKeyEntry{
		{Key: key, Value: value},
	}}
	serialized := coremedia.SerializeStringKeyDict(dict)
	parsed, err := coremedia.NewStringDictFromBytes(serialized)
	assert.NoError(t, err)
	assert.Len(t, parsed.Entries, 1)
	assert.Equal(t, key, parsed.Entries[0].Key)
	return parsed.Entries[0].Value
}

func TestURLValueRoundtrip(t *testing.T) {
	val := coremedia.URLValue("https://example.invalid/path?q=1")
	got := roundtrip(t, "k", val)
	assert.Equal(t, val, got)
}

func TestColorSpaceRoundtrip(t *testing.T) {
	for _, cs := range []coremedia.ColorSpace{
		coremedia.ColorSpaceDeviceRGB,
		coremedia.ColorSpaceDeviceGray,
	} {
		got := roundtrip(t, "k", cs)
		assert.Equal(t, cs, got)
	}
}

func TestDateValueRoundtrip(t *testing.T) {
	// Pick a non-trivial moment after the CFAbsoluteTime epoch.
	when := time.Date(2024, 6, 1, 12, 30, 45, 123456789, time.UTC)
	got := roundtrip(t, "k", when)
	gotTime, ok := got.(time.Time)
	if assert.True(t, ok, "expected time.Time, got %T", got) {
		// Allow sub-microsecond drift from float64 round-tripping.
		delta := gotTime.Sub(when).Nanoseconds()
		if delta < 0 {
			delta = -delta
		}
		assert.Less(t, delta, int64(time.Microsecond), "round-trip drift too large: %d ns", delta)
	}
}

func TestArrayValueRoundtrip(t *testing.T) {
	arr := []interface{}{
		"hello",
		true,
		common.NewNSNumberFromUInt32(42),
		[]byte{1, 2, 3, 4},
	}
	got := roundtrip(t, "k", arr)
	gotArr, ok := got.([]interface{})
	if !assert.True(t, ok, "expected []interface{}, got %T", got) {
		return
	}
	assert.Len(t, gotArr, len(arr))
	assert.Equal(t, "hello", gotArr[0])
	assert.Equal(t, true, gotArr[1])
	assert.Equal(t, common.NewNSNumberFromUInt32(42), gotArr[2])
	assert.Equal(t, []byte{1, 2, 3, 4}, gotArr[3])
}

func TestNestedArrayRoundtrip(t *testing.T) {
	inner := []interface{}{"a", "b"}
	outer := []interface{}{inner, true}
	got := roundtrip(t, "k", outer)
	gotArr, ok := got.([]interface{})
	if !assert.True(t, ok) {
		return
	}
	assert.Len(t, gotArr, 2)
	gotInner, ok := gotArr[0].([]interface{})
	if assert.True(t, ok, "expected nested []interface{}, got %T", gotArr[0]) {
		assert.Equal(t, []interface{}{"a", "b"}, gotInner)
	}
	assert.Equal(t, true, gotArr[1])
}

// Reference layout: the URL value atom on the wire should be exactly
//   [LE uint32 totalLen][LE uint32 0x75726c76][UTF-8 body bytes]
// matching Apple's sbufAtom_appendURLAtom (encoded via the strv-style
// helper but with magic 'urlv').
func TestURLValueWireLayout(t *testing.T) {
	val := coremedia.URLValue("X")
	dict := coremedia.StringKeyDict{Entries: []coremedia.StringKeyEntry{{Key: "k", Value: val}}}
	wire := coremedia.SerializeStringKeyDict(dict)
	// outer dict: 8-byte header
	// keyv:       8-byte header
	// strk(key):  8 + 1 = 9 bytes
	// urlv(val):  8 + 1 = 9 bytes
	// Find the urlv atom inside the dict.
	want := []byte{0x09, 0x00, 0x00, 0x00, 0x76, 0x6c, 0x72, 0x75, 'X'}
	assert.Contains(t, string(wire), string(want))
}

// Reference layout for clrs: 9 bytes total = 8-byte header + 1-byte body.
//   [LE 0x00000009][LE 0x636c7273 ('clrs' big-endian -> 'srlc' little-endian on the wire)][1=DeviceRGB | 0=DeviceGray]
func TestColorSpaceWireLayout(t *testing.T) {
	dict := coremedia.StringKeyDict{Entries: []coremedia.StringKeyEntry{{Key: "k", Value: coremedia.ColorSpaceDeviceRGB}}}
	wire := coremedia.SerializeStringKeyDict(dict)
	want := []byte{0x09, 0x00, 0x00, 0x00, 0x73, 0x72, 0x6c, 0x63, 0x01}
	assert.Contains(t, string(wire), string(want))
}

// Reference layout for dtev: 16 bytes total = 8-byte header + 8-byte LE
// IEEE-754 double of CFAbsoluteTime.
func TestDateValueWireLayout(t *testing.T) {
	// Exactly the CFAbsoluteTime epoch.
	val := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	dict := coremedia.StringKeyDict{Entries: []coremedia.StringKeyEntry{{Key: "k", Value: val}}}
	wire := coremedia.SerializeStringKeyDict(dict)

	// The dtev atom should encode 0.0 in the body.
	want := make([]byte, 16)
	binary.LittleEndian.PutUint32(want[0:], 16)
	binary.LittleEndian.PutUint32(want[4:], coremedia.DateValueMagic)
	binary.LittleEndian.PutUint64(want[8:], math.Float64bits(0.0))
	assert.Contains(t, string(wire), string(want))
}

// Sanity check: an unsupported CFType shouldn't crash; serializer
// should log + skip and return 0 length, mirroring Apple's
// FigSignalErrorAt3 fail-soft behavior.
func TestUnsupportedTypeDoesNotPanic(t *testing.T) {
	type bogusType struct{ x int }
	dict := coremedia.StringKeyDict{Entries: []coremedia.StringKeyEntry{{Key: "k", Value: bogusType{42}}}}
	assert.NotPanics(t, func() {
		_ = coremedia.SerializeStringKeyDict(dict)
	})
}

