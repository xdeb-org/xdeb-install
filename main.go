package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/adrg/xdg"
	"github.com/ulikunitz/xz"
	"github.com/urfave/cli/v2"
	"github.com/yargevad/filepathx"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"
)

const APPLICATION_NAME = "xdeb-install"
const XDEB_URL = "https://github.com/toluschr/xdeb/releases"

var ARCHITECTURE_MAP = map[string]string{
	"amd64": "x86_64",
	//"386":   "i386",
}

func getXdebPath() string {
	xdebPath, err := exec.LookPath("xdeb")

	if err != nil {
		log.Fatalf("Package xdeb not found. Please install from %s.", XDEB_URL)
	}

	log.Printf("Package xdeb found: %s", xdebPath)
	return xdebPath
}

func architecturePath(prefix ...string) string {
	arch, ok := ARCHITECTURE_MAP[runtime.GOARCH]

	if !ok {
		log.Fatalf("Architecture %s not supported (yet).", runtime.GOARCH)
	}

	paths := prefix
	paths = append(paths, "repositories", arch, "apt")

	return filepath.Join(paths...)
}

func pathPrefix() string {
	return architecturePath(xdg.ConfigHome, APPLICATION_NAME)
}

func aptProviders() []string {
	entries, err := os.ReadDir(pathPrefix())

	if err != nil {
		log.Fatalf("APT providers not installed. This should never happen.")
	}

	subdirs := []string{}

	for _, entry := range entries {
		if entry.IsDir() {
			subdirs = append(subdirs, entry.Name())
		}
	}

	return subdirs
}

func providerDistributions(provider string) []string {
	entriesPath := filepath.Join(pathPrefix(), provider)
	entries, err := os.ReadDir(entriesPath)

	if err != nil {
		log.Fatalf("Provider distributions not installed. This should never happen.")
	}

	subdirs := []string{}

	for _, entry := range entries {
		if entry.IsDir() {
			subdirs = append(subdirs, entry.Name())
		}
	}

	return subdirs
}

type XdebPackageDefinition struct {
	Name         string `yaml:"name"`
	Version      string `yaml:"version"`
	Url          string `yaml:"url"`
	Sha256       string `yaml:"sha256"`
	Path         string `yaml:"path,omitempty"`
	Provider     string `yaml:"provider,omitempty"`
	Distribution string `yaml:"distribution,omitempty"`
	Component    string `yaml:"component,omitempty"`
}

type XdebProviderDefinition struct {
	Xdeb []XdebPackageDefinition `yaml:"xdeb"`
}

func parseYamlDefinition(path string) (*XdebProviderDefinition, error) {
	yamlFile, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	definition := XdebProviderDefinition{}
	err = yaml.Unmarshal(yamlFile, &definition)

	if err != nil {
		return nil, err
	}

	return &definition, nil
}

func packageDefinitionWithMetadata(packageDefinition *XdebPackageDefinition, path string) *XdebPackageDefinition {
	distPath := filepath.Dir(path)

	packageObject := *packageDefinition
	packageObject.Component = filepath.Base(strings.TrimSuffix(path, filepath.Ext(path)))
	packageObject.Distribution = filepath.Base(distPath)
	packageObject.Provider = filepath.Base(filepath.Dir(distPath))

	return &packageObject
}

func findPackage(name string, path string) (*XdebPackageDefinition, error) {
	globPattern := filepath.Join(path, "**", "*.yaml")
	globbed, err := filepathx.Glob(globPattern)

	if err != nil {
		return nil, err
	}

	for _, match := range globbed {
		definition, err := parseYamlDefinition(match)

		if err != nil {
			return nil, err
		}

		for _, packageDefinition := range definition.Xdeb {
			if packageDefinition.Name == name {
				return packageDefinitionWithMetadata(&packageDefinition, match), nil
			}
		}
	}

	return nil, fmt.Errorf("Could not find package %s", name)
}

