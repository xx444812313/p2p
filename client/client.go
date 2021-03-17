package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

type C struct {
	clientName       string
	remoteServerIP   net.IP
	remoteServerPort int
}

const HAND_SHAKE_MSG = "我是打洞消息"

func NewClient(serverAddr, name string) (*C, error) {
	if len(name) == 0 {
		return nil, errors.New("name 不能为空")
	}
	ip, port := parseAddr(serverAddr)
	return &C{
		clientName:       name,
		remoteServerIP:   net.ParseIP(ip),
		remoteServerPort: port,
	}, nil
}

func (c *C) Run() error {
	srcAddr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 9901,
	}
	dstAddr := &net.UDPAddr{
		IP:   net.ParseIP("192.168.1.102"),
		Port: 9527,
	}
	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if _, err = conn.Write([]byte("hello,I'm new peer:" + c.clientName)); err != nil {
		fmt.Println(err) //写入日志
	}
	data := make([]byte, 1024)
	n, remoteAddr, err := conn.ReadFromUDP(data)
	if err != nil {
		fmt.Printf("error during read: %s", err)
	}
	conn.Close()
	anotherPeerIP, port := parseAddr(string(data[:n]))
	anotherPeer := net.UDPAddr{IP: net.ParseIP(anotherPeerIP), Port: port}
	fmt.Printf("local:%s server:%s another:%s\n", srcAddr, remoteAddr, anotherPeer.String())
	// 打洞
	c.bidirectionHole(srcAddr, &anotherPeer)
	return nil
}

func parseAddr(addr string) (string, int) {
	addrs := strings.Split(addr, ":")
	port, _ := strconv.Atoi(addrs[1])
	return addrs[0], port
}

func (c *C) bidirectionHole(srcAddr *net.UDPAddr, anotherAddr *net.UDPAddr) {
	conn, err := net.DialUDP("udp", srcAddr, anotherAddr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	// 向另一个peer发送一条udp消息(对方peer的nat设备会丢弃该消息,非法来源),用意是在自身的nat设备打开一条可进入的通道,这样对方peer就可以发过来udp消息
	if _, err = conn.Write([]byte(HAND_SHAKE_MSG)); err != nil {
		log.Println("send handshake:", err)
	}
	go func() {
		for {
			time.Sleep(10 * time.Second)
			if _, err = conn.Write([]byte("from [" + c.clientName + "]")); err != nil {
				log.Println("send msg fail", err)
			}
		}
	}()
	for {
		data := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(data)
		if err != nil {
			log.Printf("error during read: %s\n", err)
		} else {
			log.Printf("收到数据:%s\n", data[:n])
		}
	}
}
