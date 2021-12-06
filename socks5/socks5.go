package socks5

import (
	"crypto/tls"
	"io"
	"log"
	"net"
)

// Start socks5 server
func StartServer(config Config) {
	udpConn := handleUDP(config)
	handleTCP(config, udpConn)
}

func handleTCP(config Config, udpConn *net.UDPConn) {
	var l net.Listener
	var err error
	if config.TLS {
		cer, err := tls.LoadX509KeyPair(config.ServerPem, config.ServerKey)
		if err != nil {
			log.Panic(err)
		}
		c := &tls.Config{Certificates: []tls.Certificate{cer}}
		l, err = tls.Listen("tcp", config.LocalAddr, c)
		if err != nil {
			log.Panicf("[tls] failed to listen tcp %v", err)
		}
		log.Printf("socks5-server [tls] started on %s", config.LocalAddr)
	} else {
		l, err = net.Listen("tcp", config.LocalAddr)
		if err != nil {
			log.Panicf("[tcp] failed to listen tcp %v", err)
		}
		log.Printf("socks5-server [tcp] started on %s", config.LocalAddr)
	}
	for {
		tcpConn, err := l.Accept()
		if err != nil {
			continue
		}
		go tcpHandler(tcpConn, udpConn, config)
	}
}

func handleUDP(config Config) *net.UDPConn {
	udpRelay := &UDPRelay{Config: config}
	return udpRelay.Start()
}

func tcpHandler(tcpConn net.Conn, udpConn *net.UDPConn, config Config) {
	buf := make([]byte, BufferSize)
	// read version
	n, err := tcpConn.Read(buf[0:])
	if err != nil || err == io.EOF {
		return
	}
	b := buf[0:n]
	if b[0] != Socks5Version {
		return
	}
	if config.Username == "" && config.Password == "" {
		// no auth
		responseAuthType(tcpConn, NoAuth)
	} else {
		// username and password auth
		responseAuthType(tcpConn, UserPassAuth)
		username, password := getUserPwd(tcpConn)
		if username == config.Username && password == config.Password {
			responseAuthResult(tcpConn, AuthSuccess)
		} else {
			responseAuthResult(tcpConn, AuthFailure)
		}
	}
	// read cmd
	n, err = tcpConn.Read(buf[0:])
	if err != nil || err == io.EOF {
		return
	}
	b = buf[0:n]
	switch b[1] {
	case ConnectCommand:
		TCPProxy(tcpConn, b)
		return
	case AssociateCommand:
		UDPProxy(tcpConn, udpConn, config)
		return
	case BindCommand:
		responseTCP(tcpConn, CommandNotSupported)
		return
	default:
		responseTCP(tcpConn, CommandNotSupported)
		return
	}
}
