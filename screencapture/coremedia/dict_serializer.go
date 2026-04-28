package coremedia

import (
	"encoding/binary"
	"log"
	"math"
	"time"

	"github.com/danielpaulus/quicktime_video_hack/screencapture/common"
)

//SerializeStringKeyDict serializes a StringKeyDict into a []byte
func SerializeStringKeyDict(stringKeyDict StringKeyDict) []byte {
	buffer := make([]byte, 1024*1024)
	var slice = buffer[8:]
	var index = 0
	for _, entry := range stringKeyDict.Entries {
		keyvaluePair := slice[index+8:]
		keyLength := serializeKey(entry.Key, keyvaluePair)
		valueLength := serializeValue(entry.Value, keyvaluePair[keyLength:])
		common.WriteLengthAndMagic(slice[index:], keyLength+valueLength+8, KeyValuePairMagic)
		index += 8 + valueLength + keyLength
	}
	dictSizePlusHeaderAndLength := index + 4 + 4
	common.WriteLengthAndMagic(buffer, dictSizePlusHeaderAndLength, DictionaryMagic)

	return buffer[:dictSizePlusHeaderAndLength]
}

// serializeValue mirrors Apple's `sbufAtom_appendCFTypeAtom` in
// FigSampleBufferAtomSerialization.c. The branches cover every CFType
// Apple's serializer recognizes; on an unsupported type Apple calls
// FigSignalErrorAt3 with "sbuf serializer encountered an unsupported
// dictionary cftype" — we log here and return 0 to match that
// (best-effort, fail-soft) semantics rather than panicking.
func serializeValue(value interface{}, bytes []byte) int {
	switch value := value.(type) {
	case bool:
		common.WriteLengthAndMagic(bytes, 9, BooleanValueMagic)
		var boolValue uint32
		if value {
			boolValue = 1
		}
		binary.LittleEndian.PutUint32(bytes[8:], boolValue)
		return 9
	case ColorSpace:
		// `clrs` atom: 8-byte header + 1 byte body (1=RGB, 0=Gray).
		// Apple errors on anything other than DeviceRGB/DeviceGray; here we
		// trust the caller passed one of the two ColorSpace constants.
		common.WriteLengthAndMagic(bytes, 9, ColorSpaceMagic)
		bytes[8] = byte(value)
		return 9
	case common.NSNumber:
		numberBytes := value.ToBytes()
		length := len(numberBytes) + 8
		common.WriteLengthAndMagic(bytes, length, common.NumberValueMagic)
		copy(bytes[8:], numberBytes)
		return length
	case string:
		stringValue := value
		length := len(stringValue) + 8
		common.WriteLengthAndMagic(bytes, length, StringValueMagic)
		copy(bytes[8:], stringValue)
		return length
	case URLValue:
		s := string(value)
		length := len(s) + 8
		common.WriteLengthAndMagic(bytes, length, URLValueMagic)
		copy(bytes[8:], s)
		return length
	case time.Time:
		// `dtev` atom: 8-byte header + IEEE-754 double seconds since
		// CFAbsoluteTime epoch (2001-01-01 00:00:00 UTC).
		secs := value.Sub(cfAbsoluteEpoch).Seconds()
		common.WriteLengthAndMagic(bytes, 16, DateValueMagic)
		binary.LittleEndian.PutUint64(bytes[8:], math.Float64bits(secs))
		return 16
	case []byte:
		byteValue := value
		length := len(byteValue) + 8
		common.WriteLengthAndMagic(bytes, length, DataValueMagic)
		copy(bytes[8:], byteValue)
		return length
	case []interface{}:
		// `aray` atom: 8-byte header + concatenated child value atoms,
		// recursing through serializeValue for each element.
		written := 8
		for _, elem := range value {
			written += serializeValue(elem, bytes[written:])
		}
		common.WriteLengthAndMagic(bytes, written, ArrayValueMagic)
		return written
	case StringKeyDict:
		dictValue := SerializeStringKeyDict(value)
		copy(bytes, dictValue)
		return len(dictValue)
	default:
		// Match Apple's failure mode: log and skip rather than panic.
		// Apple's code emits "sbuf serializer encountered an unsupported
		// dictionary cftype" to os_log and returns an error code.
		log.Printf("dict_serializer: unsupported cftype %T while serializing dict value: %v", value, value)
		return 0
	}
}

func serializeKey(key string, bytes []byte) int {
	keyLength := len(key) + 8
	common.WriteLengthAndMagic(bytes, keyLength, StringKey)
	copy(bytes[8:], key)
	return keyLength
}
