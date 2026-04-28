package packet

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Responder-format ping wire bytes (qvh's role): length=16, magic="ping",
// body = u64 LE 0x0000000100000000 → bytes 00 00 00 00 01 00 00 00.
// Empirically confirmed required against iOS 18; see the doc comment in
// ping.go for the contrasting initiator layout.
const pingPacketHexDump = "10000000676e69700000000001000000"

func TestPingSerialization(t *testing.T) {
	assert.Equal(t, pingPacketHexDump, fmt.Sprintf("%x", NewPingPacketAsBytes()))
}