func downloadPackage(path string, url string) (string, error) {
	err := os.MkdirAll(path, os.ModePerm)

	if err != nil {
		return "", fmt.Errorf("Could not create path %s", path)
	}

	fullPath := filepath.Join(path, filepath.Base(url))
	out, err := os.Create(fullPath)

	if err != nil {
		return "", fmt.Errorf("Could not create file %s", fullPath)
	}

	defer out.Close()

	resp, err := http.Get(url)

	if err != nil {
		return "", fmt.Errorf("Could not download file %s", url)
	}

	defer resp.Body.Close()

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

func executeCommand(workdir string, args ...string) error {
	command := exec.Command(args[0], args[1:]...)
	command.Dir = workdir
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	return command.Run()
}

func convertPackage(path string, xdebArgs string) error {
	if strings.Contains(xdebArgs, "i") {
		xdebArgs = strings.ReplaceAll(xdebArgs, "i", "")
	}

	return executeCommand(filepath.Dir(path), getXdebPath(), xdebArgs, path)
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

func installRepositoryPackage(packageDefinition *XdebPackageDefinition, context *cli.Context) error {
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

func repository(context *cli.Context) error {
	path := pathPrefix()

	provider := context.String("provider")
	distribution := context.String("distribution")

	if len(provider) > 0 {
		providers := aptProviders()

		if !slices.Contains(providers, provider) {
			log.Fatalf("APT provider %s not supported. Omit or use any of %v", provider, providers)
		}

		path = filepath.Join(path, provider)

		if len(distribution) > 0 {
			distributions := providerDistributions(provider)

			if !slices.Contains(distributions, distribution) {
				log.Fatalf(
					"APT provider %s does not support distribution %s. Omit or use any of %v",
					provider, distribution, distributions,
				)
			}

			path = filepath.Join(path, distribution)
		}
	}

	packageName := strings.Trim(context.Args().First(), " ")

	if len(packageName) == 0 {
		log.Fatalf("Please provide a package name to install.")
	}

	packageDefinition, err := findPackage(packageName, path)

	if err != nil {
		return err
	}

	log.Printf(
		"Installing %s from %s @ %s/%s\n",
		packageDefinition.Name, packageDefinition.Provider, packageDefinition.Distribution, packageDefinition.Component,
	)

	return installRepositoryPackage(packageDefinition, context)
}

func url(context *cli.Context) error {
	downloadUrl := context.Args().First()

	return installRepositoryPackage(&XdebPackageDefinition{
		Name: strings.TrimSuffix(filepath.Base(downloadUrl), filepath.Ext(downloadUrl)),
		Url:  downloadUrl,
	}, context)
}

func file(context *cli.Context) error {
	filePath := context.Args().First()

	return installRepositoryPackage(&XdebPackageDefinition{
		Name: strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath)),
		Path: filePath,
	}, context)
}

func search(context *cli.Context) error {
	packageName := context.Args().First()

	if len(packageName) == 0 {
		return fmt.Errorf("No package provided to search for.")
	}

	packageDefinitions := []XdebPackageDefinition{}
	globPattern := filepath.Join(pathPrefix(), "**", "*.yaml")
	globbed, err := filepathx.Glob(globPattern)

	if err != nil {
		return err
	}

	for _, match := range globbed {
		definition, err := parseYamlDefinition(match)

		if err != nil {
			return err
		}

		for _, packageDefinition := range definition.Xdeb {
			if packageDefinition.Name == packageName {
				packageDefinitions = append(packageDefinitions, *packageDefinitionWithMetadata(&packageDefinition, match))
			}
		}
	}

	for _, packageDefinition := range packageDefinitions {
		fmt.Printf("%s/%s\n", packageDefinition.Provider, packageDefinition.Component)
		fmt.Printf("  distribution: %s\n", packageDefinition.Distribution)
		fmt.Printf("  version: %s\n", packageDefinition.Version)
		fmt.Printf("  url: %s\n", packageDefinition.Url)
		fmt.Printf("  sha256: %s\n", packageDefinition.Sha256)
		fmt.Println()
	}

	return err
}

func parsePackagesFile(urlPrefix string, packages_file string) *XdebProviderDefinition {
	definition := XdebProviderDefinition{}
	packages := strings.Split(packages_file, "\n\n")

	for _, package_data := range packages {
		if len(package_data) == 0 {
			continue
		}

		var name string
		var version string
		var url string
		var sha256 string

		for _, line := range strings.Split(package_data, "\n") {
			if strings.HasPrefix(line, "Package:") {
				name = strings.Split(line, ": ")[1]
				continue
			}

			if strings.HasPrefix(line, "Version:") {
				version = strings.Split(line, ": ")[1]
				continue
			}

			if strings.HasPrefix(line, "Filename:") {
				suffix := strings.Split(line, ": ")[1]
				url = fmt.Sprintf("%s/%s", urlPrefix, suffix)
				continue
			}

			if strings.HasPrefix(line, "SHA256:") {
				sha256 = strings.Split(line, ": ")[1]
				continue
			}
		}

		definition.Xdeb = append(definition.Xdeb, XdebPackageDefinition{
			Name:    name,
			Version: version,
			Url:     url,
			Sha256:  sha256,
		})
	}

	return &definition
}

func pullPackages(url string, dist string, component string, architecture string) (*XdebProviderDefinition, error) {
	packagesFileUrl := fmt.Sprintf(
		"%s/dists/%s/%s/binary-%s/Packages",
		url, dist, component, architecture,
	)

	packagesFile, err := http.Get(packagesFileUrl)

	if err != nil {
		return nil, err
	}

	if packagesFile.StatusCode != 200 {
		packagesFile, err = http.Get(fmt.Sprintf("%s.xz", packagesFileUrl))

		if err != nil {
			return nil, err
		}
	}

	if packagesFile.StatusCode != 200 {
		packagesFile, err = http.Get(fmt.Sprintf("%s.gz", packagesFileUrl))

		if err != nil {
			return nil, err
		}
	}

	requestUrl := fmt.Sprintf(
		"%s://%s%s",
		packagesFile.Request.URL.Scheme, packagesFile.Request.URL.Host, packagesFile.Request.URL.Path,
	)

	if packagesFile.StatusCode != 200 {
		return nil, nil
	}

	fmt.Printf("Syncing repository %s\n", requestUrl)
	var packagesFileContent string

	if strings.HasSuffix(requestUrl, ".xz") {
		reader, err := xz.NewReader(packagesFile.Body)

		if err != nil {
			log.Fatalf("NewReader error %s", err)
		}

		var outputBuffer bytes.Buffer
		outputWriter := bufio.NewWriter(&outputBuffer)

		if _, err = io.Copy(outputWriter, reader); err != nil {
			log.Fatalf("io.Copy error %s", err)
		}

		packagesFileContent = outputBuffer.String()
	} else if strings.HasSuffix(requestUrl, ".gz") {
		reader, err := gzip.NewReader(packagesFile.Body)

		if err != nil {
			return nil, err
		}

		output, err := io.ReadAll(reader)

		if err != nil {
			return nil, err
		}

		packagesFileContent = string(output)
	} else {
		output, err := io.ReadAll(packagesFile.Body)

		if err != nil {
			return nil, err
		}

		packagesFileContent = string(output)
	}

	return parsePackagesFile(url, packagesFileContent), nil
}

func dumpPackages(directory string, url string, dist string, component string, architecture string) error {
	definition, err := pullPackages(url, dist, component, architecture)

	if err != nil {
		return err
	}

	if definition != nil && len(definition.Xdeb) > 0 {
		filePath := filepath.Join(pathPrefix(), directory, dist, fmt.Sprintf("%s.yaml", component))
		err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm)

		if err != nil {
			return err
		}

		bytes, err := yaml.Marshal(definition)

		if err != nil {
			return err
		}

		fileDump, err := os.Create(filePath)

		if err != nil {
			return err
		}

		defer fileDump.Close()
		_, err = fileDump.Write(bytes)

		if err != nil {
			return err
		}
	}

	return nil
}

