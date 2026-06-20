package screencapture

import (
	"fmt"
	"time"

	"github.com/google/gousb"
	log "github.com/sirupsen/logrus"
)

// EnableQTConfig enables the hidden QuickTime Device configuration that will expose two new bulk endpoints.
// We will send a control transfer to the device via USB which will cause the device to disconnect and then
// re-connect with a new device configuration. Usually the usbmuxd will automatically enable that new config
// as it will detect it as the device's preferredConfig.
func EnableQTConfig(device IosDevice) (IosDevice, error) {
	usbSerial := device.SerialNumber
	ctx := gousb.NewContext()
	usbDevice, err := OpenDevice(ctx, device)
	if err != nil {
		return IosDevice{}, err
	}
	if isValidIosDeviceWithActiveQTConfig(usbDevice.Desc) {
		// Mode-2 coexistence (ALI / ideacatlab): usbmuxd (USBMUXD_DEFAULT_DEVICE_MODE=2)
		// already put the device on the Valeria config (config 5), which carries BOTH
		// the usbmux/control interface (0xFE) and the AV interface (0x2A) — so WDA
		// control coexists. We do NOT re-arm here: the device only PINGs right after a
		// re-arm, and at this point we are not yet reading the AV endpoint, so the PING
		// would be lost. The re-arm (disable+enable) is done in StartReading WHILE we
		// are listening, which makes the handshake reliable.
		log.Debugf("%s already on QT config (mode-2); AV session will be re-armed by the reader", usbSerial)
		return device, nil
	}

	sendQTConfigControlRequest(usbDevice)

	var i int
	for {
		log.Debugf("Checking for active QT config for %s", usbSerial)

		err = ctx.Close()
		if err != nil {
			log.Warn("failed closing context", err)
		}
		time.Sleep(500 * time.Millisecond)
		log.Debug("Reopening Context")
		ctx = gousb.NewContext()
		device, err = device.ReOpen(ctx)
		if err != nil {
			log.Debugf("device not found:%s", err)
			continue
		}
		i++
		if i > 10 {
			log.Debug("Failed activating config")
			return IosDevice{}, fmt.Errorf("could not activate Quicktime Config for %s", usbSerial)
		}
		break
	}
	log.Debugf("QTConfig for %s activated", usbSerial)
	return device, err
}

func DisableQTConfig(device IosDevice) (IosDevice, error) {
        usbSerial := device.SerialNumber
        ctx := gousb.NewContext()
        usbDevice, err := OpenDevice(ctx, device)
        if err != nil {
                return IosDevice{}, err
        }
        if !isValidIosDeviceWithActiveQTConfig(usbDevice.Desc) {
            log.Debugf("Skipping %s because it is already deactivated", usbSerial)
            return device, nil
        }

        confignum, _ := usbDevice.ActiveConfigNum()
        log.Debugf("Config is active: %d, QT config is: %d", confignum, device.QTConfigIndex)

        for i := 0; i < 20; i++{
            sendQTDisableConfigControlRequest(usbDevice)
            log.Debugf("Resetting device config (#%d)", i + 1)
            _, err := usbDevice.Config(device.UsbMuxConfigIndex)
            if err != nil {
                log.Warn(err)
            }
        }

        confignum, _ = usbDevice.ActiveConfigNum()
        log.Debugf("Config is active: %d, QT config is: %d", confignum, device.QTConfigIndex)


        return device, err
}

func sendQTConfigControlRequest(device *gousb.Device) {
	response := make([]byte, 0)
	val, err := device.Control(0x40, 0x52, 0x00, 0x02, response)
	if err != nil {
		log.Warnf("Failed sending control transfer for enabling hidden QT config. Seems like this happens sometimes but it still works usually: %s", err)
	}
	log.Debugf("Enabling QT config RC:%d", val)
}

func sendQTDisableConfigControlRequest(device *gousb.Device) {
	response := make([]byte, 0)
	val, err := device.Control(0x40, 0x52, 0x00, 0x00, response)

	if err != nil {
		log.Warnf("Failed sending control transfer for disabling hidden QT config:%s", err)

	}
	log.Debugf("Disabled QT config RC:%d", val)
}

// sendQTSetModeRequest sends Apple's vendor-specific SET_MODE (0x52) with wIndex=mode.
// This is the SAME command usbmuxd uses to switch Apple device modes
// (1 = initial/config-4, 2 = "Valeria"/config-5 AV). A proper mode-1 -> mode-2 transition
// is a REAL re-entry into Valeria and resets a stuck AV session, unlike the historical
// disable(0)+enable(2) re-cycle (wIndex 0 is not a valid mode, so it is not a true
// transition). Used by the flag-gated reset path in StartReading. See ideacatlab/usbmuxd
// FIX_PLAN_ALI.md — long term this reset belongs inside usbmuxd (it owns the device).
func sendQTSetModeRequest(device *gousb.Device, mode uint16) {
	val, err := device.Control(0x40, 0x52, 0x00, mode, make([]byte, 0))
	if err != nil {
		log.Warnf("Failed SET_MODE(%d): %s", mode, err)
	}
	log.Debugf("SET_MODE(%d) RC:%d", mode, val)
}
