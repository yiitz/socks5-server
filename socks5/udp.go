package socks5

import (
	"log"
	"net"
)

func UDPProxy(tcpConn net.Conn, udpConn *net.UDPConn, config Config) {
	defer tcpConn.Close()
	if udpConn == nil {
		log.Printf("[udp] failed to start udp server on %v", config.LocalAddr)
		return
	}
	bindAddr, _ := net.ResolveUDPAddr("udp", udpConn.LocalAddr().String())
	//send response to client
	responseUDP(tcpConn, bindAddr)
	//keep tcp conn
	done := make(chan bool)
	go keepUDPAlive(tcpConn.(*net.TCPConn), done)
	<-done
}

func keepUDPAlive(tcpConn *net.TCPConn, done chan<- bool) {
	tcpConn.SetKeepAlive(true)
	buf := make([]byte, BufferSize)
	for {
		_, err := tcpConn.Read(buf[0:])
		if err != nil {
			break
		}
	}
	done <- true
}
