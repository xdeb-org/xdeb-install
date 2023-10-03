package xdeb

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func PackageDefinitionWithMetadata(packageDefinition *XdebPackageDefinition, path string) *XdebPackageDefinition {
	distPath := filepath.Dir(path)

	packageObject := *packageDefinition
	packageObject.Component = TrimPathExtension(filepath.Base(path))
	packageObject.Distribution = filepath.Base(distPath)
	packageObject.Provider = filepath.Base(filepath.Dir(distPath))

	return &packageObject
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
	xbps = TrimPathExtension(xbps)
	xbps = TrimPathExtension(xbps)

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
		if len(packageDefinition.Url) > 0 {
			// direct URL
			LogMessage("Installing %s from %s", packageDefinition.Name, packageDefinition.Url)
		} else {
			LogMessage("Installing %s from %s", packageDefinition.Name, packageDefinition.FilePath)
		}
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

		packageDefinition.FilePath, err = DownloadFile(packageDefinition.Path, packageDefinition.Url, true, false)

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

	// cleanup
	return os.RemoveAll(packageDefinition.Path)
}
