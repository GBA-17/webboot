package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/vishvananda/netlink"
)

func interfaceNames() ([]string, error) {
	interfaces, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}

	var ifNames []string
	for _, iface := range interfaces {
		ifNames = append(ifNames, iface.Attrs().Name)
	}
	return ifNames, nil
}

func interfaceIsWireless(ifname string) bool {
	devPath := fmt.Sprintf("/sys/class/net/%s/wireless", ifname)
	if _, err := os.Stat(devPath); err != nil {
		return false
	}
	return true
}

func connected() bool {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	if _, err := client.Get("http://google.com"); err != nil {
		return false
	}
	return true
}
