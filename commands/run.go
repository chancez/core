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

var (
	coreCfg  coreos.Config
	xhyveCfg xhyve.Config
	dryRun   bool
)

func init() {
	// Xhyve specific
	RunCmd.PersistentFlags().StringVar(&xhyveCfg.UUID, "uuid", "", "UUID for the VM. Must be a V4 UUID")
	RunCmd.PersistentFlags().IntVar(&xhyveCfg.CPUs, "cpus", 1, "Number of CPUs to allocate to VM")
	RunCmd.PersistentFlags().IntVar(&xhyveCfg.Memory, "memory", 1024, "Amount of memory in MB to dedicate to VM")
	RunCmd.PersistentFlags().StringVar(&xhyveCfg.XhyvePath, "xhyve", "xhyve", "Path to the xhyve binary")
	RunCmd.PersistentFlags().StringSliceVar(&xhyveCfg.Extra, "extra", []string{}, "Any extra parameters to pass to xhyve")

	// CoreOS specific
	RunCmd.PersistentFlags().StringVar(&coreCfg.Root, "root", "", "Path to disk image to be used as the root disk of the VM")
	RunCmd.PersistentFlags().StringVar(&coreCfg.Version, "version", "", "CoreOS image version")
	RunCmd.PersistentFlags().StringVar(&coreCfg.Channel, "channel", "alpha", "CoreOS image channel")
	RunCmd.PersistentFlags().StringVar(&coreCfg.CloudConfig, "cloud-config", "", "URL or Path to a cloud-config")
	RunCmd.PersistentFlags().StringVar(&coreCfg.Cmdline, "cmdline", "", "Additional kernel cmdline parameters")
	RunCmd.PersistentFlags().StringVar(&coreCfg.SSHKey, "sshkey", "", "Path to ssh public key")
	RunCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Do all of the setup, but do not start the VM")
}

func runXhyve() {
	InitializeConfig()
	if coreCfg.Version == "" {
		var err error
		coreCfg.Version, err = coreos.GetLatestImage(coreCfg.Channel, coreCfg.ImageDirectory)
		if err != nil {
			plog.Fatalf("couldn't find anything to load locally (%s channel). please run `core fetch` first. err: %v", coreCfg.Channel, err)
		}
		plog.Infof("No version specified, using latest local image: CoreOS %s (%s)", coreCfg.Channel, coreCfg.Version)
	}
	kernelCfg, err := coreos.NewKernelConfig(coreCfg)
	if err != nil {
		plog.Fatalf("error creating kernel config: %v", err)
	}
	xhyveCfg.KernelConfig = kernelCfg
	// TODO: support more disks
	if coreCfg.Root != "" {
		xhyveCfg.Disks = []string{coreCfg.Root}
	}
	cmd, err := xhyve.Command(xhyveCfg)
	if err != nil {
		plog.Errorf("error creating command: %v", err)
	}
	if dryRun {
		plog.Infof("%s", strings.Join(cmd.Args, " "))
		return
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	plog.Debugf("executing '%s'", strings.Join(cmd.Args, " "))
	err = cmd.Run()
	if err != nil {
		plog.Errorf("error running xhyve: %v", err)
	}
}
