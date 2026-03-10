package netcode

import (
	"errors"
	"net"

	"github.com/Cam1G/Landlord-RTS/internal/protocol"
)

func SendMessage(conn net.Conn, command uint16, content string) error {
	buf := make([]byte, 4)
	buf[2] = byte((command >> 8) & 0xff)
	buf[3] = byte(command & 0xff)
	buf = append(buf, []byte(content)...)
	buf[0] = byte(((len(buf) - 2) >> 8) & 0xff)
	buf[1] = byte((len(buf) - 2) & 0xff)
	_, err := conn.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

func RecvMessage(conn net.Conn) (uint16, string, error) {
	lenb := make([]byte, 2)
	_, err := conn.Read(lenb)
	if err != nil {
		return 0, "", err
	}
	len := uint16(lenb[0])<<8 | uint16(lenb[1])
	buf := make([]byte, len)
	_, err = conn.Read(buf)
	if err != nil {
		return 0, "", err
	}
	return uint16(buf[0])<<8 | uint16(buf[1]), string(buf[2:]), nil
}

func RecvMessageHandleErr(conn net.Conn, expProtocol uint16) (string, error) {
	cmd, resp, err := RecvMessage(conn)
	if err != nil {
		return "", err
	}
	if cmd != expProtocol {
		return "", errors.New("Wrong response")
	}
	statusCode := resp[0]
	switch cmd {
	case 0:
		break
	case protocol.AuthCreateUser:
		switch statusCode {
		case 's':
			return "", nil
		case 'x':
			return "", errors.New("User already exists")
		default:
			return "", errors.New("Server error")
		}
	case protocol.Auth:
		switch statusCode {
		case 's':
			return "", nil
		case 'd':
			return "", errors.New("User does not exist")
		case 'p':
			return "", errors.New("Wrong password")
		default:
			return "", errors.New("Server error")
		}
	}
	return resp[1:], nil
}
