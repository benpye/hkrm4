package broadlink

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"
)

const learnTimeout = 30 // seconds
const sendRetries = 3

var initialKey = [...]byte{0x09, 0x76, 0x28, 0x34, 0x3f, 0xe9, 0x9e, 0x23, 0x76, 0x5c, 0x15, 0x13, 0xac, 0xcf, 0x8b, 0x02}
var initialIV = [...]byte{0x56, 0x2e, 0x17, 0x99, 0x6d, 0x09, 0x3d, 0x28, 0xdd, 0xb3, 0xba, 0x69, 0x5a, 0x2e, 0x6f, 0x58}
var initialID = [...]byte{0, 0, 0, 0}

// ResponseType denotes the type of payload.
type ResponseType int

// Enumerations of PayloadType.
const (
	Unknown ResponseType = iota
	AuthOK
	DeviceError
	Temperature
	CommandOK
	RawData
	RawRFData
	RawRFData2
)

// Response represents a decrypted payload from the device.
type Response struct {
	Type ResponseType
	Data []byte
}

type Device struct {
	remoteAddr        net.IP
	timeout           int
	deviceType        int
	mac               net.HardwareAddr
	count             uint16
	key               []byte
	iv                []byte
	id                []byte
	requestHeader     []byte
	codeSendingHeader []byte
}

type unencryptedRequest struct {
	command byte
	payload []byte
}

func newDevice(remoteAddr net.IP, mac net.HardwareAddr, timeout int, devChar deviceCharacteristics) (*Device, error) {
	rand.Seed(time.Now().Unix())
	d := &Device{
		remoteAddr:        remoteAddr,
		timeout:           timeout,
		deviceType:        devChar.deviceType,
		mac:               mac,
		count:             uint16(rand.Uint32()),
		key:               initialKey[:],
		iv:                initialIV[:],
		id:                initialID[:],
		requestHeader:     devChar.requestHeader,
		codeSendingHeader: devChar.codeSendingHeader,
	}

	_, err := d.serverRequest(authenticatePayload())
	if err != nil {
		return d, fmt.Errorf("error making authentication request: %v", err)
	}

	return d, nil
}

