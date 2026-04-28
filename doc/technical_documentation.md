# Table of Contents
- [Technical Documentation of the iOS ScreenSharing Feature](#technical-documentation-of-the-ios-screensharing-feature)
  * [0. General Information](#0-general-information)
  * [1. How to Enable it for a iOS Device on the USB Level](#1-how-to-enable-it-for-a-ios-device-on-the-usb-level)
    + [1.1 Foundations](#11-foundations)
    + [1.2 Finding Devices and Configurations using LibUsb](#12-finding-devices-and-configurations-using-libusb)
    + [1.3 Hidden Configuration](#13-hidden-configuration)
    + [1.4 Enabling the Hidden Config](#14-enabling-the-hidden-config)
  * [2. AV Session LifeCycle](#2-av-session-lifecycle)
    + [2.1 Initiate the session](#21-initiate-the-session)
    + [2.2 Receive data](#22-receive-data)
    + [2.3 Shutting down streaming](#23-shutting-down-streaming)
  * [3. Protocol Reference](#3-protocol-reference)
    + [3.1 Ping Packet](#31-ping-packet)
    + [3.2 Sync Packets](#32-sync-packets)
      - [3.2.1 General Description](#321-general-description)
      - [3.2.2. CWPA Packet and Response](#322-cwpa-packet-and-response)
        * [General Description](#general-description)
        * [Request Format Description](#request-format-description)
        * [Reply - RPLY Format Description](#reply---rply-format-description)
      - [3.2.3. AFMT Packet](#323-afmt-packet)
        * [General Description](#general-description-1)
        * [Request Format Description](#request-format-description-1)
        * [Reply - RPLY Format Description](#reply---rply-format-description-1)
      - [3.2.4. CVRP Packet](#324-cvrp-packet)
        * [General Description](#general-description-2)
        * [Request Format Description](#request-format-description-2)
        * [Reply - RPLY Format Description](#reply---rply-format-description-2)
      - [3.2.5. CLOK Packet](#325-clok-packet)
        * [General Description](#general-description-3)
        * [Request Format Description](#request-format-description-3)
        * [Reply - RPLY Format Description](#reply---rply-format-description-3)
      - [3.2.6. TIME Packet](#326-time-packet)
        * [General Description](#general-description-4)
        * [Request Format Description](#request-format-description-4)
        * [Reply - RPLY Format Description](#reply---rply-format-description-4)
      - [3.2.7. SKEW Packet](#327-skew-packet)
        * [General Description](#general-description-5)
        * [Request Format Description](#request-format-description-5)
        * [Reply - RPLY Format Description](#reply---rply-format-description-5)
      - [3.2.8. OG Packet](#328-og-packet)
        * [General Description](#general-description-6)
        * [Request Format Description](#request-format-description-6)
        * [Reply - RPLY Format Description](#reply---rply-format-description-6)
      - [3.2.9. STOP Packet](#329-stop-packet)
        * [General Description](#general-description-7)
        * [Request Format Description](#request-format-description-7)
        * [Reply - RPLY Format Description](#reply---rply-format-description-7)
    + [3.3 Asyn Packets](#33-asyn-packets)
        * [3.3.0 General Description](#330-general-description)
        * [3.3.1. Asyn SPRP - Set Property](#331-asyn-sprp---set-property)
          + [General Description](#general-description-8)
          + [Packet Format Description](#packet-format-description)
        * [3.3.2. Asyn SRAT - Set time rate and Anchor](#332-asyn-srat---set-time-rate-and-anchor)
          + [General Description](#general-description-9)
          + [Packet Format Description](#packet-format-description-1)
        * [3.3.3. Asyn TBAS - Set TimeBase](#333-asyn-tbas---set-timebase)
          + [General Description](#general-description-10)
          + [Packet Format Description](#packet-format-description-2)
        * [3.3.4. Asyn TJMP - Time Jump Notification](#334-asyn-tjmp---time-jump-notification)
          + [General Description](#general-description-11)
          + [Packet Format Description](#packet-format-description-3)
        * [3.3.5 Asyn FEED - CMSampleBuffer with h264 Video Data](#335-asyn-feed---cmsamplebuffer-with-h264-video-data)
          + [Packet Format Description](#packet-format-description-4)
        * [3.3.6 Asyn EAT! - CMSampleBuffer with Audio Data](#336-asyn-eat----cmsamplebuffer-with-audio-data)
        * [3.3.7 Asyn NEED - Tell the device to send more](#337-asyn-need---tell-the-device-to-send-more)
          + [Packet Format Description](#packet-format-description-5)
        * [3.3.8 Asyn HPD0 - Tell the device to stop video streaming](#338-asyn-hpd0---tell-the-device-to-stop-video-streaming)
          + [Packet Format Description](#packet-format-description-6)
        * [3.3.9 Asyn HPA0 - Tell the device to stop audio streaming](#339-asyn-hpa0---tell-the-device-to-stop-audio-streaming)
          + [Packet Format Description](#packet-format-description-7)
        * [3.3.10 Asyn RELS - Tell us about a released Clock on the device](#3310-asyn-rels---tell-us-about-a-released-clock-on-the-device)
          + [Packet Format Description](#packet-format-description-8)
  * [4 Serializing/Deserializing Objects](#4-serializing-deserializing-objects)
    + [4.0 General Description](#40-general-description)
    + [4.1 Dictionaries](#41-dictionaries)
      - [4.1.0 General Description](#410-general-description)
      - [4.1.1 Dictionaries with String Keys](#411-dictionaries-with-string-keys)
        * [4.1.1.1 General Dictionary Structure](#4111-general-dictionary-structure)
      - [4.1.2 Dictionaries with 4-Byte Integer Index Keys](#412-dictionaries-with-4-byte-integer-index-keys)
    + [4.2 CMTime](#42-cmtime)
      - [Example](#example)
    + [4.3 CMSampleBuffer](#43-cmsamplebuffer)
    + [4.4 NSNumber](#44-nsnumber)
      - [Example](#example-1)
    + [4.5 CMFormatDescription](#45-cmformatdescription)
  * [5. Clocks and CMSync](#5-clocks-and-cmsync)

<small><i><a href='http://ecotrust-canada.github.io/markdown-toc/'>Table of contents generated with markdown-toc</a></i></small>

# Technical Documentation of the iOS ScreenSharing Feature
## 0. General Information
This document provides you with details about the screen sharing feature of QuickTime for iOS devices. 
The information contained in this document can be used to re-implement that feature in the programming language of choice and use
the feature on other operating systems than MAC OS X. If you want to implement the feature, I recommend using my unit tests and test fixtures. I have prepared a bin dump with an example for every message type. This way you can easily make sure your codec does what it's supposed to.
The repository also contains a reference implementation in Golang. 

-- Note: All the information in this document is reverse engineered by the author, therefore it could be wrong or not entirely accurate
as it involves a lot of assumptions. If you find mistakes or more accurate descriptions please add them :-) -- 


## 1. How to Enable it for a iOS Device on the USB Level
### 1.1 Foundations
Usually devices attached on the USB Port have a set of "configurations" that you can retrieve using the LibUsb wrapper you use.
Inside of these interfaces there are a set of Usb Endpoints you can use to communicate with your device. We are interested in the 
`bulk` endpoints of iOS devices as these are used for transferring data. 
### 1.2 Finding Devices and Configurations using LibUsb
Look for a USB ConfigDescriptor that has an interface with Class `DeviceClass = 0xFF` 
By default, iOS devices only have the USBMux ConfigDescriptor which has SubClass `0xFE` and will contain one interface with two Bulk endpoints.
The Subclass for a config that allows AV streaming will be  `0x2A` and there will be an interface with 4 Bulk endpoints. It needs to be enabled first.

### 1.3 Hidden Configuration
To use video mirroring, you have to enable the hidden Quicktime Configuration (I just call it QT-Config, not an official apple name)
If you closely monitor all available USB configurations you will find that if a device has n-Configurations, as soon as you open QuickTime on a mac
and start recording the iOS device's screen, there will be n+Configurations
### 1.4 Enabling the Hidden Config
To enable the hidden QTconfig you have to send a specific Control Request to the device like so:
`val, err := device.usbDevice.Control(0x40, 0x52, 0x00, 0x02, response)`
If you did it correctly, it will cause the device to disconnect from the host machine and re-connect after a few moments with an additional config.
The new config contains 4 bulk endpoints. 2 For communication with the usbmuxd on the device, and two additional endpoints for receiving and sending AV data.
Call setActiveConfiguration on that config and you can claim the new endpoint for sending and receiving AV data.
## 2. AV Session LifeCycle

### 2.1 Initiate the session
1. enable hidden device config
2. claim endpoint
3. wait to receive a PING packet
4. respond with a PING packet
5. wait for SYNC CWPA packet to receive the clockref for the devices audio clock
6. create local clock, put clockref in reply to the SYNC CWPA and send
7. send ASYN_HPD1
8. send ASYN_HPA1 with the device audio clockref received in step 6
9. receive SYNC AFMT and reply with a zero error code
10. receive SYNC CVRP with the devices video clockRef
11. reply with the local video clockRef
12. start sending ASYN NEED with the device's video clockRef
13. receive two ASYN Set Properties
14. receive Sync Clok and reply with newly created clock
15. receive two SYNC TIME and reply with two CMTimes 
### 2.2 Receive data
FEED and EAT! Packets for video and audio will be sent by the device.
We need to send NEED packets for video periodically

### 2.3 Shutting down streaming
1. send asyn hpa0 with the deviceclockref from the cwpa sync packet to tell the device to stop sending audio
2. send hpd0 with empty clockRef to stop video
3. receive sync stop package for our video clock we created when cvrp was sent to us and that is in every feed packet
4. reply to sync stop with 8 zero bytes
5. receive a ASYN RELS for the local video clockRef (the one found in FEED packets)
6. receive a ASYN RELS for the local clock created after the SYNC CLOCK 
7. release usb endpoint
8. set the device active config to usbmux only


## 3. Protocol Reference
### 3.1 Ping Packet
As soon as we connect to the USB endpoint, we need to wait for the device to send us a ping packet. Once we received it, we will send a ping back to the device and 
then progress to the rest of the communication. Example Ping:

| 4 Byte Length (16)   |4 Byte Magic (PING)   | 4byte 0x0| 4 byte 0x1   | 
|---|---|---|---|
|10000000 |676E6970 | 00000000 | 01000000 |
  

### 3.2 Sync Packets
#### 3.2.1 General Description
All SYNC packets require us to reply with a RPLY packet. 
It seems like this is mostly used for synchronizing CMClocks and exchanging 8byte CMClockRefs (implement CMSync.h protocol)
Usually you can see that SYNC packets have a 4byte SUB-TYPE followed by a 8byte correlationID. A reply always contains the correlationID so 
I assume this is how the device knows which reply belongs to which request.


#### 3.2.2. CWPA Packet and Response
##### General Description
This packet seems to be used for intitiating the audio stream. We get a clockRef from the device and respond with our own, newly created clockRef.
The clockref send by the device needs to go in the ASYN-1APH packet we send.
##### Request Format Description

| 4 Byte Length (36)   |4 Byte Magic (SYNC)   | 8 Empty clock reference| 4 byte message type (CWPA)   | 8 byte correlation id  | 8 bytes CFTypeID of the device clock |
|---|---|---|---|---|---|
|24000000 |636E7973 |01000000 00000000 | 61707763 |E03D5713 01000000| E0740000 5A130040 |

##### Reply - RPLY Format Description

Sends back our clockRef. The device will use the clockRef from here in the SYNC_AFMT message to tell us about the audio format. 
Also this will be used for all ASYN_EAT packets containing audio sample buffers. 

| 4 Byte Length (28)   |4 Byte Magic (RPLY)   | 8 Byte correlation id  |   4 Byte: 0  | 8 bytes CFTypeID of our clock |
|---|---|---|---|---|
|1C000000 | 796C7072 |E03D5713 01000000 | 00000000 |B00CE26C A67F0000|

#### 3.2.3. AFMT Packet
##### General Description
This packet contains information about the Audio Format(AMFT). It contains a AudioStreamBasicDescription struct [see here](https://github.com/nu774/MSResampler/blob/master/CoreAudio/CoreAudioTypes.h). It is usually of MediaID Linear pulse-code modulation (LPCM), which means uncompressed audio.
The response is basically a dictionary containing an error code. Normally we send 0 to indicate everything is ok.
Note how the device references the Clock we gave it in the SYNC_CWPA_RPLY
##### Request Format Description

| 4 Byte Length (68)   |4 Byte Magic (SYNC)   | 8 bytes clock CFTypeID| 4 byte magic (AFMT)| 8 byte correlation id| AudioStreamBasicDescription struct: float64 sampling frequency (48khz), data 4 byte magic (LPCM),   28 bytes rest|
|---|---|---|---|---|---|
|44000000| 636E7973| B00CE26C A67F0000| 746D6661 | 809D2213 01000000| 00000000 0070E740 6D63706C 4C000000 04000000 01000000 04000000 02000000 10000000 00000000|

##### Reply - RPLY Format Description
Contains the correlationID from the request as well as a simple Dictionary:  {"Error":NSNumberUint32(0)}

| 4 Byte Length (62)   |4 Byte Magic (RPLY)   | 8  correlation id| 4 byte 0| 4 byte dict length(42)| 4 byte magic (DICT)| dict bytes |
|---|---|---|---|---|---|---|
|3E000000 |796C7072| 809D2213 01000000 |00000000| 2A000000| 74636964| 22000000 7679656B 0D000000 6B727473 4572726F 720D0000 0076626D 6E030000 0000|

#### 3.2.4. CVRP Packet
##### General Description
The CVRP packet is used to create a local clock for synchronizing video frames. The ClockRef we get from the device, is the one we need to put in NEED packets.
Similar to this the device sends all ASYN_FEED sample buffer with a reference to our clock. After this, two SetProperty Asyns are usually received plus CLok, TBAS and SRAT.
##### Request Format Description

Contains a Dict with a FormatDescription and timing information. With the FormatDescription for the first time we get the h264 Picture and Sequence ParameterSet(PPS/SPS) already encoded in nice NALUs ready for streaming
over the network. They are hidden inside a dictionary inside the extension dictionary.

|4 Byte Length (649)|4 Byte Magic (SYNC)|8 byte empty(?) clock reference|4 byte magic(CVRP)|8 byte correlation id|CFTypeID of clock on device (needs to be in NEED packets we send)|4 byte length of dictionary (613)|4 byte magic (DICT)| Dict bytes|
|---|---|---|---|---|---|---|---|---|
|89020000 |636E7973| 01000000 00000000 |70727663| D0595613 01000000 |A08D5313 01000000 |65020000| 74636964|   0x.....|

##### Reply - RPLY Format Description

| 4 Byte Length (28)   |4 Byte Magic (RPLY)   | 8 Byte correlation id  |  4 bytes (seem to be always 0) | 8 bytes CFTypeID of our clock(will be in all feed async packets) |
|---|---|---|---|---|
|1C000000 | 796C7072 |D0595613 01000000 | 00000000 |5002D16C A67F0000 |


#### 3.2.5. CLOK Packet
##### General Description
This requests us to create a new clock and send back the clockRef. It is usually followed by 2 TIME requests.

##### Request Format Description

| 4 Byte Length (28)   |4 Byte Magic (SYNC)   | 8 Byte clock CFTypeID  |  4 bytes magic (CLOK) | 8 bytes correlation id |
|---|---|---|---|---|
|1C000000| 636E7973| 5002D16C A67F0000| 6B6F6C63 | 70495813 01000000 |

##### Reply - RPLY Format Description

| 4 Byte Length (28)   |4 Byte Magic (RPLY)   | 8 correlation id  |  4 bytes (seem to be always 0) | 8 bytes CFTypeID of our clock(for the next two time packets) |
|---|---|---|---|---|
|1C000000| 796C7072| 70495813 01000000| 00000000 | 8079C17C A67F0000|

#### 3.2.6. TIME Packet
##### General Description
This packet requests from us to send a RPLY with the current CMTime for the ClockRef specified.
##### Request Format Description

| 4 Byte Length (28)   |4 Byte Magic (SYNC)   | 8 Byte clock CFTypeID  |  4 bytes magic (TIME) | 8 bytes correlation id |
|---|---|---|---|---|
|1C000000| 636E7973| 8079C17C A67F0000 |656D6974 | 503D2213 01000000 |

##### Reply - RPLY Format Description

| 4 Byte Length (44)   |4 Byte Magic (RPLY)   | 8 Byte correllation id  |  4 bytes 0x0 | 24 bytes CMTime struct |
|---|---|---|---|---|
|2C000000 |796C7072 |503D2213 01000000| 00000000 | E1E142C4 62BA0000 00CA9A3B 01000000 00000000 00000000|

#### 3.2.7. SKEW Packet
##### General Description
This packet tells the device about the clock skew of the audio clock (clockRef used in EAT! packets, which we sent as response to cwpa). As denoted in this [wikipedia](https://en.wikipedia.org/wiki/Clock_skew#On_a_network) article, clock skew means the difference in frequency of both clocks. In other words, both clocks supposedly 
run at 48khz, and the device wants to know how many ticks per second our clock executed during the time the device clock had one tick. 
So we have to respond with:
- 48000 if the clocks were aligned
- some value above 48000 if our clock was slower
- and some value below 48000 if our clock was faster than the device clock
If implemented correctly, we should see that the skew responses converge towards 48000 with small deviations sometimes `(48000+x where -1 < x <1)`

##### Request Format Description

| 4 Byte Length (28)   |4 Byte Magic (SYNC)   | 8 Byte clock CFTypeID  |  4 bytes magic (SKEW) | 8 bytes correlation id | 
|---|---|---|---|---|
|20000000| 636E7973| 8079C17C A67F0000 |77656B73 | 60B9FD02 01000000 | 

##### Reply - RPLY Format Description

| 4 Byte Length (24)   |4 Byte Magic (RPLY)   | 8 Byte correllation id  | 4bytes padding 0x0| 8 bytes floating point number (48000.0) | 
|---|---|---|---|---|
|18000000 |796C7072 |60B9FD02 01000000| 00000000 | 00000000 0070E740 |

#### 3.2.8. OG Packet
##### General Description
**Magic constant:** `OG = 0x676F2120` (defined in
`screencapture/packet/sync.go`). On the wire, this serializes
little-endian to the bytes `20 21 6F 67` — read left-to-right that's
`" !og"`, which is where the section name "OG" comes from (the readable
suffix after the leading space byte).

**Status: still sent by iOS, but unhandled by Apple's modern macOS host stack. qvh's 8-zero-byte reply is what keeps the session alive.**

Empirical (iOS 18, macOS 15.5):

- **The device still sends OG.** Running qvh against a current iPhone for
  ~12 seconds shows exactly one inbound `SYNC_OG{Unknown:1}` arriving
  during initial session setup. So the packet is not deprecated on the
  device side; it remains part of the handshake.
- **Apple's macOS host has no OG handler anywhere.** The wire bytes
  (FourCC `0x676F2120`, either byte ordering) appear **nowhere** in
  `MediaToolbox.framework`, `CoreMedia.framework`,
  `iOSScreenCaptureAssistant`, the `iOSScreenCapture` plugin, or
  anywhere else in the macOS Sequoia (15.5)
  `dyld_shared_cache_arm64e` (1.6 GB+ of arm64e code across both cache
  files). The only byte-pattern hit is in
  `HealthDaemon`'s log message *"Here we go! Delegate asked to perform
  work!"* — completely unrelated.
  This was verified two ways: (1) byte-pattern search across all
  extracted Mach-Os and the dyld cache; (2) disassembly-based scan that
  reconstructs 32-bit immediates from arm64 `movz`/`movk` pairs (per
  the methodology note in section 7).
- **What Apple's host does on receipt:** the inbound packet dispatcher
  in `MediaToolbox.framework` (the giant function near
  `_FigNeroTeardown` on macOS 15.5) is a binary-tree FourCC dispatch.
  Any FourCC that doesn't match a known case falls through to the
  function's default return path — i.e., the OG packet is silently
  dropped without acknowledgement.
- **What qvh does on receipt** (in `screencapture/packet/sync_og.go`):
  parses the 4-byte payload (`Unknown:1`), constructs a 24-byte
  `RPLY` with the correlation ID and 8 zero bytes of body, and sends
  it back. The device evidently expects *some* reply (otherwise the
  session would stall waiting for it), but the content is irrelevant
  — 8 zeros is what qvh has always sent and the device is happy.

So OG is essentially a **vestigial handshake step**: the device asks,
the host has to answer, the host's answer is ignored, and the session
continues. Apple's QuickTime Player presumably either still has OG
support somewhere we haven't found or has deprecated it (with the
device tolerating the dropped reply). qvh's existing handling is
correct and necessary — do not remove it.

##### Request Format Description (historical, from qvh's original capture)

| 4 Byte Length (28)   |4 Byte Magic (SYNC)   | 8 Byte clock CFTypeID  |  4 bytes magic (OG!.) | 8 bytes correlation id | 4byte unknown int|
|---|---|---|---|---|---|
|20000000| 636E7973| 8079C17C A67F0000 |20216F67 | 302FD302 01000000 | 01000000 |

##### Reply - RPLY Format Description

| 4 Byte Length (24)   |4 Byte Magic (RPLY)   | 8 Byte correllation id  |  8 bytes 0x0 | 
|---|---|---|---|
|18000000 |796C7072 |302FD302 01000000| 00000000 00000000|

#### 3.2.9. STOP Packet
##### General Description
This one tells us to stop our clock
##### Request Format Description

| 4 Byte Length (28)   |4 Byte Magic (SYNC)   | 8 Byte clock CFTypeID  |  4 bytes magic (STOP) | 8 bytes correlation id | 
|---|---|---|---|---|
|1C000000| 636E7973| F05F4235 BA7F0000 | 706F7473 | 1049FD02 01000000 | 


##### Reply - RPLY Format Description

| 4 Byte Length (24)   |4 Byte Magic (RPLY)   | 8 Byte correllation id  |  4 bytes 0x0 | 4 bytes 0x0 |
|---|---|---|---|---|
|18000000 |796C7072 |1049FD02 01000000| 00000000 | 00000000|

### 3.3 Asyn Packets
##### 3.3.0 General Description
Asyn packets contain information like CMFormatDescriptions, Properties and most importantly the CMSampleBuffers that contain audio and video data.
They start with 4 byte length, 4 byte magic and then a 8 byte ClockRef.
ASYN packets do not require us to respond or ack.
##### 3.3.1. Asyn SPRP - Set Property
###### General Description
This packet is used to set properties for the video stream. Usually you get only two of them containing a pair each referencing the video stream.
1. ObeyEmptyMediaMarkers = true
2. RenderEmptyMedia = false

###### Packet Format Description

| 4 Byte Length (varies)   |4 Byte Magic (ASYN)   | 8 Byte clock CFTypeID  |  4 bytes magic (SPRP) | Key value pair bytes|
|---|---|---|---|---|
|20000000| 6E797361| 18BC2311 01000000 |70727073 | 00..... |


##### 3.3.2. Asyn SRAT - Set Rate and Anchor
###### General Description
**Confirmed: this is a `CMTimebaseSetRateAndAnchorTime` invocation from the device.**

Reverse-engineering of Apple's host-side handler in `MediaToolbox.framework`
(inside the inbound packet dispatcher near the `_FigNeroTeardown` exported
symbol; the SRAT comparison sits at vaddr `0x191583444`) confirms the
guess in the original document. The handler:
1. Reads 4 bytes at payload offset 0 → a 32-bit value (interpreted by
   the host as the new playback rate; commonly the IEEE-754 float `1.0`,
   i.e., bytes `00 00 80 3F`).
2. Reads 4 bytes at offset 4 → a second 32-bit value (likely a flags
   word; observed identical to the rate, but Apple's code treats it
   independently).
3. Reads a 24-byte `CMTime` at offset 8 → the anchor time.
4. Looks up the timebase via `CMAudioFormatDescriptionGetFormatList`
   (function pointer mislabeled by the disassembler due to PAC; the
   actual call is into a CoreMedia timebase API, equivalent to
   [`CMTimebaseSetRateAndAnchorTime`](https://developer.apple.com/documentation/coremedia/1641894-cmtimebasesetrateandanchortime)).

So the device is telling the host: "set the timebase rate to X with
anchor at this CMTime." qvh's send-side / response-side handling
(which doesn't reply, since this is asynchronous) is correct.

###### Packet Format Description


| 4 Byte Length (49)   |4 Byte Magic (ASYN)   | 8 Byte clock CFTypeID  |  4 bytes magic (SRAT) | 32bit float (1.0)| 32bit float (1.0)| 24 byte CMTime|
|---|---|---|---|---|---|---|
|31000000| 6E797361| 18BC2311 01000000 |74617273 | 0000803F | 0000803F| CB448EA1 CF10CC15 00CA9A3B 01000000 00000000 00000000|

##### 3.3.3. Asyn TBAS - Set TimeBase
###### General Description
**Confirmed: the device informs the host of a CMTimebaseRef it should
treat as the source-of-truth timebase.**

Reverse-engineering of Apple's host-side handler at vaddr
`0x191584148` in `MediaToolbox.framework` shows that on receipt of a
TBAS packet, the host:
1. Reads the 8-byte payload (the unknown ClockRef from the original doc).
2. Compares it against the currently-stored timebase ref at offset
   `+0x78` of the dispatcher's per-stream state.
3. If different, **stores the new ref** in that slot, overwriting the
   previous value.
4. Calls `_FigRenderPipelineGetFigBaseObject` to update the render
   pipeline binding.

So the original speculation in the doc — *"the device created a
CMTimeBase and just tells us about it"* — is correct. The 8-byte value is
a `CMTimebaseRef` minted on the device side; the host stashes it and
references it for subsequent sample-buffer timing operations on the
device-side timebase. (The reason qvh "could not find another usage"
of the ref is that it's used purely as a timebase identity tag for
later cross-referencing; nothing visible ever queries it back over the
wire.)

###### Packet Format Description

| 4 Byte Length (24)   |4 Byte Magic (ASYN)   | 8 Byte clock CFTypeID  |  4 bytes magic (TBAS) | 8byte Unknown ClockRef|
|---|---|---|---|---|
|18000000| 6E797361| 18BC2311 01000000 |73616274 | C0904402 01000000|

##### 3.3.4. Asyn TJMP - Time Jump Notification
###### General Description
**Confirmed: this is a "time jump on the device's timebase" notification.**

Reverse-engineering of Apple's host-side handler in
`MediaToolbox.framework` (inside the inbound dispatcher; the TJMP
comparison sits at vaddr `0x1915832dc`) shows that the payload is
**not** 56/72 bytes of unknown data — it's structured. The handler
reads:
1. 8 bytes at payload offset 0 → a `CFTypeID` / `CMClockRef` of the
   timebase that jumped.
2. 24 bytes at offset 8 → a `CMTime` — the **anchor time** of the new
   timebase position.
3. 24 bytes at offset 32 → a `CMTime` — the **current time** at the
   moment of the jump.

The handler then calls
[`CMTimebaseCreateReadOnlyTimebaseWithFlags`](https://developer.apple.com/documentation/coremedia/4030977-cmtimebasecreatereadonlytimebase)
with those two `CMTime` values, creating a host-side read-only timebase
that mirrors the device's. So TJMP is the device saying: *"my timebase
has discontinuously jumped — re-anchor your read-only mirror to these
(anchor, current) values."*

This is consistent with `CMTimebase`'s semantics: a read-only timebase
needs an anchor pair to compute future time. Whenever the device's
clock seeks, pauses, or resyncs, it ships a TJMP so the host's mirror
stays accurate.

###### Packet Format Description (corrected)

| 4 Byte Length (72)   |4 Byte Magic (ASYN)   | 8 Byte clock CFTypeID  |  4 bytes magic (TJMP) | 8 bytes timebase CFTypeID | 24 bytes anchor CMTime | 24 bytes current CMTime |
|---|---|---|---|---|---|---|
|48000000| 6E797361| 18BC2311 01000000 |706D6A74 | C0904402 01000000 | …24 bytes… | …24 bytes… |


##### 3.3.5 Asyn FEED - CMSampleBuffer with h264 Video Data
For video data ASYN FEED packets, the device will use the ClockRef we sent as a reply to the CVRP Sync request. 
###### Packet Format Description

| 4 Byte Length (varies, 91607 in this example)   |4 Byte Magic (ASYN)   | 8 Byte clock CFTypeID  |  4 bytes magic (FEED) | 4 bytes length of CMSampleBuf(varies, here 91587)| CMSampleBuf Magic (sbuf) | CMSampleBuf bytes|
|---|---|---|---|---|---|---|
|D7650100| 6E797361| 18BC2311 01000000 |64656566 | C3650100 | 66756273 | ... |



##### 3.3.6 Asyn EAT! - CMSampleBuffer with Audio Data
Just like FEED only with different Magic (0x21746165) and a CMSampleBuf containing audio. 

##### 3.3.7 Asyn NEED - Tell the device to send more
For telling the device to keep sending video data ASYN FEED packets, we need to send NEED packets with the ClockRef the device gave us in the  SYNC CVRP packet.
NEED Packets are constant over the whole session, so you can just init them once you received the correct clockRef and then just keep sending the same bytes over and over.
I think sending NEED packets is something you can do based on a timer (every 5 seconds f.ex.)
For easier implementation I just send one whenever I received a FEED.  
###### Packet Format Description

| 4 Byte Length (20)   |4 Byte Magic (ASYN)   | 8 Byte clock CFTypeID  |  4 bytes magic (NEED) |
|---|---|---|---|
|14000000| 6E797361| A08D5313 01000000 |6465656E | 

##### 3.3.8 Asyn HPD0 - Tell the device to stop video streaming
Send this to stop the device from streaming

###### Packet Format Description

| 4 Byte Length (20)   |4 Byte Magic (ASYN)   | 8 Byte empty clock CFTypeID  |  4 bytes magic (HPD0) |
|---|---|---|---|
|14000000| 6E797361| 01000000 00000000 |30617068 | 

##### 3.3.9 Asyn HPA0 - Tell the device to stop audio streaming

Send this to stop the device from streaming

###### Packet Format Description

| 4 Byte Length (20)   |4 Byte Magic (ASYN)   | 8 Byte clock CFTypeID of audio clock on device  |  4 bytes magic (HPA0) |
|---|---|---|---|
|14000000| 6E797361| 10FCC502 01000000 |30617068 | 

##### 3.3.10 Asyn RELS - Tell us about a released Clock on the device

###### Packet Format Description

| 4 Byte Length (20)   |4 Byte Magic (ASYN)   | 8 Byte clock CFTypeID of audio clock on device  |  4 bytes magic (RELS) |
|---|---|---|---|
|14000000| 6E797361| 008A6035 BA7F0000 | 736C6572 | 

## 4 Serializing/Deserializing Objects
### 4.0 General Description
This chapter explains how to serialize and deserialize all the necessary payload objects that you will find in the various SYNC and ASYN packets. 
### 4.1 Dictionaries
#### 4.1.0 General Description
Dictionaries are used throughout the protocol so they are pretty important to get right :-D
They are pretty easy to implement however note that there are two distinct types. Some dictionaries use only strings as keys and others use only ints or index numbers as keys.
Sometimes you will see dictionaries with other magic markers or single key value entries, but they always work the same way.
#### 4.1.1 Dictionaries with String Keys
##### 4.1.1.1 General Dictionary Structure

Dictionaries always start with a length int, dict magic which is then followed by a number of key value pairs each starting with a length field and keyv magic.
Every entry has a key starting with a 4byte int keylength and then followed by a strk magic int. Finally a string with the actual key.
Values work the same way as they start with a length, then a magic and the actual value. 
This example of a string key dictionary containing one boolean value nicely illustrates how dictionaries work. 

| 4 Byte Length (40)   |4 Byte Magic (DICT)   | 4byte length of first key value pair |  4 bytes magic of first key value pair (KEYV) | 4 byte length of first key(15)| stringkey magic (strk)| key string (Valeria) | 4 byte length of value(9)| 4byte value type magic (bulv==boolean) | value (0x1 == true) |
|---|---|---|---|---|---|---|---|---|---|
|28000000| 74636964| 20000000 |7679656B | 0F000000|6B727473|56616C65 726961|09000000|766C7562| 01|

Here are the value types I know about:

| magic little endian| magic big endian | description | value example |
|---|---|---|---|
|vlub|bulv|Boolean|0x1 or 0x0|
|vrts|strv|String|BlaBla|
|vtad|datv|Byte Array|0x010203|
|vbmn|nmbv|NSNumber|[4.4 NSNumber](#44-nsnumber)|
|tcid|dict|String or Index Key dict| see above |
|csdf|fdsc|CMFormatDescription|[4.5 CMFormatDescription](#45-cmformatdescription)|
|vlru|urlv|CFURL — body is the UTF-8 of `CFURLCopyAbsoluteURL(value).GetString()`. Same on-wire format as `strv`, just a different magic. |https://example.com/x|
|vetd|dtev|CFDate — body is exactly 8 bytes, an IEEE-754 little-endian `double` representing seconds since the **CFAbsoluteTime epoch (2001-01-01 00:00:00 UTC)**.|`0.0` is exactly that epoch|
|yara|aray|CFArray — wrapping atom whose body is a sequence of nested value atoms (each one a recursive value of any of the types in this table, including another `aray`).|`[strv("hi"), bulv(true)]`|
|srlc|clrs|CGColorSpace — 1-byte body. `0x01` = `kCGColorSpaceDeviceRGB`, `0x00` = `kCGColorSpaceDeviceGray`. Apple's serializer (`sbufAtom_appendColorSpaceAtom`) errors out on any other CGColorSpace.|`0x01`|

The four entries above (`urlv`, `dtev`, `aray`, `clrs`) were not in earlier versions of this document; they were discovered by reversing Apple's modern `sbufAtom_appendCFTypeAtom` (in `iOSScreenCaptureAssistant`, source file `FigSampleBufferAtomSerialization.c`). qvh implements all of them as of the version that ships this doc — see `screencapture/coremedia/dict.go` and `dict_serializer.go`.


#### 4.1.2 Dictionaries with 4-Byte Integer Index Keys
They work the same way as String Key Dictionaries with the only difference that all keys are 4 byte integers and they have 0x6B786469 (idxk) as a magic marker.
### 4.2 CMTime
This is exactly like in the CMTime.h https://github.com/phracker/MacOSX-SDKs/blob/master/MacOSX10.8.sdk/System/Library/Frameworks/CoreMedia.framework/Versions/A/Headers/CMTime.h
And there is plenty of documentation so check it out there :-D

#### Example
|CMTimeValue  |CMTimeScale (Nanoseconds is the default)   | CMTimeFlags |  CMTimeEpoch |
|---|---|---|---|
|CB448EA1 CF10CC15| 00CA9A3B| 01000000 |00000000 00000000 | 


### 4.3 CMSampleBuffer
### 4.4 NSNumber
A very simple represenation of what probably is a NSNumber.
I have seen three different types:
- Type 3, 32 bit Integer
- Type 4, 64 bit Integer
- Type 6, 64 bit Float

#### Example
|4 byte Magic (nmbv)  |4 Byte int type   | Number, either 4 or 8 bytes depending on type |
|---|---|---|
|76626D6E | 03000000| 01000000 |

#### Note on the full set of types

Apple's encoder uses the generic `[u8 CFNumberType][N raw bytes]` layout
where `N = CFNumberGetByteSize(value)`. All 16 `CFNumberType` values
defined by CoreFoundation can in principle appear on the wire (kCFNumberSInt8Type=1
through kCFNumberCGFloatType=16). The three types listed above (3, 4, 6) are
the ones observed during the original reverse engineering, but a robust
parser should handle the rest gracefully rather than panicking.



### 4.5 CMFormatDescription
Check out https://github.com/phracker/MacOSX-SDKs/blob/master/MacOSX10.9.sdk/System/Library/Frameworks/CoreMedia.framework/Versions/A/Headers/CMFormatDescription.h


## 5. Clocks and CMSync
I think the references in ASYN and SYNC packets are for CMClocks. So for sending a CMTime request I just a monotonic (DO NOT USE WALLCLOCK TIME) clock
to send a time difference in nanoseconds (Scale == 1000000000). It seems to work fine :-D

## 6. Format Constraints — What's Locked vs What's Negotiable

This section documents empirical results from **testing actual HPD1 dictionary
variants against a current iOS device** (iPhone, iOS 18), comparing the
recorded output against the changes Apple's modern host code (in
`MediaToolbox.framework/FigNero` and `iOSScreenCaptureAssistant`) makes to
the HPD1 dict at runtime.

### What the protocol locks (do not bother changing in qvh)

- **Video codec is always H.264.** Apple's `HPD1` builder advertises HEVC
  decoder capability via the dictionary keys `HEVCDecoderSupports444` and
  (when the user opts in via `defaults write com.apple.coremedia neroEnableHEVC44410 -int 1`)
  `HEVCDecoderSupports44410`. When tested against a current iPhone, the
  device **ignores these keys** — adding `HEVCDecoderSupports44410: true`
  to the HPD1 dict produces an output whose codec is still
  `H.264 / High profile` (verified with `ffprobe`). The HEVC capability
  advertisement is decorative on this transport; the iOS-USB QuickTime
  mirror protocol always emits H.264 NALUs in the FEED packets.
- **Audio codec is always LPCM 48 kHz, 16-bit, stereo.** Apple's
  `iOSScreenCaptureAssistant` references no other audio FormatID
  (`lpcm`/`0x6c70636d`) anywhere in its disassembly — no `aac `, `ac-3`,
  `ec-3`, etc. The AAC/AC-3 references that appear in `MediaToolbox.framework`
  are for HLS/AirPlay/AVFoundation, not for this transport.
- **Video resolution is always the iPhone's native screen resolution.**
  The HPD1 dict has a `DisplaySize: { Width, Height }` sub-dict, but the
  device ignores the requested size — setting `DisplaySize: 640x480` still
  produces a stream at the device's native resolution
  (e.g., 1206×2622 on a recent iPhone). `DisplaySize` is advisory.

### What the HPD1 dict *does* require

- **`Valeria: true` is mandatory.** Removing this key from qvh's HPD1
  dict (e.g., to send a "pure Apple-style" dict containing only
  `H264DecoderSupports444 / HEVCDecoderSupports444 / HEVCDecoderSupports44410`)
  causes the device's response stream to be malformed: the FEED packets
  arrive without proper PPS, and `ffprobe` reports
  `non-existing PPS 0 referenced` / `decode_slice_header error`. The
  resulting `.h264` file is a few kilobytes of garbage instead of the
  expected several megabytes per second.
  Empirical takeaway: **`Valeria: true` must stay** even though Apple's
  own host code does not include it. It's part of the activation contract.

### Which dictionary keys Apple's modern host code uses

Reverse-engineered from `iOSScreenCaptureAssistant`'s call into the HPD1
builder helper in `MediaToolbox.framework`:

```c
// Pseudocode of Apple's HPD1 builder (function at vaddr 0x1915827e0 on macOS 15.5):
buildHpd1Dict() {
    if (FVDUtilsH264DecoderSupports444()) {
        dict["H264DecoderSupports444"] = true;
    }
    if (FVDUtilsHEVCDecodeSupported()) {
        if (CFPreferencesGetAppIntegerValue("neroEnableHEVC", "com.apple.coremedia") >= 1) {
            dict["HEVCDecoderSupports444"] = true;
        }
        if (CFPreferencesGetAppIntegerValue("neroEnableHEVC44410", "com.apple.coremedia") >= 1) {
            dict["HEVCDecoderSupports44410"] = true;
        }
    }
    sendHPD1(dict);
}
```

So Apple's HPD1 is purely a runtime capability advertisement that mirrors
the host's own decoder support and user-toggleable preferences. It does
not control the device's encoder behavior on this transport.

### Practical takeaway

qvh's hardcoded `Valeria: true`, `HEVCDecoderSupports444: true`, and
`DisplaySize: 1920x1200` in `CreateHpd1DeviceInfoDict()` are
operationally correct as-is. There is no audio or video format flexibility
to expose; the protocol is **single-codec, native-resolution** by design.
If you ever discover an iOS firmware change that loosens this, please
update this section — but as of iOS 18 / macOS 15.5, the constraint
holds.

## 7. Verifying / Extending This Document with Claude Code

This document was originally hand-reverse-engineered by sniffing USB traffic.
On modern macOS, the same protocol is **still implemented end-to-end by
Apple's own host-side stack** (QuickTime Player → CoreMediaIO →
`iOSScreenCaptureAssistant` → MediaToolbox/`FigNero` → CoreMedia/`FigTransportConnectionUSB` →
IOUSBLib bulk endpoints). That means anyone with a Mac can cross-check
this document against Apple's compiled implementation, find new packet
types Apple may have added, and submit fixes.

The easiest way to do this is to paste the prompt below into
[Claude Code](https://claude.com/claude-code) inside a fresh checkout of
[`quicktime_video_hack`](https://github.com/danielpaulus/quicktime_video_hack)
on a Mac with the iPhone screen-recording stack installed (i.e., any
recent macOS with QuickTime Player.app). Claude will install the tools it
needs, extract the relevant Apple frameworks from the dyld shared cache,
disassemble them, and report concrete diffs against this repo.

### Required tools (Claude will install if missing)
- [`ipsw`](https://github.com/blacktop/ipsw) — pulls individual frameworks out of `dyld_shared_cache_arm64e`
- Ghidra (optional, for decompilation; otool / `dyld_info` / `nm` are usually enough)
- An iPhone connected via USB (only needed for the empirical activation test)

### Copy-paste verification prompt

> Reverse-engineer Apple's modern macOS implementation of the QuickTime
> iOS-USB screen-mirroring protocol and audit
> `doc/technical_documentation.md` in this repo against it.
> Report concrete divergences as patchable diffs.
>
> **Key context the user already verified, do not re-derive:**
> - The iOS device retains the legacy "QuickTime mirror" USB
>   configuration descriptor and reveals it on demand. macOS's modern
>   stack still uses it, just split across two frameworks.
> - Apple's host-side code lives in two extracted dylibs from
>   `/System/Volumes/Preboot/Cryptexes/OS/System/Library/dyld/dyld_shared_cache_arm64e`:
>     * `CoreMedia.framework` — transport layer (`FigTransportConnectionUSB`,
>       `usb_clientThreadMain`, `usb_readCompleted`, `usb_messageSendingThreadMain`,
>       `usb_clientSendStartupPing`, etc.) — owns the
>       `[u32 LE length][u32 LE FourCC magic][payload]` framing where length
>       INCLUDES the 4 length bytes, plus the ping/keepalive.
>     * `MediaToolbox.framework` — protocol semantics (`FigNero`) — owns
>       the inner sync/asyn subtype dispatch (cwpa, cvrp, afmt, feed, eat!,
>       hpd0, hpd1, hpa0, clok, skew, srat, sprp, tbas, tjmp, need, rels).
> - Atom serialization (CMSampleBuffer → QuickTime atoms) lives in
>   `/System/Library/Frameworks/CoreMediaIO.framework/Versions/A/Resources/iOSScreenCapture.plugin/Contents/Resources/iOSScreenCaptureAssistant`,
>   in functions named `sbufAtom_*` (the source file is
>   `FigSampleBufferAtomSerialization.c`, leaked in the binary as a
>   `__cstring` reference). Apple's `sbufAtom_appendCFTypeAtom`
>   dispatches on 10 CFType variants whose magics map to the value-type
>   table in `doc/technical_documentation.md` §4.1.
>
> **Methodology gotcha (read this before you grep for FourCC bytes):**
> arm64 has no single 32-bit-immediate move. A constant like
> `0x61666D74` ("afmt") is loaded as `movz wN, #0x6D74; movk wN, #0x6166, lsl #16`,
> so the four ASCII bytes `61 66 6D 74` are NEVER contiguous in the binary.
> A naïve `xxd | grep` will not find them. Either disassemble and
> reconstruct the immediate from the movz/movk pair, or scan
> `__TEXT,__const` and `__DATA_CONST,__const` for u32 windows that decode
> to printable ASCII (in either endianness — Apple stores them
> little-endian numerically so the bytes appear reversed).
>
> **Concrete steps:**
>
> 1. Install `ipsw` if missing: `brew install ipsw`. Optional: `brew install ghidra`
>    if you want decompilation. Otherwise `otool`, `nm`, `dyld_info` from
>    Xcode Command Line Tools are sufficient.
>
> 2. Extract the three relevant binaries to `/tmp/qvh-re/`:
>    ```bash
>    mkdir -p /tmp/qvh-re
>    DSC=/System/Volumes/Preboot/Cryptexes/OS/System/Library/dyld/dyld_shared_cache_arm64e
>    ipsw dyld extract "$DSC" CoreMedia    -o /tmp/qvh-re --slide --objc
>    ipsw dyld extract "$DSC" MediaToolbox -o /tmp/qvh-re --slide --objc
>    lipo -thin arm64e -output /tmp/qvh-re/iSCAssistant.arm64e \
>      "/System/Library/Frameworks/CoreMediaIO.framework/Versions/A/Resources/iOSScreenCapture.plugin/Contents/Resources/iOSScreenCaptureAssistant"
>    ```
>
> 3. Dump every FourCC immediate Apple's code uses (handles the arm64
>    movz/movk trap correctly). Use a small Python script that scans
>    `otool -arch arm64e -tV <binary>` output, reconstructing each
>    32-bit immediate by pairing `movz`/`movk` instructions targeting the
>    same register, then filtering for printable-ASCII results. Run it
>    against all three binaries and produce three sorted FourCC lists.
>
> 4. Diff each list against the constants defined in qvh:
>    - `screencapture/packet/sync.go`, `packet/asyn.go`, `packet/ping.go`,
>      and the per-subtype `packet/sync_*.go` / `packet/asyn_*.go` files —
>      these are the wire-protocol subtypes and should match what's in
>      `MediaToolbox`.
>    - `screencapture/coremedia/dict.go` (and `common/nsnumber.go`) — the
>      atom value-type magics and any new ones from `iSCAssistant`'s
>      `sbufAtom_appendCFTypeAtom` (which lives at the address that
>      contains immediates `0x6e6d6276` (nmbv) and `0x73747276` (strv) —
>      the dispatcher hub).
>    - `screencapture/coremedia/cmsamplebuf.go` — the CMSampleBuffer atom
>      layout. Compare against `iSCAssistant`'s `sbufAtom_appendSampleSizes`,
>      `sbufAtom_appendSampleTimingInfo`, etc.
>
>    Highlight any FourCC Apple uses that qvh does not. For each, find
>    the function it lives in (walk backward from the immediate's address
>    looking for the `pacibsp` prologue) and read enough surrounding
>    disassembly to determine the body layout (length, fields, byte
>    order). Report the on-wire byte format, the corresponding qvh source
>    file, and the concrete change needed.
>
> 5. Optional empirical sanity check: while an iPhone is plugged in,
>    capture `ioreg -p IOUSB -l` and `system_profiler SPUSBDataType`,
>    then open QuickTime Player → File → New Movie Recording → select
>    the iPhone, and re-capture. The iPhone should re-enumerate and
>    `bNumConfigurations` should increment by 1, with
>    `kUSBCurrentConfiguration` switching into the new config. That
>    confirms QuickTime Player is using the same legacy USB
>    configuration descriptor that qvh activates via
>    `Control(0x40, 0x52, 0x00, 0x02, response)`.
>
> 6. Output a `FINDINGS.md` summarizing:
>    - The architectural map (transport vs semantics layering)
>    - For each diverging FourCC: file:line in qvh + concrete byte layout + a one-line patch description
>    - Anything you couldn't determine from disasm alone
>    - The full disassembly methodology you used (so the next person can reproduce)
>
> 7. If patches are obvious and small, write them as concrete code
>    changes against the qvh source tree (do not commit). Run
>    `go test ./...` and only call the patches done if all tests pass.
>    Do NOT push or create PRs without explicit approval.

The prompt above intentionally tells Claude what's already known so it
doesn't waste time rediscovering the architecture, and warns about the
specific arm64-FourCC-search trap that consumed many hours during the
original investigation.
