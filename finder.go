package main

import (
	"log"
	"net"
	"os"
	"sync"
)

var _, isAgent = os.LookupEnv("VM_FINDER_AGENT")

func main() {
	senders := &sync.WaitGroup{}
	listeners := &sync.WaitGroup{}

	if isAgent {
		log.Printf("running as agent\n")
		listeners.Add(1)
		go listenDiscoveryPacket(nil, net.IPv4bcast, listeners)
	} else {
		setupLogFile()

		ifs, _ := net.Interfaces()
		for idx, intf := range ifs {
			if (intf.Flags&net.FlagUp == 0) || (intf.Flags&net.FlagLoopback != 0) {
				log.Printf("ignore interface: %v %v\n", idx, intf)
				continue
			}
			log.Printf("interface: %v %v\n", idx, intf)

			if !isAgent && !isVirtualNetInterface(intf) {
				continue
			}
			addrs, err := intf.Addrs()
			if err != nil {
				log.Printf("cannot obtain address of %v: %v\n", intf, err)
			}
			for jdx, addr := range addrs {
				if ip, _, err := net.ParseCIDR(addr.String()); err != nil {
					log.Printf("    unparsable addr %d %v: %s\n", jdx, addr, err)
				} else if ip.To4() != nil {
					log.Printf("    IPv4 addr %d %v: %s\n", jdx, addr, ip.To4())
					listeners.Add(1)
					go listenResponsePacket(&intf, ip.To4(), listeners)
					senders.Add(1)
					go sendPacket(ip.To4(), nil, senders)
				} else if ip.To16() != nil {
					log.Printf("    IPv6 addr %d %v: %s\n", jdx, addr, ip.To16())
				} else {
					log.Printf("    unknown addr version %d %v: %s\n", jdx, addr, ip)
				}
			}
		}

		senders.Wait()
		log.Printf("all discovery packets sended\n")
	}

	listeners.Wait()
	log.Printf("no more vm found\n")
}

func setupLogFile() {
	logOutput, err := os.OpenFile("vm-finder.log", os.O_APPEND|os.O_CREATE, 0666)
	if err == nil {
		log.SetOutput(logOutput)
	} else {
		log.Printf("cannot open log file, use stderr (%s)\n", err)
	}
}
