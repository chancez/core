package coreos

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/coreos/pkg/capnslog"
)

var plog = capnslog.NewPackageLogger("github.com/ecnahc515/core", "coreos")

const (
	rootUrl = "http://%s.release.core-os.net/amd64-usr/%s"
	Vmlinuz = "coreos_production_pxe.vmlinuz"
	Initrd  = "coreos_production_pxe_image.cpio.gz"
)

func getVersionURL(channel string) string {
	return fmt.Sprintf(rootUrl, channel, "current/version.txt")
}

func getDownloadURL(channel, version string) string {
	return fmt.Sprintf(rootUrl, channel, version)
}

func getFileURL(channel, version, file string) string {
	root := getDownloadURL(channel, version)
	return fmt.Sprintf("%s/%s", root, file)
}

func GetVersionID(channel string) (string, error) {
	const wantKey = "COREOS_VERSION_ID"
	url := getVersionURL(channel)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	r := bufio.NewReader(resp.Body)
	line, err := r.ReadString('\n')
	for err == nil {
		items := strings.SplitN(strings.TrimSpace(line), "=", 2)
		key, value := items[0], items[1]
		if key == wantKey {
			return value, nil
		}
		line, err = r.ReadString('\n')
	}
	return "", fmt.Errorf("Unable to find %s in response", wantKey)
}

type Downloader struct {
	stop           chan struct{}
	cleanupDone    chan struct{}
	Channel        string
	Version        string
	ImageDirectory string
	tmpDir         string
}

func NewDownloader(channel, version, imageDirectory string) *Downloader {
	return &Downloader{
		stop:           make(chan struct{}),
		cleanupDone:    make(chan struct{}),
		Channel:        channel,
		Version:        version,
		ImageDirectory: imageDirectory,
	}
}

func (d *Downloader) Download(file string) (err error) {
	endFile := fmt.Sprintf("%s.%s.%s", d.Channel, d.Version, file)
	loc := path.Join(d.ImageDirectory, endFile)
	// check if we've already downloaded this
	if _, err := os.Stat(loc); err == nil {
		plog.Infof("Found cached %s (%s/%s)", file, d.Channel, d.Version)
		return nil
	}

	errChan := make(chan error)

	if d.tmpDir == "" {
		// create a directory to temporarily hold our downloads
		d.tmpDir, err = ioutil.TempDir("", "coreos-install")
		if err != nil {
			return
		}
	}

	var res DownloadResult
	go func() {
		res, err = Download(d.Channel, d.Version, file, d.tmpDir)
		errChan <- err
	}()

	// Wait for the download to finish or for it to be canceled
	select {
	case err := <-errChan:
		if err != nil {
			return err
		}
	case <-d.stop:
		os.RemoveAll(d.tmpDir)
		d.cleanupDone <- struct{}{}
		return errors.New("Recieved cancel signal")
	}

	defer os.RemoveAll(d.tmpDir)

	// move the files into the final location
	var errs []error
	err = os.Rename(res.FileLocation, path.Join(d.ImageDirectory, endFile))
	if err != nil {
		errs = append(errs, err)
	}
	err = os.Rename(res.SignatureLocation, path.Join(d.ImageDirectory, endFile+".sig"))
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		err = fmt.Errorf("Encountered the following errors when moving files: %v", errs)
	}

	return
}

func (d *Downloader) Cleanup() {
	d.stop <- struct{}{}
	// Wait until cleanup is finished
	<-d.cleanupDone
}

type DownloadResult struct {
	FileLocation      string
	SignatureLocation string
}

func Download(channel, version, file, imageDirectory string) (res DownloadResult, err error) {
	plog.Infof("Downloading %s, Channel: %s, Version: %s", file, channel, version)

	// The actual file
	url := getFileURL(channel, version, file)
	res.FileLocation = path.Join(imageDirectory, file)

	// The signature file for our download
	sigUrl := url + ".sig"
	res.SignatureLocation = path.Join(imageDirectory, file+".sig")

	// Download both the file and it's signature
	plog.Debugf("Downloading %s to %s", url, res.FileLocation)
	err = get(url, res.FileLocation)
	if err != nil {
		return
	}

	plog.Debugf("Downloading %s to %s", sigUrl, res.SignatureLocation)
	err = get(sigUrl, res.SignatureLocation)
	if err != nil {
		return
	}
	// TODO: verify signatures
	err = verify(res.FileLocation, res.SignatureLocation)
	if err != nil {
		return
	}
	plog.Infof("Signature verified!")
	return
}

func get(url, outputFile string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
