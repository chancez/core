package coreos

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"sort"
)

var (
	Vmlinuz          = "coreos_production_pxe.vmlinuz"
	Initrd           = "coreos_production_pxe_image.cpio.gz"
	re               = regexp.MustCompile("^alpha.([0-9.]+).coreos_production_pxe.vmlinuz$")
	ErrNoLocalImages = errors.New("no local image files")
)

func GetLatestImage(channel, imageDirectory string) (string, error) {
	pat := fmt.Sprintf("%s.*.vmlinuz", channel)
	names, err := filepath.Glob(path.Join(imageDirectory, pat))
	if err != nil {
		return "", err
	}
	if len(names) == 0 {
		return "", ErrNoLocalImages
	}
	sort.Strings(names)
	name := path.Base(names[len(names)-1])
	matches := re.FindStringSubmatch(name)
	if len(matches) != 2 {
		return "", ErrNoLocalImages
	}
	return matches[1], nil
}
