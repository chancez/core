package commands

import (
	"os"
	"strings"

	"github.com/ecnahc515/core/xhyve"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start a CoreOS VM",
	Long:  "Start a CoreOS VM using Xhyve",
	Run: func(cmd *cobra.Command, args []string) {
		runXhyve()
	},
}

func init() {
	runCmd.PersistentFlags().StringVar(&cfg.CloudConfig, "cloud-config", "", "URL or Path to a cloud-config")
	runCmd.PersistentFlags().StringVar(&cfg.UUID, "uuid", "", "UUID for the VM. Must be a V4 UUID")
	runCmd.PersistentFlags().StringVar(&cfg.Version, "version", "773.1.0", "CoreOS image version")
	runCmd.PersistentFlags().StringVar(&cfg.Channel, "channel", "alpha", "CoreOS image channel")
	runCmd.PersistentFlags().IntVar(&cfg.CPUs, "cpus", 1, "Number of CPUs to allocate to VM")
	runCmd.PersistentFlags().IntVar(&cfg.Memory, "memory", 1024, "Amount of memory in MB to dedicate to VM")
	runCmd.PersistentFlags().StringVar(&cfg.Root, "root", "", "Path to disk image to be used as the root disk of the VM")
	runCmd.PersistentFlags().StringVar(&cfg.XhyvePath, "xhyve", "xhyve", "Path to the xhyve binary")
	runCmd.PersistentFlags().StringVar(&cfg.Cmdline, "cmdline", "", "Additional kernel cmdline parameters")
	runCmd.PersistentFlags().StringVar(&cfg.SSHKey, "sshkey", "", "Text version of ssh public key or if it's an absolute path, the file will be read.")
	runCmd.PersistentFlags().StringVar(&cfg.Extra, "extra", "", "Any extra parameters to pass to xhyve")
}

func runXhyve() {
	InitializeConfig()
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
