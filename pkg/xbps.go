package xdeb

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

func PackageDefinitionWithMetadata(packageDefinition *XdebPackageDefinition, path string) *XdebPackageDefinition {
	distPath := filepath.Dir(path)

	packageObject := *packageDefinition
	packageObject.Component = filepath.Base(strings.TrimSuffix(path, filepath.Ext(path)))
	packageObject.Distribution = filepath.Base(distPath)
	packageObject.Provider = filepath.Base(filepath.Dir(distPath))

	return &packageObject
}

func downloadPackage(path string, url string) (string, error) {
	client := &http.Client{}
	resp, err := client.Get(url)

	if err != nil {
		return "", fmt.Errorf("Could not download file %s", url)
	}

	finalUrl := resp.Request.URL.String()
	resp, err = client.Get(finalUrl)

	if err != nil {
		return "", fmt.Errorf("Could not download file %s", url)
	}

	defer resp.Body.Close()

	err = os.MkdirAll(path, os.ModePerm)

	if err != nil {
		return "", fmt.Errorf("Could not create path %s", path)
	}

	fullPath := filepath.Join(path, filepath.Base(finalUrl))
	out, err := os.Create(fullPath)

	if err != nil {
		return "", fmt.Errorf("Could not create file %s", fullPath)
	}

	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return fullPath, err
}

func comparePackageChecksums(path string, expected string) error {
	hasher := sha256.New()
	contents, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	hasher.Write(contents)
	actual := hex.EncodeToString(hasher.Sum(nil))

	if actual != expected {
		return fmt.Errorf("Checksums don't match: actual=%s expected=%s", actual, expected)
	}

	return nil
}

func installPackage(path string) error {
	workdir := filepath.Dir(path)
	binpkgs := filepath.Join(workdir, "binpkgs")

	files, err := filepath.Glob(filepath.Join(binpkgs, "*.xbps"))

	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("Could not find any XBPS packages to install within %s.", binpkgs)
	}

	xbps := filepath.Base(files[0])
	xbps = strings.TrimSuffix(xbps, filepath.Ext(xbps))
	xbps = strings.TrimSuffix(xbps, filepath.Ext(xbps))

	args := []string{}

	if os.Getuid() > 0 {
		args = append(args, "sudo")
	}

	args = append(args, "xbps-install", "-R", "binpkgs", "-y", xbps)
	return executeCommand(workdir, args...)
}

func InstallPackage(packageDefinition *XdebPackageDefinition, context *cli.Context) error {
	log.Printf(
		"Installing %s from %s @ %s/%s\n",
		packageDefinition.Name, packageDefinition.Provider,
		packageDefinition.Distribution, packageDefinition.Component,
	)

	path := filepath.Join(context.String("temp"), packageDefinition.Name)
	fullPath := packageDefinition.Path

	// download if an URL is provided
	if len(packageDefinition.Url) > 0 {
		var err error
		fullPath, err = downloadPackage(path, packageDefinition.Url)

		if err != nil {
			return err
		}
	}

	// compare checksums if available
	if len(packageDefinition.Sha256) > 0 {
		if err := comparePackageChecksums(fullPath, packageDefinition.Sha256); err != nil {
			return err
		}
	}

	// xdeb convert
	if err := convertPackage(fullPath, context.String("options")); err != nil {
		return err
	}

	// xbps-install
	if err := installPackage(fullPath); err != nil {
		return err
	}

	// cleanup
	return os.RemoveAll(path)
}
