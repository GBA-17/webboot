package bootiso

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/u-root/u-root/pkg/boot"
	"github.com/u-root/u-root/pkg/boot/kexec"
	"github.com/u-root/u-root/pkg/boot/syslinux"
	"github.com/u-root/u-root/pkg/mount"
	"golang.org/x/sys/unix"
)

// ParseConfigFromISO mounts an iso file to a
// temp dir to get the config options
func ParseConfigFromISO(isoPath string) ([]boot.OSImage, error) {
	tmp, err := ioutil.TempDir("", "mnt")
	if err != nil {
		return nil, fmt.Errorf("Error creating mount dir: %v", err)
	}

	cmd := exec.Command("mount", isoPath, tmp)
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Error mounting ISO: %v, try running as sudo", err)
	}

	configOpts, sysErr := syslinux.ParseLocalConfig(context.Background(), tmp)
	cmd = exec.Command("umount", tmp, "-l")
	umountErr := cmd.Run()

	if sysErr != nil {
		return nil, fmt.Errorf("Error parsing config: %v", sysErr)
	} else if umountErr != nil {
		return nil, fmt.Errorf("Error unmounting ISO: %v", umountErr)
	} else { // success
		return configOpts, nil
	}
}

// BootFromPmem ff
func BootFromPmem(isoPath string, configIndex int) error {
	pmem, err := os.OpenFile("/dev/pmem0", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return (err)
	}

	iso, err := os.Open(isoPath)
	if err != nil {
		return err
	}
	defer iso.Close()

	if _, err := io.Copy(pmem, iso); err != nil {
		return err
	}
	if err = pmem.Close(); err != nil {
		return err
	}

	tmp, err := ioutil.TempDir("", "mnt")
	if err != nil {
		return err
	}

	if _, err := mount.Mount("/dev/pmem0", tmp, "iso9660", "", unix.MS_RDONLY|unix.MS_NOATIME); err != nil {
		return err
	}

	configOpts, err := syslinux.ParseLocalConfig(context.Background(), tmp)
	if err != nil {
		return err
	}

	// Need to convert from boot.OSImage to
	// boot.LinuxImage to edit the Cmdline
	if configIndex < 0 || configIndex >= len(configOpts) {
		return fmt.Errorf("Bad config index")
	}
	image := configOpts[configIndex]
	linuxImage, ok := image.(*boot.LinuxImage)
	if !ok {
		return err
	}

	localCmd, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		return err
	}
	cmdline := strings.TrimSuffix(string(localCmd), "\n") + " " + linuxImage.Cmdline
	linuxImage.Cmdline = cmdline

	fmt.Println(linuxImage)
	if err := linuxImage.Load(true); err != nil {
		return err
	}
	if err := kexec.Reboot(); err != nil {
		return err
	}

	return nil
}
