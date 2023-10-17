package xdeb

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func comparePackageChecksums(path string, expected string) error {
	hasher := sha256.New()
	contents, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	hasher.Write(contents)
	actual := hex.EncodeToString(hasher.Sum(nil))

	if actual != expected {
		return fmt.Errorf("checksums don't match: actual=%s expected=%s", actual, expected)
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
		return fmt.Errorf("could not find any XBPS packages to install within '%s'", binpkgs)
	}

	xbps := TrimPathExtension(filepath.Base(files[0]), 2)
	args := []string{}

	if os.Getuid() > 0 {
		args = append(args, "sudo")
	}

	args = append(args, "xbps-install", "-R", "binpkgs", xbps)
	return ExecuteCommand(workdir, args...)
}

func InstallPackage(packageDefinition *XdebPackageDefinition, context *cli.Context) error {
	packageDefinition.Configure(context.String("temp"))

	if packageDefinition.Provider == "localhost" {
		LogMessage("Installing %s from %s", packageDefinition.Name, packageDefinition.FilePath)
	} else if packageDefinition.Provider == "remote" {
		LogMessage("Installing %s from %s", packageDefinition.Name, packageDefinition.Url)
	} else {
		LogMessage(
			"Installing %s from %s @ %s/%s",
			packageDefinition.Name, packageDefinition.Provider, packageDefinition.Distribution, packageDefinition.Component,
		)
	}

	// download if an URL is provided
	if len(packageDefinition.Url) > 0 {
		err := os.RemoveAll(packageDefinition.Path)

		if err != nil {
			return err
		}

		packageDefinition.FilePath, err = DownloadFile(
			filepath.Join(packageDefinition.Path, fmt.Sprintf("%s.deb", packageDefinition.Name)),
			packageDefinition.Url, true, false,
		)

		if err != nil {
			return err
		}
	}

	// compare checksums if available
	if len(packageDefinition.Sha256) > 0 {
		if err := comparePackageChecksums(packageDefinition.FilePath, packageDefinition.Sha256); err != nil {
			return err
		}
	}

	// xdeb convert
	if err := convertPackage(packageDefinition.FilePath, context.String("options")); err != nil {
		return err
	}

	// xbps-install
	if err := installPackage(packageDefinition.FilePath); err != nil {
		return err
	}

	// run post install hooks
	if err := packageDefinition.runPostInstallHooks(); err != nil {
		return err
	}

	// cleanup
	return os.RemoveAll(packageDefinition.Path)
}
