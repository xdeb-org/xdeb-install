package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/adrg/xdg"
	xdeb "github.com/thetredev/xdeb-install/pkg"
	"github.com/urfave/cli/v2"
)

func repositoryPath() (string, error) {
	arch, err := xdeb.FindArchitecture()

	if err != nil {
		return "", err
	}

	return filepath.Join(xdg.ConfigHome, xdeb.APPLICATION_NAME, "repositories", arch), nil
}

func readPath(subdir string) ([]string, error) {
	path, err := repositoryPath()

	if err != nil {
		return nil, err
	}

	entriesPath := filepath.Join(path, subdir)
	entries, err := os.ReadDir(entriesPath)

	if err != nil {
		return nil, fmt.Errorf("No entries found. Please sync the repositories first.")
	}

	subdirs := []string{}

	for _, entry := range entries {
		if entry.IsDir() {
			subdirs = append(subdirs, entry.Name())
		}
	}

	return subdirs, nil
}

func findDistribution(provider string, distribution string) (string, error) {
	if len(distribution) > 0 && distribution != "*" {
		distributions, err := readPath(provider)

		if err != nil {
			return "", err
		}

		if !slices.Contains(distributions, distribution) {
			return "", fmt.Errorf(
				"APT provider %s does not support distribution %s. Omit or use any of %v",
				provider, distribution, distributions,
			)
		}

		return distribution, nil
	}

	return "*", nil
}

func findProvider(provider string) (string, error) {
	if len(provider) > 0 {
		providers, err := readPath("")

		if err != nil {
			return "", err
		}

		if !slices.Contains(providers, provider) {
			return "", fmt.Errorf("APT provider %s not supported. Omit or use any of %v", provider, providers)
		}

		return provider, nil
	}

	return "*", nil
}

func repository(context *cli.Context) error {
	path, err := repositoryPath()

	if err != nil {
		return nil
	}

	provider, err := findProvider(context.String("provider"))

	if err != nil {
		return err
	}

	distribution, err := findDistribution(provider, context.String("distribution"))

	if err != nil {
		return err
	}

	packageName := strings.Trim(context.Args().First(), " ")

	if len(packageName) == 0 {
		return fmt.Errorf("Please provide a package name to install.")
	}

	packageDefinitions, err := xdeb.FindPackage(packageName, path, provider, distribution)

	if err != nil {
		return err
	}

	return xdeb.InstallPackage(packageDefinitions[0], context)
}

func url(context *cli.Context) error {
	downloadUrl := context.Args().First()

	return xdeb.InstallPackage(&xdeb.XdebPackageDefinition{
		Name: xdeb.TrimPathExtension(filepath.Base(downloadUrl)),
		Url:  downloadUrl,
	}, context)
}

func file(context *cli.Context) error {
	filePath := context.Args().First()

	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("File %s does not exist.", filePath)
		}

		return err
	}

	return xdeb.InstallPackage(&xdeb.XdebPackageDefinition{
		Name: xdeb.TrimPathExtension(filepath.Base(filePath)),
		Path: filePath,
	}, context)
}

func search(context *cli.Context) error {
	packageName := context.Args().First()

	if len(packageName) == 0 {
		return fmt.Errorf("No package provided to search for.")
	}

	path, err := repositoryPath()

	if err != nil {
		return err
	}

	provider, err := findProvider(context.String("provider"))

	if err != nil {
		return err
	}

	distribution, err := findDistribution(provider, context.String("distribution"))

	if err != nil {
		return err
	}

	packageDefinitions, err := xdeb.FindPackage(packageName, path, provider, distribution)

	if err != nil {
		return err
	}

	for _, packageDefinition := range packageDefinitions {
		fmt.Printf("%s/%s\n", packageDefinition.Provider, packageDefinition.Component)
		fmt.Printf("  distribution: %s\n", packageDefinition.Distribution)

		if len(packageDefinition.Version) > 0 {
			fmt.Printf("  version: %s\n", packageDefinition.Version)
		}

		fmt.Printf("  url: %s\n", packageDefinition.Url)

		if len(packageDefinition.Sha256) > 0 {
			fmt.Printf("  sha256: %s\n", packageDefinition.Sha256)
		}

		fmt.Println()
	}

	return nil
}

