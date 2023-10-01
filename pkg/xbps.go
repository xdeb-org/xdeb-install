package xdeb

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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
	provider := packageDefinition.Provider
	distribution := packageDefinition.Distribution
	component := packageDefinition.Component

	if len(packageDefinition.Path) > 0 {
		// local file
		provider = "localhost"
		distribution = fmt.Sprintf("file:///%s", strings.TrimPrefix(packageDefinition.Path, "/"))
		component = packageDefinition.Path
	}

	if len(provider) == 0 {
		// direct URL
		log.Printf("Installing %s from %s\n", packageDefinition.Name, packageDefinition.Url)
	} else {
		log.Printf(
			"Installing %s from %s @ %s/%s\n",
			packageDefinition.Name, provider, distribution, component,
		)
	}

	path := filepath.Join(context.String("temp"), packageDefinition.Name)

	// prepare xdeb directory
	if err := os.RemoveAll(path); err != nil {
		return err
	}

	fullPath := packageDefinition.Path

	// download if an URL is provided
	if len(packageDefinition.Url) > 0 {
		var err error
		fullPath, err = DownloadFile(path, packageDefinition.Url, true)

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
