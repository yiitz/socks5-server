package socks5

import "net"

/*
+----+------+----------+------+----------+
 |VER | ULEN | UNAME | PLEN | PASSWD |
 +----+------+----------+------+----------+
 | 1 | 1 | 1 to 255 | 1 | 1 to 255 |
 +----+------+----------+------+----------+
*/
func getUserPwd(conn net.Conn) (user, pwd string) {
	ver := make([]byte, 1)
	n, err := conn.Read(ver)
	if err != nil || n == 0 {
		return "", ""
	}
	if uint(ver[0]) != uint(UserAuthVersion) {
		return "", ""
	}
	ulen := make([]byte, 1)
	n, err = conn.Read(ulen)
	if err != nil || n == 0 {
		return "", ""
	}
	if uint(ulen[0]) < 1 {
		return "", ""
	}
	uname := make([]byte, uint(ulen[0]))
	n, err = conn.Read(uname)
	if err != nil || n == 0 {
		return "", ""
	}
	user = string(uname)

	plen := make([]byte, 1)
	n, err = conn.Read(plen)
	if err != nil || n == 0 {
		return "", ""
	}
	if uint(plen[0]) < 1 {
		return "", ""
	}
	passwd := make([]byte, uint(plen[0]))
	n, err = conn.Read(passwd)
	if err != nil || n == 0 {
		return "", ""
	}
	pwd = string(passwd)
	return user, pwd
}
