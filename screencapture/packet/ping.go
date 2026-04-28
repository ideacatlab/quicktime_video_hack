package packet

import "encoding/binary"

// Constants for creating a Ping packet.
//
// The ping packet is exactly 16 bytes on the wire. There are TWO layouts,
// and which one to use depends on whether you are the connection
// initiator or the responder:
//
//   Initiator (Apple's _usb_clientSendStartupPing in CoreMedia.framework
//   on macOS 15.5; sends first after USB activation):
//     bytes 0-3:   0x00000010                  (length, LE u32)
//     bytes 4-7:   "ping" 0x70696E67           (magic, LE u32)
//     bytes 8-11:  0x00000001                  (constant 1, LE u32)
//     bytes 12-15: 32-bit value from FigTransportConnectionUSB
//                  derived-storage[0x1c]. Read once per session, not
//                  incremented per ping (no write site visible in any
//                  usb_* function, and usb_pingAsyncCallback is just
//                  the async completion callback — no counter bump).
//                  Either initialized to 0 by the CMBase object
//                  constructor or set to a session-unique nonce
//                  somewhere we haven't traced. Apple sends exactly
//                  one ping per session, same as qvh, so the value
//                  effectively stays constant in either case.
//
//   Responder (what qvh is — the device sends a ping first; qvh echoes
//   back this format):
//     bytes 0-3:   0x00000010                  (length)
//     bytes 4-7:   "ping" 0x70696E67           (magic)
//     bytes 8-11:  0x00000000                  (zero)
//     bytes 12-15: 0x00000001                  (constant 1)
//
// The two layouts are NOT interchangeable. Empirical test on iOS 18:
// using the initiator layout in qvh's responder role causes the device
// to never complete its handshake (no FEED packets arrive, recording
// produces a 0-byte file). So qvh's responder layout below is correct;
// do NOT swap to Apple's startup-ping layout even though it looks
// "more like Apple's code".
const (
	PingPacketMagic uint32 = 0x70696E67
	PingLength      uint32 = 16
	// PingHeader is the 8-byte body of the responder-format ping:
	// LE u64 0x0000000100000000 → bytes "00 00 00 00 01 00 00 00".
	PingHeader uint64 = 0x0000000100000000
)

//NewPingPacketAsBytes generates a new default Ping packet
func NewPingPacketAsBytes() []byte {
	packetBytes := make([]byte, 16)
	binary.LittleEndian.PutUint32(packetBytes, PingLength)
	binary.LittleEndian.PutUint32(packetBytes[4:], PingPacketMagic)
	binary.LittleEndian.PutUint64(packetBytes[8:], PingHeader)
	return packetBytes
}
