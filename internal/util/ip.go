package util

import "net"

func GetSelfIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:1")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	addr := conn.LocalAddr().(*net.UDPAddr)
	return addr.IP, nil
}
