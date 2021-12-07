package socks5

import (
	"bytes"
	"io"
	"log"
	"net"
	"sync"
	"time"
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
	go keepTCPAlive(tcpConn.(*net.TCPConn), done)
	<-done
}

func keepTCPAlive(tcpConn *net.TCPConn, done chan<- bool) {
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

type UDPRelay struct {
	Config        Config
	LocalUDPConn  *net.UDPConn
	DstMap        sync.Map
	RemoteConnMap sync.Map
}

func (relay *UDPRelay) Start() *net.UDPConn {
	udpAddr, _ := net.ResolveUDPAddr("udp", relay.Config.LocalAddr)
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Printf("[udp] failed to listen udp %v", err)
		return nil
	}
	relay.LocalUDPConn = udpConn
	go relay.toRemote()
	log.Printf("socks5-server [udp] started on %v", relay.Config.LocalAddr)
	return relay.LocalUDPConn
}

func (relay *UDPRelay) toRemote() {
	defer relay.LocalUDPConn.Close()
	buf := make([]byte, BufferSize)
	for {
		relay.LocalUDPConn.SetReadDeadline(time.Now().Add(time.Duration(Timeout) * time.Second))
		n, cliAddr, err := relay.LocalUDPConn.ReadFromUDP(buf)
		if err != nil || err == io.EOF || n == 0 {
			continue
		}
		b := buf[:n]
		dstAddr, header, data := relay.getAddr(b)
		if dstAddr == nil || header == nil || data == nil {
			continue
		}
		key := cliAddr.String()
		value, ok := relay.RemoteConnMap.Load(key)
		if ok && value != nil {
			remoteConn := value.(*net.UDPConn)
			remoteConn.Write(data)
		} else {
			remoteConn, err := net.DialUDP("udp", nil, dstAddr)
			if remoteConn == nil || err != nil {
				log.Printf("failed to dial udp:%v", dstAddr)
				continue
			}
			relay.RemoteConnMap.Store(key, remoteConn)
			relay.DstMap.Store(key, header)
			go relay.toLocal(remoteConn, cliAddr)
			remoteConn.Write(data)
		}
	}
}

func (relay *UDPRelay) toLocal(remoteConn *net.UDPConn, cliAddr *net.UDPAddr) {
	defer remoteConn.Close()
	key := cliAddr.String()
	buf := make([]byte, BufferSize)
	remoteConn.SetReadDeadline(time.Now().Add(time.Duration(Timeout) * time.Second))
	for {
		n, _, err := remoteConn.ReadFromUDP(buf)
		if n == 0 || err != nil {
			break
		}
		if header, ok := relay.DstMap.Load(key); ok {
			var data bytes.Buffer
			data.Write(header.([]byte))
			data.Write(buf[:n])
			relay.LocalUDPConn.WriteToUDP(data.Bytes(), cliAddr)
		}
	}
	relay.DstMap.Delete(key)
	relay.RemoteConnMap.Delete(key)
}

func (relay *UDPRelay) getAddr(b []byte) (dstAddr *net.UDPAddr, header []byte, data []byte) {
	/*
	   +----+------+------+----------+----------+----------+
	   |RSV | FRAG | ATYP | DST.ADDR | DST.PORT |   DATA   |
	   +----+------+------+----------+----------+----------+
	   |  2 |   1  |   1  | Variable |     2    | Variable |
	   +----+------+------+----------+----------+----------+
	*/
	if b[2] != 0x00 {
		log.Printf("[udp] not support frag %v", b[2])
		return nil, nil, nil
	}
	switch b[3] {
	case Ipv4Address:
		dstAddr = &net.UDPAddr{
			IP:   net.IPv4(b[4], b[5], b[6], b[7]),
			Port: int(b[8])<<8 | int(b[9]),
		}
		header = b[0:10]
		data = b[10:]
	case FqdnAddress:
		domainLength := int(b[4])
		domain := string(b[5 : 5+domainLength])
		ipAddr, err := net.ResolveIPAddr("ip", domain)
		if err != nil {
			log.Printf("[udp] failed to resolve dns %s:%v", domain, err)
			return nil, nil, nil
		}
		dstAddr = &net.UDPAddr{
			IP:   ipAddr.IP,
			Port: int(b[5+domainLength])<<8 | int(b[6+domainLength]),
		}
		header = b[0 : 7+domainLength]
		data = b[7+domainLength:]
	case Ipv6Address:
		{
			dstAddr = &net.UDPAddr{
				IP:   net.IP(b[4:20]),
				Port: int(b[20])<<8 | int(b[21]),
			}
			header = b[0:22]
			data = b[22:]
		}
	default:
		return nil, nil, nil
	}
	return dstAddr, header, data
}
