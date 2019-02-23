package main

import (
	"log"
	"net"
	"strings"
)

var vendorIdVMware = interfaceMatchRule{[]byte{0x00, 0x50, 0x56}, "VMware"}
var vendorIdVirtualBox = interfaceMatchRule{[]byte{0x0a, 0x00, 0x27}, "VirtualBox"}

type interfaceMatchRule struct {
	macVendorId []byte
	nameKeyword string
}

func (r interfaceMatchRule) match(int net.Interface) bool {
	return r.matchMacAddr(int) || r.matchName(int)
}

func (r interfaceMatchRule) matchName(int net.Interface) bool {
	match := strings.Contains(int.Name, r.nameKeyword)
	if match {
		log.Printf("Name %s belongs to vendor %s\n", int.Name, r.nameKeyword)
	} else {
		log.Printf("Name %s not fits vendor %s\n", int.Name, r.nameKeyword)
	}
	return match
}

func (r interfaceMatchRule) matchMacAddr(int net.Interface) bool {
	if int.HardwareAddr == nil || len(int.HardwareAddr) != 6 {
		log.Printf("cannot understand hardware address %v\n", int.HardwareAddr)
		return false
	}
	for idx, b := range r.macVendorId {
		if int.HardwareAddr[idx] != b {
			log.Printf("MAC address %s do not fits vendor %s\n", int.HardwareAddr.String(), r.nameKeyword)
			return false
		}
	}
	log.Printf("MAC address %s belongs to vendor %s\n", int.HardwareAddr.String(), r.nameKeyword)
	return true
}

func isVirtualNetInterface(int net.Interface) bool {
	return vendorIdVMware.match(int) || vendorIdVirtualBox.match(int)
}
