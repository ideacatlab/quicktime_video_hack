package screencapture

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"

	"github.com/google/gousb"
	log "github.com/sirupsen/logrus"
)

//UsbAdapter reads and writes from AV Quicktime USB Bulk endpoints
type UsbAdapter struct {
	outEndpoint   *gousb.OutEndpoint
	Dump          bool
	DumpOutWriter io.Writer
	DumpInWriter  io.Writer
}

//WriteDataToUsb implements the UsbWriter interface and sends the byte array to the usb bulk endpoint.
func (usbAdapter *UsbAdapter) WriteDataToUsb(bytes []byte) {
	_, err := usbAdapter.outEndpoint.Write(bytes)
	if err != nil {
		log.Error("failed sending to usb", err)
	}
	if usbAdapter.Dump {
		_, err := usbAdapter.DumpOutWriter.Write(bytes)
		if err != nil {
			log.Fatalf("Failed dumping data:%v", err)
		}
	}
}

//StartReading claims the AV Quicktime USB Bulk endpoints and starts reading until a stopSignal is sent.
//Every received data is added to a frameextractor and when it is complete, sent to the UsbDataReceiver.
func (usbAdapter *UsbAdapter) StartReading(device IosDevice, receiver UsbDataReceiver, stopSignal chan interface{}) error {
	ctx, cleanUp := createContext()
	defer cleanUp()

	usbDevice, err := OpenDevice(ctx, device)
	if err != nil {
		return err
	}
	if !device.IsActivated() {
		return errors.New("device not activated for screen mirroring")
	}

	confignum, _ := usbDevice.ActiveConfigNum()
	log.Debugf("Config is active: %d, QT config is: %d", confignum, device.QTConfigIndex)

	config, err := usbDevice.Config(device.QTConfigIndex)
	if err != nil {
		return errors.New("Could not retrieve config")
	}

	log.Debugf("QT Config is active: %s", config.String())

	iface, err := findAndClaimQuickTimeInterface(config)
	if err != nil {
		log.Debug("could not get Quicktime Interface")
		return err
	}
	log.Debugf("Got QT iface:%s", iface.String())

	inboundBulkEndpointIndex, inboundBulkEndpointAddress, err := findBulkEndpoint(iface.Setting, gousb.EndpointDirectionIn)
	if err != nil {
		return err
	}

	outboundBulkEndpointIndex, outboundBulkEndpointAddress, err := findBulkEndpoint(iface.Setting, gousb.EndpointDirectionOut)
	if err != nil {
		return err
	}

	err = clearFeature(usbDevice, inboundBulkEndpointAddress, outboundBulkEndpointAddress)
	if err != nil {
		return err
	}

	inEndpoint, err := iface.InEndpoint(inboundBulkEndpointIndex)
	if err != nil {
		log.Error("couldnt get InEndpoint")
		return err
	}
	log.Debugf("Inbound Bulk: %s", inEndpoint.String())

	outEndpoint, err := iface.OutEndpoint(outboundBulkEndpointIndex)
	if err != nil {
		log.Error("couldnt get OutEndpoint")
		return err
	}
	log.Debugf("Outbound Bulk: %s", outEndpoint.String())
	usbAdapter.outEndpoint = outEndpoint

	stream, err := inEndpoint.NewStream(4096, 5)
	if err != nil {
		log.Fatal("couldnt create stream")
		return err
	}
	log.Debug("Endpoint claimed")
	log.Infof("Device '%s' USB connection ready, waiting for ping..", device.SerialNumber)
	var gotData int32
	go func() {
		lengthBuffer := make([]byte, 4)
		for {
			n, err := io.ReadFull(stream, lengthBuffer)
			if err != nil {
				log.Errorf("Failed reading 4bytes length with err:%s only received: %d", err, n)
				return
			}
			atomic.StoreInt32(&gotData, 1)
			//the 4 bytes header are included in the length, so we need to subtract them
			//here to know how long the payload will be
			length := binary.LittleEndian.Uint32(lengthBuffer) - 4
			dataBuffer := make([]byte, length)

			n, err = io.ReadFull(stream, dataBuffer)
			if err != nil {
				log.Errorf("Failed reading payload with err:%s only received: %d/%d bytes", err, n, length)
				var signal interface{}
				stopSignal <- signal
				return
			}
			if usbAdapter.Dump {
				_, err := usbAdapter.DumpInWriter.Write(dataBuffer)
				if err != nil {
					log.Fatalf("Failed dumping data:%v", err)
				}
			}
			receiver.ReceiveData(dataBuffer)
		}
	}()

	// AV arming (ideacatlab). The device emits PING/CWPA right after it enters the Valeria
	// config (config 5). Two strategies:
	//
	//   - JUST-LISTEN (default): the reader above simply listens and catches the PING/CWPA that
	//     is pending from the config-5 transition usbmuxd already performed (mode-2). This is the
	//     correct path on the ALI fleet, where usbmuxd OWNS config 5 (USBMUXD_DEFAULT_DEVICE_MODE=2):
	//     issuing our own SET_MODE 0x52 here just FIGHTS usbmuxd for the device mode (proven to
	//     wedge usbmuxd blind → whole-fleet outage) AND does not actually re-arm a settled device
	//     (SET_MODE is ignored once config 5 is claimed — verified). So by default we do NOT
	//     re-cycle: no contention.
	//
	//   - RE-CYCLE (QVH_FORCE_RECYCLE=1): the legacy behavior — re-cycle the QT config (SET_MODE
	//     disable+enable) repeatedly while reading. Only useful when qvh, not usbmuxd, drives the
	//     device mode (no mode-2 usbmuxd). Kept as an escape hatch.
	if os.Getenv("QVH_FORCE_RECYCLE") == "1" {
		go func() {
			for r := 0; r < 25 && atomic.LoadInt32(&gotData) == 0; r++ {
				time.Sleep(1200 * time.Millisecond)
				if atomic.LoadInt32(&gotData) != 0 {
					return
				}
				log.Debugf("no AV data yet (retry %d); re-cycling QT config to re-arm PING", r)
				sendQTDisableConfigControlRequest(usbDevice)
				sendQTConfigControlRequest(usbDevice)
			}
		}()
	} else {
		log.Debug("QVH just-listen mode (no SET_MODE re-cycle) — usbmuxd owns config 5; catching the pending CWPA")
	}

	<-stopSignal
	receiver.CloseSession()
	log.Info("Closing usb stream")

	err = stream.Close()
	if err != nil {
		log.Error("Error closing stream", err)
	}
	log.Info("Closing usb interface")
	iface.Close()

	// On teardown, only drive the device mode back to usbmux when WE own it (legacy re-cycle
	// mode). On the mode-2 fleet usbmuxd owns config 5: issuing SET_MODE here fights it
	// (fails busy[-6]) and would drift the device toward config 4, breaking the next arm. So in
	// just-listen mode we leave the device on config 5 and let usbmuxd keep owning it.
	if os.Getenv("QVH_FORCE_RECYCLE") == "1" {
		sendQTDisableConfigControlRequest(usbDevice)
		log.Debug("Resetting device config")
		_, err = usbDevice.Config(device.UsbMuxConfigIndex)
		if err != nil {
			log.Warn(err)
		}
	}

	return nil
}