func sync(context *cli.Context) error {
	arch, err := xdeb.FindArchitecture()

	if err != nil {
		return err
	}

	path, err := repositoryPath()

	if err != nil {
		return err
	}

	return xdeb.SyncRepositories(path, arch)
}

func prepare(context *cli.Context) error {
	version := context.Args().First()
	var url string

	if len(version) == 0 {
		// install master
		url = "https://raw.githubusercontent.com/toluschr/xdeb/master/xdeb"
		version = "master"
	} else {
		url = fmt.Sprintf("https://github.com/toluschr/xdeb/releases/download/%s/xdeb", version)
	}

	path := filepath.Join(os.TempDir(), "xdeb-download")

	xdeb.LogMessage("Downloading xdeb [%s] from %s to %s ...", version, url, path)
	xdebFile, err := xdeb.DownloadFile(path, url, false)

	if err != nil {
		return err
	}

	// move to /usr/local/bin
	args := []string{}

	if os.Getuid() > 0 {
		xdeb.LogMessage("Detected non-root user, appending 'sudo' to command")
		args = append(args, "sudo")
	}

	targetFile := "/usr/local/bin/xdeb"
	args = append(args, "mv", xdebFile, targetFile)

	err = xdeb.ExecuteCommand("", args...)

	if err != nil {
		return err
	}

	args = make([]string, 0)

	if os.Getuid() > 0 {
		xdeb.LogMessage("Detected non-root user, appending 'sudo' to command")
		args = append(args, "sudo")
	}

	args = append(args, "chmod", "u+x", targetFile)
	err = xdeb.ExecuteCommand("", args...)

	if err != nil {
		return err
	}

	path = filepath.Dir(xdebFile)

	xdeb.LogMessage("Removing temporary download location %s ...", path)
	err = os.RemoveAll(path)

	if err != nil {
		return err
	}

	args = make([]string, 0)

	if os.Getuid() > 0 {
		args = append(args, "sudo")
	}

	args = append(args, "xbps-install", "-S", "binutils", "tar", "curl", "xz")
	return xdeb.ExecuteCommand("", args...)
}

func clean(context *cli.Context) error {
	tempPath := context.String("temp")

	if len(tempPath) == 0 {
		return fmt.Errorf("Please provide a temporary xdeb context root path.")
	}

	xdeb.LogMessage("Cleaning up temporary xdeb context root path: %s", tempPath)
	err := os.RemoveAll(tempPath)

	if err != nil {
		return err
	}

	if context.Bool("lists") {
		path, err := repositoryPath()

		if err != nil {
			return err
		}

		xdeb.LogMessage("Cleaning up repository path: %s", path)
		return os.RemoveAll(path)
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:        xdeb.APPLICATION_NAME,
		Usage:       "Automation wrapper for the xdeb utility",
		Description: "Simple tool to automatically download, convert, and install DEB packages via the awesome xdeb tool.\nBasically just a wrapper to automate the process.",
		Commands: []*cli.Command{
			{
				Name:   "xdeb",
				Usage:  "installs the xdeb utility to the system along with its dependencies",
				Action: prepare,
			},
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
			{
				Name:    "clean",
				Usage:   "cleanup temporary xdeb context root path, optionally the repository lists as well",
				Aliases: []string{"c"},
				Action:  clean,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "lists",
						Aliases: []string{"l"},
						Usage:   "cleanup repository lists as well",
						Value:   false,
					},
				},
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
		xdeb.LogMessage(err.Error())
		os.Exit(1)
	}
}
