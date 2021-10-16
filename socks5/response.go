package socks5

import (
	"bytes"
	"encoding/binary"
	"net"
)

func responseAuthType(conn net.Conn, authType uint8) {
	/**
	  +----+--------+
	  |VER | METHOD |
	  +----+--------+
	  | 1  |   1    |
	  +----+--------+
	*/
	conn.Write([]byte{Socks5Version, authType})
}

func responseAuthResult(conn net.Conn, status uint8) {
	/**
	  +----+--------+
	  |VER | STATUS |
	  +----+--------+
	  | 1  |   1    |
	  +----+--------+
	*/
	conn.Write([]byte{UserAuthVersion, status})
}

func responseTCP(conn net.Conn, rep byte) {
	/**
	  +----+-----+-------+------+----------+----------+
	  |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
	  +----+-----+-------+------+----------+----------+
	  | 1  |  1  | X'00' |  1   | Variable |    2     |
	  +----+-----+-------+------+----------+----------+
	*/
	conn.Write([]byte{Socks5Version, rep, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func responseUDP(conn net.Conn, bindAddr *net.UDPAddr) {
	/**
	  +----+-----+-------+------+----------+----------+
	  |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
	  +----+-----+-------+------+----------+----------+
	  | 1  |  1  | X'00' |  1   | Variable |    2     |
	  +----+-----+-------+------+----------+----------+
	*/
	response := []byte{Socks5Version, SuccessReply, 0x00, 0x01}
	buffer := bytes.NewBuffer(response)
	binary.Write(buffer, binary.BigEndian, bindAddr.IP.To4())
	binary.Write(buffer, binary.BigEndian, uint16(bindAddr.Port))
	conn.Write(buffer.Bytes())
}
