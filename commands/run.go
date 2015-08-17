package commands

import (
	"os"
	"strings"

	"github.com/ecnahc515/core/coreos"
	"github.com/ecnahc515/core/xhyve"
	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Start a CoreOS VM",
	Long:  "Start a CoreOS VM using Xhyve",
	Run: func(cmd *cobra.Command, args []string) {
		runXhyve()
	},
}

func init() {
	RunCmd.PersistentFlags().StringVar(&cfg.CloudConfig, "cloud-config", "", "URL or Path to a cloud-config")
	RunCmd.PersistentFlags().StringVar(&cfg.UUID, "uuid", "", "UUID for the VM. Must be a V4 UUID")
	RunCmd.PersistentFlags().StringVar(&cfg.Version, "version", "", "CoreOS image version")
	RunCmd.PersistentFlags().StringVar(&cfg.Channel, "channel", "alpha", "CoreOS image channel")
	RunCmd.PersistentFlags().IntVar(&cfg.CPUs, "cpus", 1, "Number of CPUs to allocate to VM")
	RunCmd.PersistentFlags().IntVar(&cfg.Memory, "memory", 1024, "Amount of memory in MB to dedicate to VM")
	RunCmd.PersistentFlags().StringVar(&cfg.Root, "root", "", "Path to disk image to be used as the root disk of the VM")
	RunCmd.PersistentFlags().StringVar(&cfg.XhyvePath, "xhyve", "xhyve", "Path to the xhyve binary")
	RunCmd.PersistentFlags().StringVar(&cfg.Cmdline, "cmdline", "", "Additional kernel cmdline parameters")
	RunCmd.PersistentFlags().StringVar(&cfg.SSHKey, "sshkey", "", "Text version of ssh public key or if it's an absolute path, the file will be read.")
	RunCmd.PersistentFlags().StringVar(&cfg.Extra, "extra", "", "Any extra parameters to pass to xhyve")
}

func runXhyve() {
	InitializeConfig()
	if cfg.Version == "" {
		var err error
		cfg.Version, err = coreos.GetLatestImage(cfg.Channel, cfg.ImageDirectory)
		if err != nil {
			plog.Fatalf("couldn't find anything to load locally (%s channel). please run `core fetch` first. err: %v", cfg.Channel, err)
		}
		plog.Infof("No version specified, using latest local image: CoreOS %s (%s)", cfg.Channel, cfg.Version)
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
}
