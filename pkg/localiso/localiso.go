package localiso

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/u-root/u-root/pkg/boot"
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

	if _, err := mount.Mount(isoPath, tmp, "iso9660", "--bind", unix.MS_RDONLY|unix.MS_NOATIME); err != nil {
		return nil, fmt.Errorf("Error mounting ISO: %v", err)
	}

	configOpts, sysErr := syslinux.ParseLocalConfig(context.Background(), tmp)
	umountErr := mount.Unmount(tmp, true, false) //force umount
	if sysErr != nil {
		return nil, fmt.Errorf("Error parsing config: %v", sysErr)
	} else if umountErr != nil {
		return nil, fmt.Errorf("Error unmounting ISO: %v", umountErr)
	} else { // success
		return configOpts, nil
	}
}
