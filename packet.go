package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const vmDiscoverPort = ":33991"
const vmResponsePort = ":33992"

var discoveryTargetAddr, _ = net.ResolveUDPAddr("udp", "255.255.255.255"+vmDiscoverPort)
var hostname, _ = os.Hostname()

type packetInfo struct {
	Hostname  string
	IPAddress string
}

type packetDir uint

const (
	cDirToVm packetDir = iota
	cDirToHost
)

func sendPacket(source net.IP, raddr *net.UDPAddr, wait *sync.WaitGroup) {
	defer wait.Done()
	addr := source.String() + ":0"
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Printf("cannot get resolve udp addr %s: %s\n", addr, err)
		return
	}
	if raddr == nil {
		raddr = discoveryTargetAddr
	}
	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		log.Printf("cannot connect to udp addr %s from %s: %s\n", discoveryTargetAddr, laddr, err)
		return
	}
	defer conn.Close()

	packet := packetInfo{
		Hostname:  hostname,
		IPAddress: source.String(),
	}
	out, err := json.Marshal(packet)
	if err != nil {
		log.Printf("internal error, unmarshable packet %v: %s\n", packet, err)
		return
	}

	if cnt, err := conn.Write(out); err != nil {
		log.Printf("cannot send udp packet from %s: %s\n", conn.LocalAddr(), err)
	} else {
		log.Printf("%d bytes write to %s from %s\n", cnt, discoveryTargetAddr, conn.LocalAddr())
	}
}

func listenResponsePacket(intf *net.Interface, ip net.IP, wait *sync.WaitGroup) {
	listenUDPPacket(intf, ip, wait, cDirToHost)
}

func listenDiscoveryPacket(intf *net.Interface, ip net.IP, wait *sync.WaitGroup) {
	listenUDPPacket(intf, ip, wait, cDirToVm)
}

func listenUDPPacket(intf *net.Interface, ip net.IP, wait *sync.WaitGroup, dir packetDir) {
	defer wait.Done()

	var addr string
	switch dir {
	case cDirToHost:
		addr = ip.String() + vmResponsePort
	case cDirToVm:
		addr = "255.255.255.255" + vmDiscoverPort
	}
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Printf("cannot get resolve udp addr %s: %s\n", addr, err)
		return
	}
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		log.Printf("cannot listen on udp addr %s: %s\n", laddr, err)
		return
	} else {
		log.Printf("%v listening on %s\n", conn, laddr)
	}

	defer conn.Close()

	var buffer [100]byte

	for {
		if dir == cDirToHost {
			if err = conn.SetReadDeadline(time.Now().Add(15 * time.Second)); err != nil {
				log.Printf("set read timeout failed: %s\n", err)
				break
			}
		}

		cnt, raddr, err := conn.ReadFromUDP(buffer[:])
		if err != nil && strings.Contains(strings.ToLower(err.Error()), "timeout") {
			break
		} else if err != nil {
			log.Printf("read from udp connection error: %s\n", err)
			break
		} else if raddr.IP.Equal(ip) {
			log.Printf("%d bytes received from %s (loopback, ignored)\n", cnt, raddr)
		} else {
			log.Printf("%d bytes received from %s\n", cnt, raddr)
			log.Printf("    -> %v %s", buffer[:cnt], string(buffer[:cnt]))
			var pkt packetInfo
			if err := json.Unmarshal(buffer[:cnt], &pkt); err != nil {
				log.Printf("cannot parse packet: %s\n", err)
			} else {
				log.Printf("packet: %v\n", pkt)
			}
			if dir == cDirToVm {
				log.Printf("going to reply\n")
				wait.Add(1)
				sendPacket(net.IPv4zero, &net.UDPAddr{IP: raddr.IP, Port: 33992}, wait)
			} else {
				fmt.Printf("%s\t%s\n", raddr.IP, pkt.Hostname)
			}
		}
	}
}
