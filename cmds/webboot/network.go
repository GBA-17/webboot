package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/u-root/NiChrome/pkg/wifi"
	"github.com/u-root/webboot/pkg/menu"
	"github.com/vishvananda/netlink"
)

func connected() bool {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	if _, err := client.Get("http://google.com"); err != nil {
		return false
	}
	return true
}

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

func setupNetwork(uiEvents <-chan ui.Event) (bool, error) {
	iface, err := selectNetworkInterface(uiEvents)
	if err != nil {
		return false, err
	}

	err = selectWirelessNetwork(uiEvents, iface)
	if err != nil {
		return false, nil
	}

	return true, nil
}

func selectNetworkInterface(uiEvents <-chan ui.Event) (string, error) {
	interfaces, err := interfaceNames()
	if err != nil {
		return "", err
	}

	entries := []menu.Entry{}
	for _, iface := range interfaces {
		entries = append(entries, &Interface{label: iface})
	}

	for {
		entry, err := menu.DisplayMenu("Network Interfaces", "Choose an option", entries, uiEvents)
		if err != nil {
			return "", err
		}

		if !interfaceIsWireless(entry.Label()) {
			menu.DisplayResult([]string{"Only wireless network interfaces are supported."}, uiEvents)
		} else {
			return entry.Label(), nil
		}
	}
}

func selectWirelessNetwork(uiEvents <-chan ui.Event, iface string) error {
	worker, err := wifi.NewIWLWorker(iface)
	if err != nil {
		return err
	}

	for {
		networks, err := worker.Scan()
		if err != nil {
			return err
		}

		entries := []menu.Entry{}
		for _, network := range networks {
			entries = append(entries, &Network{info: network})
		}

		entry, err := menu.DisplayMenu("Wireless Networks", "Choose an option", entries, uiEvents)
		if err != nil {
			return err
		}

		network, ok := entry.(*Network)
		if !ok {
			return fmt.Errorf("Bad menu entry.")
		}

		var setupParams = []string{network.info.Essid}
		authSuite := network.info.AuthSuite
		if authSuite == wifi.NotSupportedProto {
			menu.DisplayResult([]string{"Security protocol is not supported."}, uiEvents)
			continue
		} else if authSuite == wifi.WpaPsk || authSuite == wifi.WpaEap {
			credentials, err := enterCredentials(uiEvents, authSuite)
			if err != nil {
				return err
			}
			setupParams = append(setupParams, credentials...)
		}

		if err := worker.Connect(setupParams...); err != nil {
			menu.DisplayResult([]string{err.Error()}, uiEvents)
			continue
		}
		return nil
	}
}

func enterCredentials(uiEvents <-chan ui.Event, authSuite wifi.SecProto) ([]string, error) {
	var credentials []string
	pass, err := menu.NewInputWindow("Enter password:", menu.AlwaysValid, uiEvents)
	if err != nil {
		return nil, err
	}

	credentials = append(credentials, pass)
	if authSuite == wifi.WpaPsk {
		return credentials, nil
	}

	identity, err := menu.NewInputWindow("Enter identity:", menu.AlwaysValid, uiEvents)
	if err != nil {
		return nil, err
	}
	credentials = append(credentials, identity)
	return credentials, nil
}
