package packet

import (
	"encoding/binary"
	"fmt"

	"github.com/danielpaulus/quicktime_video_hack/screencapture/coremedia"
)

// AsynTjmpPacket carries a "time jump" notification from the device. The
// device sends one whenever its timebase discontinuously moves — for
// example on a clock reset, seek, or pause/resume — so the host can
// re-anchor a read-only mirror of the device's timebase to stay in sync.
//
// Reverse-engineered from Apple's host-side handler in
// MediaToolbox.framework (inside the inbound dispatcher near
// _FigNeroTeardown on macOS 15.5). Apple receives the same payload and
// calls CMTimebaseCreateReadOnlyTimebaseWithFlags with (clock=DeviceTimebaseID,
// anchorTime=AnchorTime, currentTime=CurrentTime).
type AsynTjmpPacket struct {
	// ClockRef is the per-stream clock the ASYN packet is scoped to (8
	// bytes from the ASYN header), echoing the local clock the host
	// sent in CVRP_RPLY.
	ClockRef CFTypeID

	// DeviceTimebaseID is the CFTypeID of the timebase on the iOS
	// device that just jumped. 8 bytes at payload offset 0.
	DeviceTimebaseID CFTypeID

	// AnchorTime is the anchor passed to
	// CMTimebaseCreateReadOnlyTimebaseWithFlags. 24 bytes at payload
	// offset 8.
	AnchorTime coremedia.CMTime

	// CurrentTime is the current time of the timebase at the moment of
	// the jump. 24 bytes at payload offset 32.
	CurrentTime coremedia.CMTime
}

// NewAsynTjmpPacketFromBytes parses an AsynTjmpPacket from a complete
// ASYN packet (length+magic+clockRef+TJMP magic+payload). The payload
// must be exactly 56 bytes per the wire format.
func NewAsynTjmpPacketFromBytes(data []byte) (AsynTjmpPacket, error) {
	var packet AsynTjmpPacket
	remainingBytes, clockRef, err := ParseAsynHeader(data, TJMP)
	if err != nil {
		return packet, err
	}
	if len(remainingBytes) < 8+coremedia.CMTimeLengthInBytes*2 {
		return packet, fmt.Errorf("TJMP payload too short: %d bytes (need 56)", len(remainingBytes))
	}
	packet.ClockRef = clockRef
	packet.DeviceTimebaseID = binary.LittleEndian.Uint64(remainingBytes[0:8])

	packet.AnchorTime, err = coremedia.NewCMTimeFromBytes(remainingBytes[8 : 8+coremedia.CMTimeLengthInBytes])
	if err != nil {
		return packet, fmt.Errorf("parsing TJMP AnchorTime: %w", err)
	}
	off := 8 + coremedia.CMTimeLengthInBytes
	packet.CurrentTime, err = coremedia.NewCMTimeFromBytes(remainingBytes[off : off+coremedia.CMTimeLengthInBytes])
	if err != nil {
		return packet, fmt.Errorf("parsing TJMP CurrentTime: %w", err)
	}
	return packet, nil
}

func (sp AsynTjmpPacket) String() string {
	return fmt.Sprintf("ASYN_TJMP{ClockRef:%x, DeviceTimebaseID:%x, AnchorTime:%s, CurrentTime:%s}",
		sp.ClockRef, sp.DeviceTimebaseID, sp.AnchorTime.String(), sp.CurrentTime.String())
}
