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
		// Mode-2 coexistence (ALI / ideacatlab): when usbmuxd runs with
		// USBMUXD_DEFAULT_DEVICE_MODE=2 it puts the device on the Valeria config
		// (config 5), where the usbmux/control interface (subclass 0xFE) and the AV
		// interface (subclass 0x2A) live in the SAME config — so WDA/Appium control
		// keeps working over usbmux while we stream H.264 over the AV interface.
		// The device is already on config 5, but its AV session was never (re)armed
		// for us, so it never sends PING. Re-cycle the QT config via disable+enable
		// control transfers to re-arm the AV session WITHOUT a full re-enumeration,
		// so control stays attached. (Verified on iPhone/iOS 16.7: PING + full audio
		// handshake while the device remained controlled by WDA.)
		log.Debugf("%s already on QT config (mode-2); re-cycling AV session via disable+enable", usbSerial)
		sendQTDisableConfigControlRequest(usbDevice)
		sendQTConfigControlRequest(usbDevice)
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
