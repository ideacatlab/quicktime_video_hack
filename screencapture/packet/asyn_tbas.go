package packet

import (
	"encoding/binary"
	"fmt"
)

// AsynTbasPacket carries a "set timebase" notification. The device tells
// the host to use a specific CMTimebaseRef (minted on the device) as the
// source-of-truth timebase for subsequent sample-buffer timing.
//
// Reverse-engineered from Apple's host-side handler in
// MediaToolbox.framework at vaddr 0x191584148 on macOS 15.5: on receipt
// the host stashes DeviceTimebaseRef at offset +0x78 of its per-stream
// state and uses it as the timebase identity for subsequent operations.
// (Earlier qvh versions called this field "SomeOtherRef" because its
// purpose was unknown.)
type AsynTbasPacket struct {
	// ClockRef is the per-stream clock the ASYN packet is scoped to.
	ClockRef CFTypeID

	// DeviceTimebaseRef is the CFTypeID of the CMTimebase on the iOS
	// device. 8 bytes from payload offset 0.
	DeviceTimebaseRef CFTypeID
}

// NewAsynTbasPacketFromBytes parses a AsynTbasPacket from bytes.
func NewAsynTbasPacketFromBytes(data []byte) (AsynTbasPacket, error) {
	var packet AsynTbasPacket
	remainingBytes, clockRef, err := ParseAsynHeader(data, TBAS)
	if err != nil {
		return packet, err
	}
	if len(remainingBytes) < 8 {
		return packet, fmt.Errorf("TBAS payload too short: %d bytes (need 8)", len(remainingBytes))
	}
	packet.ClockRef = clockRef
	packet.DeviceTimebaseRef = binary.LittleEndian.Uint64(remainingBytes[:8])
	return packet, nil
}

func (sp AsynTbasPacket) String() string {
	return fmt.Sprintf("ASYN_TBAS{ClockRef:%x, DeviceTimebaseRef:%x}", sp.ClockRef, sp.DeviceTimebaseRef)
}
