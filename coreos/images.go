package coreos

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/ecnahc515/core/xhyve"
)

const (
	Vmlinuz               = "coreos_production_pxe.vmlinuz"
	Initrd                = "coreos_production_pxe_image.cpio.gz"
	DefaultImageDirectory = "$HOME/.core/images"
)

var (
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

func CreateImageDirIfNotExist(cfg *xhyve.Config) error {
	if cfg.ImageDirectory == DefaultImageDirectory {
		cfg.ImageDirectory = os.ExpandEnv(cfg.ImageDirectory)
	}
	if _, err := os.Stat(cfg.ImageDirectory); os.IsNotExist(err) {
		plog.Debugf("Image directory %s does not exist, attempting to create it.", cfg.ImageDirectory)
		return os.MkdirAll(cfg.ImageDirectory, 0700)
	}
	return nil
}
