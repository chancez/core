package main

import (
	"flag"
	"os"
	"strings"

	"github.com/coreos/pkg/capnslog"
	"github.com/coreos/pkg/flagutil"
	"github.com/ecnahc515/core/xhyve"
)

var cfg xhyve.Config

func main() {
	plog := capnslog.NewPackageLogger("github.com/ecnahc515/core", "main")
	rl := capnslog.MustRepoLogger("github.com/ecnahc515/core")
	capnslog.SetFormatter(capnslog.NewStringFormatter(os.Stderr))
	capnslog.SetGlobalLogLevel(capnslog.INFO)

	fs := flag.NewFlagSet("corex", flag.ExitOnError)

	fs.StringVar(&cfg.CloudConfig, "cloud-config", "", "URL or Path to a cloud-config")
	fs.StringVar(&cfg.UUID, "uuid", "", "UUID for the VM. Must be a V4 UUID")
	fs.StringVar(&cfg.Version, "version", "773.1.0", "CoreOS image version")
	fs.StringVar(&cfg.Channel, "channel", "alpha", "CoreOS image channel")
	fs.StringVar(&cfg.ImageDirectory, "image-dir", "imgs", "Directory of where images are located")
	fs.IntVar(&cfg.CPUs, "cpus", 1, "Number of CPUs to allocate to VM")
	fs.IntVar(&cfg.Memory, "memory", 1024, "Amount of memory in MB to dedicate to VM")
	fs.StringVar(&cfg.Root, "root", "", "Path to disk image to be used as the root disk of the VM")
	fs.StringVar(&cfg.XhyvePath, "xhyve", "xhyve", "Path to the xhyve binary")
	fs.StringVar(&cfg.Cmdline, "cmdline", "", "Additional kernel cmdline parameters")
	fs.StringVar(&cfg.SSHKey, "sshkey", "", "Text version of ssh public key or if it's an absolute path, the file will be read.")
	fs.StringVar(&cfg.Extra, "extra", "", "Any extra parameters to pass to xhyve")

	logLevel := fs.String("log-level", "", "level of logging information by package (pkg=level)")

	if err := fs.Parse(os.Args[1:]); err != nil {
		plog.Printf(err.Error())
		os.Exit(1)
	}

	if err := flagutil.SetFlagsFromEnv(fs, "COREX"); err != nil {
		plog.Printf(err.Error())
		os.Exit(1)
	}

	if *logLevel != "" {
		llc, err := rl.ParseLogLevelConfig(*logLevel)
		if err != nil {
			plog.Fatal(err)
		}
		rl.SetLogLevel(llc)
		plog.Infof("Setting log level to %s", *logLevel)
	}

	cmd, err := xhyve.Command(cfg)
	if err != nil {
		plog.Errorf("error creating command: %v", err)
	}
	plog.Debugf("executing '%s'", strings.Join(cmd.Args, " "))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		plog.Errorf("error running xhyve: %v", err)
	}
	os.Exit(0)
}