func clearFeature(usbDevice *gousb.Device, inboundBulkEndpointAddress gousb.EndpointAddress, outboundBulkEndpointAddress gousb.EndpointAddress) error {
	val, err := usbDevice.Control(0x02, 0x01, 0, uint16(inboundBulkEndpointAddress), make([]byte, 0))
	if err != nil {
		return errors.Wrap(err, "clear feature failed")
	}
	log.Debugf("Clear Feature RC: %d", val)

	val, err = usbDevice.Control(0x02, 0x01, 0, uint16(outboundBulkEndpointAddress), make([]byte, 0))
	log.Debugf("Clear Feature RC: %d", val)
	return errors.Wrap(err, "clear feature failed")
}

func findBulkEndpoint(setting gousb.InterfaceSetting, direction gousb.EndpointDirection) (int, gousb.EndpointAddress, error) {
	for _, v := range setting.Endpoints {
		if v.Direction == direction {
			return v.Number, v.Address, nil

		}
	}
	return 0, 0, errors.New("Inbound Bulkendpoint not found")
}

func findAndClaimQuickTimeInterface(config *gousb.Config) (*gousb.Interface, error) {
	log.Debug("Looking for quicktime interface..")
	found, ifaceIndex := findInterfaceForSubclass(config.Desc, QuicktimeSubclass)
	if !found {
		return nil, fmt.Errorf("did not find interface %v", config)
	}
	log.Debugf("Found Quicktimeinterface: %d", ifaceIndex)
	return config.Interface(ifaceIndex, 0)
}
