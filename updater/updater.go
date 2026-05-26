package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"

	"github.com/minio/selfupdate"
)

// CurrentVersion is the current version of the application.
// This should be updated for each release, or injected via ldflags.
var CurrentVersion = "v0.0.0"

const (
	owner = "yiweichiu"
	repo  = "AyayaBot"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type ReleaseInfo struct {
	Version     string
	DownloadURL string
}

// CheckUpdate checks if a new version is available on GitHub.
func CheckUpdate() (*ReleaseInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to check update: %s", resp.Status)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	if release.TagName == CurrentVersion {
		return nil, nil // No update available
	}

	// Find the asset for the current platform
	// Example names: ayayabot-windows-amd64.zip, ayayabot-darwin-arm64.tar.gz
	targetSuffix := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, targetSuffix) {
			return &ReleaseInfo{
				Version:     release.TagName,
				DownloadURL: asset.BrowserDownloadURL,
			}, nil
		}
	}

	return nil, fmt.Errorf("no suitable asset found for %s", targetSuffix)
}

// DoUpdate downloads and applies the update.
func DoUpdate(downloadURL string) error {
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download update: %s", resp.Status)
	}

	var binaryReader io.Reader

	if strings.HasSuffix(downloadURL, ".zip") {
		// Handle ZIP (Windows)
		// We need the entire body to open a zip reader, so we read it to a temp buffer or use a temp file.
		// For simplicity in memory, we can use a temp file.
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		importReader := strings.NewReader(string(bodyBytes))
		zipReader, err := zip.NewReader(importReader, int64(len(bodyBytes)))
		if err != nil {
			return err
		}

		found := false
		for _, f := range zipReader.File {
			if f.Name == "ayayabot.exe" || f.Name == "ayayabot" {
				rc, err := f.Open()
				if err != nil {
					return err
				}
				defer rc.Close()
				binaryReader = rc
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("binary not found in zip")
		}
	} else if strings.HasSuffix(downloadURL, ".tar.gz") {
		// Handle Tar.GZ (macOS)
		gzr, err := gzip.NewReader(resp.Body)
		if err != nil {
			return err
		}
		defer gzr.Close()

		tr := tar.NewReader(gzr)
		found := false
		for {
			header, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			if header.Typeflag == tar.TypeReg && (header.Name == "ayayabot" || header.Name == "ayayabot.exe") {
				binaryReader = tr
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("binary not found in tar.gz")
		}
	} else {
		// Assume direct binary
		binaryReader = resp.Body
	}

	err = selfupdate.Apply(binaryReader, selfupdate.Options{})
	if err != nil {
		if rerr := selfupdate.RollbackError(err); rerr != nil {
			return fmt.Errorf("update failed and rollback failed: %v", rerr)
		}
		return err
	}

	return nil
}
