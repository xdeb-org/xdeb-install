package main

import (
	"fmt"
	"log"
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

func installRepositoryPackage(packageDefinition *XdebPackageDefinition) {
	// download
	// install
	// cleanup
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

	installRepositoryPackage(packageDefinition)
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
				Usage:   "override XDEB_OPTS",
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