type PackageListsProvider struct {
	Name          string   `yaml:"name"`
	Url           string   `yaml:"url"`
	Architecture  string   `yaml:"architecture"`
	Components    []string `yaml:"components"`
	Distributions []string `yaml:"dists"`
}

type PackageListsDefinition struct {
	Providers []PackageListsProvider `yaml:"providers"`
}

func sync(context *cli.Context) error {
	listsFile := filepath.Join(filepath.Dir(architecturePath()), "lists.yaml")
	yamlFile, err := os.ReadFile(listsFile)

	if err != nil {
		return err
	}

	lists := PackageListsDefinition{}
	err = yaml.Unmarshal(yamlFile, &lists)

	if err != nil {
		return err
	}

	for _, provider := range lists.Providers {
		for _, distribution := range provider.Distributions {
			for _, component := range provider.Components {
				err = dumpPackages(provider.Name, provider.Url, distribution, component, provider.Architecture)

				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:        APPLICATION_NAME,
		Usage:       "Automation wrapper for the xdeb utility",
		Description: "Simple tool to automatically download, convert, and install DEB packages via the awesome xdeb tool.\nBasically just a wrapper to automate the process.",
		Commands: []*cli.Command{
			{
				Name:    "repository",
				Usage:   "install a package from an online APT repository",
				Aliases: []string{"r"},
				Action:  repository,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "provider",
						Aliases: []string{"p"},
					},
					&cli.StringFlag{
						Name:    "distribution",
						Aliases: []string{"dist", "d"},
					},
				},
			},
			{
				Name:    "search",
				Usage:   "search online APT repositories for a package",
				Aliases: []string{"s"},
				Action:  search,
			},
			{
				Name:   "sync",
				Usage:  "sync online APT repositories",
				Action: sync,
			},
			{
				Name:    "url",
				Usage:   "install a package from a URL directly",
				Aliases: []string{"u"},
				Action:  url,
			},
			{
				Name:    "file",
				Usage:   "install a package from a local DEB file",
				Aliases: []string{"f"},
				Action:  file,
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "options",
				Aliases: []string{"o"},
				Usage:   "override XDEB_OPTS, '-i' will be removed if provided",
				Value:   "-Sde",
			},
			&cli.StringFlag{
				Name:    "temp",
				Aliases: []string{"t"},
				Usage:   "temporary xdeb context root path",
				Value:   filepath.Join(os.TempDir(), "xdeb"),
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
