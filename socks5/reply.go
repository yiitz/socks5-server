package socks5

import (
	"bytes"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type UDPReply struct {
	config       Config
	localUDPConn *net.UDPConn
}

type uProxy struct {
	config        Config
	localConn     *net.UDPConn
	dstMap        sync.Map
	remoteConnMap sync.Map
}

func (u *UDPReply) Start() {
	udpAddr, _ := net.ResolveUDPAddr("udp", u.config.LocalAddr)
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Printf("[udp] failed to listen udp %v", err)
		return
	}
	u.localUDPConn = udpConn
	defer u.localUDPConn.Close()
	log.Printf("socks5-server [udp] started on %v", u.config.LocalAddr)
	u.proxy()
}

func (u *UDPReply) proxy() {
	proxy := &uProxy{localConn: u.localUDPConn, config: u.config}
	proxy.toRemote()
}

func (proxy *uProxy) toRemote() {
	buf := make([]byte, BufferSize)
	for {
		proxy.localConn.SetReadDeadline(time.Now().Add(time.Duration(Timeout) * time.Second))
		n, cliAddr, err := proxy.localConn.ReadFromUDP(buf)
		if err != nil || err == io.EOF || n == 0 {
			continue
		}
		b := buf[:n]
		dstAddr, header, data := proxy.getAddr(b)
		if dstAddr == nil || header == nil || data == nil {
			continue
		}
		key := cliAddr.String()
		var remoteConn *net.UDPConn
		if value, ok := proxy.remoteConnMap.Load(key); ok {
			remoteConn = value.(*net.UDPConn)
		} else {
			remoteConn, err := net.DialUDP("udp", nil, dstAddr)
			if err != nil {
				continue
			}
			proxy.remoteConnMap.Store(key, remoteConn)
			proxy.dstMap.Store(key, header)
			go proxy.toLocal(remoteConn, cliAddr)
		}
		remoteConn.Write(b)
	}
}

func (proxy *uProxy) toLocal(remoteConn *net.UDPConn, cliAddr *net.UDPAddr) {
	defer remoteConn.Close()
	key := cliAddr.String()
	buf := make([]byte, BufferSize)
	remoteConn.SetReadDeadline(time.Now().Add(time.Duration(Timeout) * time.Second))
	for {
		n, _, err := remoteConn.ReadFromUDP(buf)
		if n == 0 || err != nil {
			break
		}
		if header, ok := proxy.dstMap.Load(key); ok {
			var data bytes.Buffer
			data.Write(header.([]byte))
			data.Write(buf[:n])
			proxy.localConn.WriteToUDP(data.Bytes(), cliAddr)
		}
	}
	proxy.dstMap.Delete(key)
	proxy.remoteConnMap.Delete(key)
}

func (proxy *uProxy) getAddr(b []byte) (dstAddr *net.UDPAddr, header []byte, data []byte) {
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
				IP:   net.IP(b[4:19]),
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
