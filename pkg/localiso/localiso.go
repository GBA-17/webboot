package localiso

import (
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"

	"github.com/u-root/u-root/pkg/boot"
	"github.com/u-root/u-root/pkg/boot/syslinux"
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
		return nil, fmt.Errorf("Error mounting ISO: %v", err)
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
