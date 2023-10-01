package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"

	"github.com/adrg/xdg"
	xdeb "github.com/thetredev/xdeb-install/pkg"
	"github.com/urfave/cli/v2"
	"github.com/yargevad/filepathx"
)

const APPLICATION_NAME = "xdeb-install"

var ARCHITECTURE_MAP = map[string]string{
	"amd64": "x86_64",
	//"386":   "i386",
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

	packageDefinition, err := xdeb.FindPackage(packageName, path)

	if err != nil {
		return err
	}

	log.Printf(
		"Installing %s from %s @ %s/%s\n",
		packageDefinition.Name, packageDefinition.Provider, packageDefinition.Distribution, packageDefinition.Component,
	)

	return xdeb.InstallRepositoryPackage(packageDefinition, context)
}

func url(context *cli.Context) error {
	downloadUrl := context.Args().First()

	return xdeb.InstallRepositoryPackage(&xdeb.XdebPackageDefinition{
		Name: strings.TrimSuffix(filepath.Base(downloadUrl), filepath.Ext(downloadUrl)),
		Url:  downloadUrl,
	}, context)
}

func file(context *cli.Context) error {
	filePath := context.Args().First()

	return xdeb.InstallRepositoryPackage(&xdeb.XdebPackageDefinition{
		Name: strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath)),
		Path: filePath,
	}, context)
}

func search(context *cli.Context) error {
	packageName := context.Args().First()

	if len(packageName) == 0 {
		return fmt.Errorf("No package provided to search for.")
	}

	packageDefinitions := []xdeb.XdebPackageDefinition{}
	globPattern := filepath.Join(pathPrefix(), "**", "*.yaml")
	globbed, err := filepathx.Glob(globPattern)

	if err != nil {
		return err
	}

	for _, match := range globbed {
		definition, err := xdeb.ParseYamlDefinition(match)

		if err != nil {
			return err
		}

		for _, packageDefinition := range definition.Xdeb {
			if packageDefinition.Name == packageName {
				packageDefinitions = append(packageDefinitions, *xdeb.PackageDefinitionWithMetadata(&packageDefinition, match))
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

func sync(context *cli.Context) error {
	listsFile := filepath.Join(filepath.Dir(architecturePath()), "lists.yaml")
	yamlFile, err := os.ReadFile(listsFile)

	if err != nil {
		return err
	}

	lists := xdeb.PackageListsDefinition{}
	err = yaml.Unmarshal(yamlFile, &lists)

	if err != nil {
		return err
	}

	for _, provider := range lists.Providers {
		for _, distribution := range provider.Distributions {
			for _, component := range provider.Components {
				err = xdeb.DumpPackages(filepath.Join(pathPrefix(), provider.Name), provider.Url, distribution, component, provider.Architecture)

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
