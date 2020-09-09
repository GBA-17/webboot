package bootiso

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/u-root/u-root/pkg/boot"
	"github.com/u-root/u-root/pkg/boot/grub"
	"github.com/u-root/u-root/pkg/boot/kexec"
	"github.com/u-root/u-root/pkg/boot/syslinux"
	"github.com/u-root/u-root/pkg/mount/loop"
	"golang.org/x/sys/unix"
)

// ParseConfigFromISO mounts an iso file to a
// temp dir to get the config options
func ParseConfigFromISO(isoPath string, configType string) ([]boot.OSImage, error) {
	tmp, err := ioutil.TempDir("", "mnt")
	if err != nil {
		return nil, fmt.Errorf("Error creating mount dir: %v", err)
	}
	defer os.RemoveAll(tmp)

	loopdev, err := loop.New(isoPath, "iso9660", "")
	if err != nil {
		return nil, fmt.Errorf("Error creating loop device: %v", err)
	}

	mp, err := loopdev.Mount(tmp, unix.MS_RDONLY|unix.MS_NOATIME)
	if err != nil {
		return nil, fmt.Errorf("Error mounting loop device: %v", err)
	}
	defer mp.Unmount(0)

	configOpts, err := parseConfigFile(tmp, configType)
	if err != nil {
		return nil, fmt.Errorf("Error parsing config: %v", err)
	}

	return configOpts, nil
}

func BootCachedISO(isoPath string, configLabel string, configType string, kernelParams string) error {
	configOpts, err := ParseConfigFromISO(isoPath, configType)
	if err != nil {
		return fmt.Errorf("Error retrieving config options: %v", err)
	}

	osImage := findConfigOptionByLabel(configOpts, configLabel)
	if osImage == nil {
		return fmt.Errorf("Config option with the requested label does not exist")
	}

	// Need to convert from boot.OSImage to boot.LinuxImage to edit the Cmdline
	linuxImage, ok := osImage.(*boot.LinuxImage)
	if !ok {
		return fmt.Errorf("Error converting from boot.OSImage to boot.LinuxImage")
	}

	linuxImage.Cmdline = linuxImage.Cmdline + " " + kernelParams

	if err := linuxImage.Load(true); err != nil {
		return err
	}

	if err := kexec.Reboot(); err != nil {
		return err
	}

	return nil
}

func findConfigOptionByLabel(configOptions []boot.OSImage, configLabel string) boot.OSImage {
	for _, config := range configOptions {
		if config.Label() == configLabel {
			return config
		}
	}
	return nil
}

func parseConfigFile(mountDir string, configType string) ([]boot.OSImage, error) {
	if configType == "syslinux" {
		return syslinux.ParseLocalConfig(context.Background(), mountDir)
	} else if configType == "grub" {
		return grub.ParseLocalConfig(context.Background(), mountDir)
	}

	// If no config type was specified, try both grub and syslinux
	configOpts, err := syslinux.ParseLocalConfig(context.Background(), mountDir)
	if err == nil && len(configOpts) != 0 {
		return configOpts, err
	}
	return grub.ParseLocalConfig(context.Background(), mountDir)
}
