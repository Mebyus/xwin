package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"xwin/x11"
)

func fatal(v any) {
	fmt.Fprintf(os.Stderr, "%v\n", v)
	os.Exit(1)
}

func main() {
	sock, err := net.Dial("unix", "/tmp/.X11-unix/X0")
	if err != nil {
		fatal(err)
	}
	defer sock.Close()

	err = sendInitRequest(sock)
	if err != nil {
		fatal(err)
	}
}

func sendInitRequest(sock net.Conn) error {
	var buf [12]byte

	buf[0] = 'l' // indicate little endian byte order (for big endian use 'B')
	buf[2] = 11  // protocol version

	_, err := sock.Write(buf[:])
	if err != nil {
		return fmt.Errorf("send init: %w", err)
	}

	// read response to init request

	// read first 8 bytes to determine response type
	_, err = sock.Read(buf[:8])
	if err != nil {
		return fmt.Errorf("receive init repsonse: %w", err)
	}

	switch buf[0] {
	case x11.StatusFailed: // fail
		reasonLength := buf[1]
		majorVersion := binary.LittleEndian.Uint16(buf[2:4])
		minorVersion := binary.LittleEndian.Uint16(buf[4:6])
		dataLength := binary.LittleEndian.Uint16(buf[6:8]) // Length is encoded as 4-byte units

		var reason [256]byte
		n, err := sock.Read(reason[:reasonLength])
		if err != nil {
			return fmt.Errorf("read init fail reason: %w", err)
		}

		fmt.Fprintf(os.Stderr, "status: %d\n", buf[0])
		fmt.Fprintf(os.Stderr, "version: %d.%d\n", majorVersion, minorVersion)
		fmt.Fprintf(os.Stderr, "data length: %d\n", dataLength)
		fmt.Fprintf(os.Stderr, "reason: %s\n", reason[:n])
		return fmt.Errorf("init failed: %s", reason[:n])
	case x11.StatusSuccess: // success
		fmt.Fprintf(os.Stderr, "init returned success\n")

		majorVersion := binary.LittleEndian.Uint16(buf[2:4])
		minorVersion := binary.LittleEndian.Uint16(buf[4:6])
		dataLength := uint32(binary.LittleEndian.Uint16(buf[6:8])) << 2 // Length is encoded as 4-byte units

		fmt.Fprintf(os.Stderr, "status: %d\n", buf[0])
		fmt.Fprintf(os.Stderr, "version: %d.%d\n", majorVersion, minorVersion)
		fmt.Fprintf(os.Stderr, "data length: %d\n", dataLength)
		// TODO: check minimal data length 
		return readInitResponse(sock, dataLength)
	case x11.StatusAuth: // auth
		return fmt.Errorf("auth not implemented")
	default:
		return fmt.Errorf("unexpected init response type %d", buf[0])
	}

	return nil
}

func readInitResponse(sock net.Conn, size uint32) error {
	var buf [1 << 16]byte
	n, err := sock.Read(buf[:size])
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "bytes read from response: %d\n", n)
	if n != int(size) {
		return fmt.Errorf("unexpected response data length %d", n)
	}

	resourceIdBase := binary.LittleEndian.Uint32(buf[4:8])
	resourceIdMask := binary.LittleEndian.Uint32(buf[8:12])
	lengthOfVendor := binary.LittleEndian.Uint16(buf[16:18])
	numberOfFormants := binary.LittleEndian.Uint16(buf[21:23])
	// *Vendor = (uint8_t *)&ReadBuffer[40]; (- 8 bytes to buf start)

	fmt.Fprintf(os.Stderr, "id base: 0x%x\n", resourceIdBase)
	fmt.Fprintf(os.Stderr, "id mask: 0x%x\n", resourceIdMask)
	fmt.Fprintf(os.Stderr, "vendor length: %d\n", lengthOfVendor)

	// TODO: make sure that this value is aligned by 4
	// vendorPadLength := uint32(lengthOfVendor)
	formatByteLength := uint32(numberOfFormants) << 3
	screensStartOffset := 32 + uint32(lengthOfVendor) + uint32(formatByteLength)
	fmt.Fprintf(os.Stderr, "screens offset: %d\n", screensStartOffset)

	rootWindowID := binary.LittleEndian.Uint32(buf[screensStartOffset : screensStartOffset+4])
	rootVisualID := binary.LittleEndian.Uint32(buf[screensStartOffset+32 : screensStartOffset+36])
	fmt.Fprintf(os.Stderr, "root window id: 0x%x\n", rootWindowID)
	fmt.Fprintf(os.Stderr, "root visual id: 0x%x\n", rootVisualID)

	return nil
}
