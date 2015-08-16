package xhyve

import (
	"fmt"
	"os/exec"
	"strconv"

	"code.google.com/p/go-uuid/uuid"
)

type Config struct {
	CloudConfig    string
	UUID           string
	Version        string
	Channel        string
	CPUs           int
	Memory         int
	Root           string
	XhyvePath      string
	Cmdline        string
	SSHKey         string
	Extra          string
	ImageDirectory string
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

	cmdline := "earlyprintk=serial console=ttyS0 coreos.autologin"
	if cfg.CloudConfig != "" {
		cmdline = fmt.Sprintf("%s cloud-config-url=%s", cmdline, cfg.CloudConfig)
	}
	if cfg.SSHKey != "" {
		cmdline = fmt.Sprintf("%s sshkey=%s", cmdline, cfg.SSHKey)
	}
	if cfg.Root != "" {
		args = append(args,
			"-s", fmt.Sprintf("4:0,virtio-blk,%s", cfg.Root),
		)
		cmdline = fmt.Sprintf("%s root=/dev/vda", cmdline)
	}
	cmdline = fmt.Sprintf("%s %s", cmdline, cfg.Cmdline)

	payload := fmt.Sprintf("%s.%s.coreos_production_pxe", cfg.Channel, cfg.Version)
	vmlinuz := fmt.Sprintf("%s/%s.vmlinuz", cfg.ImageDirectory, payload)
	initrd := fmt.Sprintf("%s/%s_image.cpio.gz", cfg.ImageDirectory, payload)
	firmware := fmt.Sprintf("kexec,%s,%s,%s", vmlinuz, initrd, cmdline)

	args = append(args, "-f", firmware)

	return exec.Command(cfg.XhyvePath, args...), nil
}
