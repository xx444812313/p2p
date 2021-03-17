package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

type S struct {
	port int
}

func NewServer(port int) *S {
	return &S{
		port: port,
	}
}

func (s *S) Run() error {
	udpListener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: s.port})
	if err != nil {
		fmt.Println(err)
		return err
	}
	log.Printf("本地地址: <%s> \n", udpListener.LocalAddr().String())
	peers := make([]net.UDPAddr, 0, 2)
	data := make([]byte, 1024)
	for {
		n, remoteAddr, err := udpListener.ReadFromUDP(data)
		if err != nil {
			fmt.Printf("error during read: %s", err)
			time.Sleep(10 * time.Second)
			continue
		}
		log.Printf("<%s> %s\n", remoteAddr.String(), data[:n])
		peers = append(peers, *remoteAddr)
		if len(peers) == 2 {
			log.Printf("进行UDP打洞,建立 %s <--> %s 的连接\n", peers[0].String(), peers[1].String())
			udpListener.WriteToUDP([]byte(peers[1].String()), &peers[0])
			udpListener.WriteToUDP([]byte(peers[0].String()), &peers[1])
			time.Sleep(10 * time.Second)
			log.Println("中转服务器退出,仍不影响peers间通信")
			break
		}
	}
	return nil
}

func main() {
	s := NewServer(9537)
	s.Run()
}