// serverRequest sends a request to the device and waits for a response.
func (d *Device) serverRequest(req unencryptedRequest) ([]byte, error) {
	encryptedReq, err := d.encryptRequest(req)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenPacket("udp4", "")
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	retries := 0
	for {
		retries++

		destAddr, err := net.ResolveUDPAddr("udp", d.remoteAddr.String()+":80")
		if err != nil {
			err = fmt.Errorf("could not resolve device address %v: %v", d.remoteAddr, err)
			return nil, err
		}

		_, err = conn.WriteTo(encryptedReq, destAddr)
		if err != nil {
			if retries < sendRetries {
				continue
			}

			err = fmt.Errorf("could not send packet: %v", err)
			return nil, err
		}

		conn.SetReadDeadline(time.Now().Add(time.Duration(d.timeout) * time.Second))

		var buf [2048]byte
		plen, _, err := conn.ReadFrom(buf[:])
		if err != nil {
			return nil, fmt.Errorf("error while waiting for device response: %v", err)
		}

		if plen < 0x30 {
			return nil, fmt.Errorf("expected at least 0x30 bytes, got: %d", plen)
		}

		err = d.checkError(buf[:plen])
		if err != nil {
			return nil, err
		}

		resp, err := d.decryptResponse(buf[:plen])
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
}

func (d *Device) encryptRequest(req unencryptedRequest) ([]byte, error) {
	if len(req.payload)%16 != 0 {
		return []byte{}, fmt.Errorf("length of unencrypted request payload must be a multiple of 16 - got %d instead", len(req.payload))
	}

	d.count = d.count + 1

	header := make([]byte, 0x38, 0x38)
	header[0x00] = 0x5a
	header[0x01] = 0xa5
	header[0x02] = 0xaa
	header[0x03] = 0x55
	header[0x04] = 0x5a
	header[0x05] = 0xa5
	header[0x06] = 0xaa
	header[0x07] = 0x55
	header[0x24] = 0x2a
	header[0x25] = 0x27
	header[0x26] = req.command
	header[0x28] = (byte)(d.count & 0xff)
	header[0x29] = (byte)(d.count >> 8)
	header[0x2a] = d.mac[5]
	header[0x2b] = d.mac[4]
	header[0x2c] = d.mac[3]
	header[0x2d] = d.mac[2]
	header[0x2e] = d.mac[1]
	header[0x2f] = d.mac[0]
	header[0x30] = d.id[0]
	header[0x31] = d.id[1]
	header[0x32] = d.id[2]
	header[0x33] = d.id[3]

	checksum := 0xbeaf
	for _, v := range req.payload {
		checksum += (int)(v)
		checksum = checksum & 0xffff
	}

	block, err := aes.NewCipher(d.key)
	if err != nil {
		return []byte{}, fmt.Errorf("unable to create new AES cipher: %v", err)
	}
	mode := cipher.NewCBCEncrypter(block, d.iv)
	encryptedPayload := make([]byte, len(req.payload))
	mode.CryptBlocks(encryptedPayload, req.payload)

	packet := make([]byte, len(header)+len(encryptedPayload))
	copy(packet, header)
	copy(packet[len(header):], encryptedPayload)

	packet[0x34] = (byte)(checksum & 0xff)
	packet[0x35] = (byte)(checksum >> 8)

	checksum = 0xbeaf
	for _, v := range packet {
		checksum += (int)(v)
		checksum = checksum & 0xffff
	}
	packet[0x20] = (byte)(checksum & 0xff)
	packet[0x21] = (byte)(checksum >> 8)

	return packet, nil
}

func (d *Device) checkError(resp []byte) error {
	errorCode := (int)(resp[0x22]) | ((int)(resp[0x23]) << 8)
	if errorCode != 0 {
		return fmt.Errorf("error code %d", errorCode)
	}
	return nil
}

func (d *Device) decryptResponse(resp []byte) ([]byte, error) {
	encryptedPayload := make([]byte, len(resp)-0x38, len(resp)-0x38)
	copy(encryptedPayload, resp[0x38:])

	block, err := aes.NewCipher(d.key)
	if err != nil {
		return nil, fmt.Errorf("error creating new decryption cipher: %v", err)
	}
	payload := make([]byte, len(encryptedPayload), len(encryptedPayload))
	mode := cipher.NewCBCDecrypter(block, d.iv)

	if len(encryptedPayload)%16 != 0 {
		return nil, errors.New("crypto/cipher: input not full blocks")
	}

	mode.CryptBlocks(payload, encryptedPayload)

	// Update IV and key from auth response.
	command := resp[0x26]
	if command == 0xe9 {
		copy(d.key, payload[0x04:0x14])
		copy(d.id, payload[:0x04])
	}

	return payload[len(d.requestHeader)+0x4:], nil
}

func (d *Device) SendData(data []byte) error {
	header := d.codeSendingHeader

	reqLength := (len(header) + len(data) + 4 + 15) / 16 * 16
	reqPayload := make([]byte, reqLength, reqLength)
	copy(reqPayload, header)

	reqPayload[len(header)] = 0x02
	reqPayload[len(header)+1] = 0x00
	reqPayload[len(header)+2] = 0x00
	reqPayload[len(header)+3] = 0x00
	copy(reqPayload[len(header)+4:], data)
	req := unencryptedRequest{
		command: 0x6a,
		payload: reqPayload,
	}

	_, err := d.serverRequest(req)
	if err != nil {
		return fmt.Errorf("error reading response while trying to send data to device: %v", err)
	}

	return nil
}

func (d *Device) CheckSensors() (float64, float64, error) {
	resp, err := d.serverRequest(d.basicPayload(0x24))
	if err != nil {
		return 0, 0, fmt.Errorf("error making CheckSensors request: %v", err)
	}

	temperature := float64(resp[0]) + float64(resp[1])/100.0
	humidity := float64(resp[2]) + float64(resp[3])/100.0

	return temperature, humidity, nil
}

func authenticatePayload() unencryptedRequest {
	req := unencryptedRequest{
		command: 0x65,
		payload: make([]byte, 0x50, 0x50),
	}
	req.payload[0x04] = 0x31
	req.payload[0x05] = 0x31
	req.payload[0x06] = 0x31
	req.payload[0x07] = 0x31
	req.payload[0x08] = 0x31
	req.payload[0x09] = 0x31
	req.payload[0x0a] = 0x31
	req.payload[0x0b] = 0x31
	req.payload[0x0c] = 0x31
	req.payload[0x0d] = 0x31
	req.payload[0x0e] = 0x31
	req.payload[0x0f] = 0x31
	req.payload[0x10] = 0x31
	req.payload[0x11] = 0x31
	req.payload[0x12] = 0x31
	req.payload[0x1e] = 0x01
	req.payload[0x2d] = 0x01
	req.payload[0x30] = 'T'
	req.payload[0x31] = 'e'
	req.payload[0x32] = 's'
	req.payload[0x33] = 't'
	req.payload[0x34] = ' '
	req.payload[0x35] = ' '
	req.payload[0x36] = '1'

	return req
}

func (d *Device) basicPayload(command byte) unencryptedRequest {
	payload := make([]byte, 16, 16)
	header := d.requestHeader
	copy(payload, header)
	payload[len(header)] = command

	return unencryptedRequest{
		command: 0x6a,
		payload: payload,
	}
}
