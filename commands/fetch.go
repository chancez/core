package commands

import (
	"os"
	"os/signal"

	"github.com/ecnahc515/core/coreos"
	"github.com/spf13/cobra"
)

const defaultChannel = "alpha"

var FetchCmd = &cobra.Command{
	Use:   "fetch [channel] [version]",
	Short: "Download a CoreOS image",
	Long:  "Downloads a CoreOS image from release.core-os.net, storing it locally. Defaults to alpha channel if unspecified.",
	Run: func(cmd *cobra.Command, args []string) {
		fetchImage(cmd, args)
	},
}

func fetchImage(cmd *cobra.Command, args []string) {
	InitializeConfig()
	channel := defaultChannel
	if len(args) > 1 {
		channel = args[0]
	}
	var (
		version string
		err     error
	)
	if len(args) == 2 {
		version = args[1]
	} else {
		version, err = coreos.GetVersionID(defaultChannel)
		if err != nil {
			plog.Fatalf("Unable to get version for channel %s. err: %v", channel, err)
		}
	}
	plog.Debugf("Channel: %s, Version: %s\n", channel, version)

	downloader := coreos.NewDownloader(channel, version, cfg.ImageDirectory)

	// Setup signal handlers so we properly cleanup when we get a signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	go func() {
		<-sigChan
		downloader.Stop()
		return
	}()

	err = downloader.Download(coreos.Vmlinuz)
	if err != nil {
		downloader.Cleanup()
		plog.Fatalf("Error downloading %s to %s. err: %v", coreos.Vmlinuz, cfg.ImageDirectory, err)
		return
	}
	err = downloader.Download(coreos.Initrd)
	if err != nil {
		downloader.Cleanup()
		plog.Fatalf("Error downloading %s to %s. err: %v", coreos.Vmlinuz, cfg.ImageDirectory, err)
		return
	}
	plog.Infof("Successfully downloaded CoreOS %s (%s)", channel, version)
}
