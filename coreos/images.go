package coreos

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

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

type Config struct {
	Version        string
	Channel        string
	Cmdline        string
	SSHKey         string
	CloudConfig    string
	ImageDirectory string
	Root           string
}

func NewKernelConfig(cfg Config) (xhyve.KernelConfig, error) {
	cmdline := "earlyprintk=serial console=ttyS0 coreos.autologin"
	if cfg.SSHKey != "" {
		contents, err := ioutil.ReadFile(cfg.SSHKey)
		if err != nil {
			return xhyve.KernelConfig{}, err
		}
		sshkey := strings.TrimSpace(string(contents))
		cmdline = fmt.Sprintf("%s sshkey=\"%s\"", cmdline, sshkey)
	}
	if cfg.CloudConfig != "" {
		cmdline = fmt.Sprintf("%s cloud-config-url=%s", cmdline, cfg.CloudConfig)
	}
	// TODO: support more disks and don't hardcode the location
	if cfg.Root != "" {
		cmdline = fmt.Sprintf("%s root=/dev/vda", cmdline)
	}
	cmdline = fmt.Sprintf("%s %s", cmdline, cfg.Cmdline)

	image := fmt.Sprintf("%s.%s.coreos_production_pxe", cfg.Channel, cfg.Version)
	vmlinuz := path.Join(cfg.ImageDirectory, fmt.Sprintf("%s.vmlinuz", image))
	initrd := path.Join(cfg.ImageDirectory, fmt.Sprintf("%s_image.cpio.gz", image))
	return xhyve.KernelConfig{
		Vmlinuz: vmlinuz,
		Initrd:  initrd,
		Cmdline: cmdline,
	}, nil
}
