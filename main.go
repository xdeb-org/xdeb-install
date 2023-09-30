package main

import (
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

	"github.com/urfave/cli/v2"
	"github.com/yargevad/filepathx"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"
)

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

func pathPrefix() string {
	arch, ok := ARCHITECTURE_MAP[runtime.GOARCH]

	if !ok {
		log.Fatalf("Architecture %s not supported (yet).", runtime.GOARCH)
	}

	return fmt.Sprintf("repositories/%s/apt", arch)
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
	entries, err := os.ReadDir(fmt.Sprintf("%s/%s", pathPrefix(), provider))

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
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Url     string `yaml:"url"`
	Sha256  string `yaml:"sha256"`
}

type XdebProviderDefinition struct {
	Xdeb []XdebPackageDefinition `yaml:"xdeb"`
}

func findPackage(name string, path string) (*XdebPackageDefinition, error) {
	globbed, err := filepathx.Glob(fmt.Sprintf("%s/**/*.yaml", path))

	if err != nil {
		return nil, err
	}

	for _, match := range globbed {
		yamlFile, err := os.ReadFile(match)

		if err != nil {
			return nil, err
		}

		definition := XdebProviderDefinition{}
		err = yaml.Unmarshal(yamlFile, &definition)

		if err != nil {
			return nil, err
		}

		for _, packageDefinition := range definition.Xdeb {
			if packageDefinition.Name == name {
				return &packageDefinition, nil
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
	fmt.Printf("%v\n", args)
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

func installRepositoryPackage(packageDefinition *XdebPackageDefinition, path string, xdebArgs string) {
	// download
	fullPath, err := downloadPackage(path, packageDefinition.Url)

	if err != nil {
		log.Fatal(err)
	}

	// compare checksums
	err = comparePackageChecksums(fullPath, packageDefinition.Sha256)

	if err != nil {
		log.Fatal(err)
	}

	// xdeb convert
	err = convertPackage(fullPath, xdebArgs)

	if err != nil {
		log.Fatal(err)
	}

	// xbps-install
	err = installPackage(fullPath)

	if err != nil {
		log.Fatal(err)
	}

	// cleanup
	err = os.RemoveAll(path)

	if err != nil {
		log.Fatal(err)
	}
}

func repository(context *cli.Context) error {
	//xdebPath := getXdebPath()
	arch, ok := ARCHITECTURE_MAP[runtime.GOARCH]

	if !ok {
		log.Fatalf("Architecture %s not supported (yet).", runtime.GOARCH)
	}

	fmt.Println(arch)

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

	packageName := strings.Trim(context.Args().Get(0), " ")

	if len(packageName) == 0 {
		log.Fatalf("Please provide a package name to install.")
	}

	packageDefinition, err := findPackage(packageName, path)

	if err != nil {
		log.Fatal(err)
	}

	installRepositoryPackage(packageDefinition, filepath.Join(context.String("temp"), packageName), context.String("options"))
	return nil
}

func url(context *cli.Context) error {
	//xdebPath := getXdebPath()
	return nil
}

func file(context *cli.Context) error {
	//xdebPath := getXdebPath()
	return nil
}

func main() {
	app := &cli.App{
		Name:        "xdeb-install",
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
				Value:   fmt.Sprintf("%s/xdeb", os.TempDir()),
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
