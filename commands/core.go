package commands

import (
	"os"

	"github.com/ecnahc515/core/coreos"

	"github.com/coreos/pkg/capnslog"
	"github.com/ecnahc515/core/xhyve"
	"github.com/spf13/cobra"
)

var (
	plog    = capnslog.NewPackageLogger("github.com/ecnahc515/core", "command")
	CoreCmd = &cobra.Command{
		Use: "core",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cfg      xhyve.Config
	logLevel string
)

func Execute() {
	AddCommands()
	CoreCmd.Execute()
}

func AddCommands() {
	CoreCmd.AddCommand(RunCmd)
	CoreCmd.AddCommand(FetchCmd)
}

func init() {
	CoreCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "level of logging information by package (pkg=level)")
	CoreCmd.PersistentFlags().StringVar(&cfg.ImageDirectory, "image-dir", coreos.DefaultImageDirectory, "Directory of where images are located")
}

func InitializeConfig() {
	rl := capnslog.MustRepoLogger("github.com/ecnahc515/core")
	capnslog.SetFormatter(capnslog.NewStringFormatter(os.Stderr))
	capnslog.SetGlobalLogLevel(capnslog.INFO)

	if logLevel != "" {
		llc, err := rl.ParseLogLevelConfig(logLevel)
		if err != nil {
			plog.Fatal(err)
		}
		rl.SetLogLevel(llc)
		plog.Printf("Setting log level to %s", logLevel)
	}

	if cfg.ImageDirectory == coreos.DefaultImageDirectory {
		err := coreos.CreateImageDirIfNotExist(&cfg)
		if err != nil {
			plog.Errorf("Unable to create default image directory, err: %s", err)
		}
	}
}
