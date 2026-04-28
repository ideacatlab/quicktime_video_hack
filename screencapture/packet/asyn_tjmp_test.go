package packet_test

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/danielpaulus/quicktime_video_hack/screencapture/coremedia"
	"github.com/danielpaulus/quicktime_video_hack/screencapture/packet"
	"github.com/stretchr/testify/assert"
)

func TestTjmp(t *testing.T) {
	dat, err := ioutil.ReadFile("fixtures/asyn-tjmp")
	if err != nil {
		log.Fatal(err)
	}
	tjmpPacket, err := packet.NewAsynTjmpPacketFromBytes(dat)
	if !assert.NoError(t, err) {
		return
	}

	// Per the ASYN header
	assert.Equal(t, uint64(0x11123bc18), tjmpPacket.ClockRef)

	// The 56-byte payload of this fixture is:
	//   [8 bytes: DeviceTimebaseID = 0]
	//   [24 bytes: AnchorTime  CMTime{ value:0, scale:1, flags:KCMTimeFlagsHasBeenRounded, epoch:0 }]
	//   [24 bytes: CurrentTime CMTime{ value:0, scale:1, flags:KCMTimeFlagsHasBeenRounded, epoch:0 }]
	assert.Equal(t, uint64(0), tjmpPacket.DeviceTimebaseID)

	expectedTime := coremedia.CMTime{
		CMTimeValue: 0,
		CMTimeScale: 1,
		CMTimeFlags: coremedia.KCMTimeFlagsHasBeenRounded,
		CMTimeEpoch: 0,
	}
	assert.Equal(t, expectedTime, tjmpPacket.AnchorTime)
	assert.Equal(t, expectedTime, tjmpPacket.CurrentTime)
}
