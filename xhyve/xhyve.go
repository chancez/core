package xhyve

import (
	"fmt"
	"os/exec"
	"strconv"

	"code.google.com/p/go-uuid/uuid"
)

type Config struct {
	UUID         string
	CPUs         int
	Memory       int
	XhyvePath    string
	Extra        []string
	Disks        []string
	KernelConfig KernelConfig
}

type KernelConfig struct {
	Initrd  string
	Vmlinuz string
	Cmdline string
}

func (cfg Config) Validate() error {
	if cfg.UUID != "" && uuid.Parse(cfg.UUID) == nil {
		return fmt.Errorf("Invalid UUID: %s", cfg.UUID)
	}
	if cfg.CPUs < 1 {
		return fmt.Errorf("Invalid number of CPUs: %s", cfg.CPUs)
	}
	return nil
}

func Command(cfg Config) (*exec.Cmd, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if cfg.XhyvePath == "" {
		cfg.XhyvePath = "xhyve"
	}
	if cfg.UUID == "" {
		cfg.UUID = uuid.New()
	}

	args := []string{
		"-m", fmt.Sprintf("%dM", cfg.Memory),
		"-c", strconv.Itoa(cfg.CPUs),
		"-A",
		"-s", "0:0,hostbridge",
		"-s", "31,lpc",
		"-l", "com1,stdio",
		"-s", "2:0,virtio-net",
		"-U", cfg.UUID,
	}

	for i, diskPath := range cfg.Disks {
		disk := NewDisk(i, diskPath)
		args = append(args, "-s", disk)
	}

	args = append(args, cfg.Extra...)

	firmware := fmt.Sprintf("kexec,%s,%s,%s",
		cfg.KernelConfig.Vmlinuz, cfg.KernelConfig.Initrd, cfg.KernelConfig.Cmdline)
	args = append(args, "-f", firmware)

	return exec.Command(cfg.XhyvePath, args...), nil
}

func NewDisk(pciSlot int, path string) string {
	return fmt.Sprintf("4:%d,virtio-blk,%s", pciSlot, path)
}
