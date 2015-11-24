package main

import (
	"fmt"
	"net"
)

func main() {
	socket, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 8089,
	})
	if err != nil {
		fmt.Println("监听失败!", err)

		return
	}

	defer socket.Close()

	for {
		data := make([]byte, 4096)
		read, remoteAddr, err := socket.ReadFromUDP(data)
		if err != nil {
			fmt.Println("读取数据失败!", err)
			continue
		}

		fmt.Println(read, remoteAddr)
		fmt.Printf("%s\n\n", data)

	}

}
