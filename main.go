package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/exp/slices"

	xdeb "github.com/thetredev/xdeb-install/pkg"
	"github.com/urfave/cli/v2"
)

func readPath(subdir string) ([]string, error) {
	path, err := xdeb.RepositoryPath()

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
				"Provider %s does not support distribution %s. Omit or use any of %v",
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
			return "", fmt.Errorf("Provider %s not supported. Omit or use any of %v", provider, providers)
		}

		return provider, nil
	}

	return "*", nil
}

func repository(context *cli.Context) error {
	path, err := xdeb.RepositoryPath()

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
		return fmt.Errorf("No package provided to install.")
	}

	packageDefinitions, err := xdeb.FindPackage(packageName, path, provider, distribution, true)

	if err != nil {
		return err
	}

	return xdeb.InstallPackage(packageDefinitions[0], context)
}

func url(context *cli.Context) error {
	downloadUrl := context.Args().First()

	return xdeb.InstallPackage(&xdeb.XdebPackageDefinition{
		Name: xdeb.TrimPathExtension(filepath.Base(downloadUrl), 1),
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

	if !strings.HasSuffix(filePath, ".deb") {
		return fmt.Errorf("File %s is not a valid DEB package.", filePath)
	}

	packageDefinition := xdeb.XdebPackageDefinition{
		Name: xdeb.TrimPathExtension(filepath.Base(filePath), 1),
	}

	packageDefinition.Configure(context.String("temp"))

	if filePath != packageDefinition.FilePath {
		// copy file to temporary xdeb context path
		if err := os.MkdirAll(packageDefinition.Path, os.ModePerm); err != nil {
			return err
		}

		data, err := os.ReadFile(filePath)

		if err != nil {
			return err
		}

		if err = os.WriteFile(packageDefinition.FilePath, data, os.ModePerm); err != nil {
			return err
		}
	}

	return xdeb.InstallPackage(&packageDefinition, context)
}

func search(context *cli.Context) error {
	packageName := context.Args().First()

	if len(packageName) == 0 {
		return fmt.Errorf("No package provided to search for.")
	}

	path, err := xdeb.RepositoryPath()

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

	packageDefinitions, err := xdeb.FindPackage(packageName, path, provider, distribution, context.Bool("exact"))

	if err != nil {
		return err
	}

	for _, packageDefinition := range packageDefinitions {
		fmt.Printf("%s/%s\n", packageDefinition.Provider, packageDefinition.Component)
		fmt.Printf("  package: %s\n", packageDefinition.Name)
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
	lists, err := xdeb.ParsePackageLists()

	if err != nil {
		return err
	}

	args := context.Args()
	providerNames := []string{}

	for i := 0; i < args.Len(); i++ {
		providerName := args.Get(i)
		providerNames = append(providerNames, providerName)
	}

	err = xdeb.SyncRepositories(lists, providerNames...)

	if err != nil {
		return err
	}

	xdeb.LogMessage("Finished syncing: %s", strings.ReplaceAll(lists.Path, os.Getenv("HOME"), "~"))
	return nil
}

func providers(context *cli.Context) error {
	lists, err := xdeb.ParsePackageLists()

	if err != nil {
		return err
	}

	for _, provider := range lists.Providers {
		fmt.Println(provider.Name)
		fmt.Printf("  architecture: %s\n", provider.Architecture)
		fmt.Printf("  url: %s\n", provider.Url)

		for _, distribution := range provider.Distributions {
			fmt.Printf("    distribution: %s\n", distribution)

			for _, component := range provider.Components {
				fmt.Printf("      component: %s\n", component)
			}
		}

		fmt.Println()
	}

	return nil
}

func prepare(context *cli.Context) error {
	version := context.Args().First()
	var url string

	if len(version) == 0 {
		// install master
		url = xdeb.XDEB_MASTER_URL
		version = "master"
	} else {
		url = fmt.Sprintf(xdeb.XDEB_RELEASE_URL, version)
	}

	path := filepath.Join(os.TempDir(), "xdeb-download")

	xdeb.LogMessage("Downloading xdeb [%s] from %s to %s ...", version, url, path)
	xdebFile, err := xdeb.DownloadFile(path, url, false, false)

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
		path, err := xdeb.RepositoryPath()

		if err != nil {
			return err
		}

		xdeb.LogMessage("Cleaning up repository path: %s", path)
		return os.RemoveAll(path)
	}

	return nil
}

var (
	VersionString = "dev"
	VersionDate   = "now"
	VersionAuthor = "me"
)

func version(context *cli.Context) error {
	xdeb.LogMessage("version information")
	fmt.Printf("  version:  %s %s/%s\n", VersionString, runtime.GOOS, runtime.GOARCH)
	fmt.Printf("  created:  %s\n", VersionDate)
	fmt.Printf("  author:   %s\n", VersionAuthor)
	return nil
}

func main() {
	app := &cli.App{
		Name:        xdeb.APPLICATION_NAME,
		Usage:       "Automation wrapper for the xdeb utility",
		Description: "Simple tool to automatically download, convert, and install DEB packages via the awesome xdeb utility.\nBasically just a wrapper to automate the process.",
		Commands: []*cli.Command{
			{
				Name:     "xdeb",
				HelpName: "xdeb [release version]",
				Usage:    "install the xdeb utility to the system along with its dependencies",
				Action:   prepare,
			},
			{
				Name:    "providers",
				Usage:   "list available providers",
				Aliases: []string{"p"},
				Action:  providers,
			},
			{
				Name:     "sync",
				HelpName: "sync [provider list]",
				Usage:    "synchronize remote repositories",
				Aliases:  []string{"S"},
				Action:   sync,
			},
			{
				Name:    "search",
				Usage:   "search remote repositories for a package",
				Aliases: []string{"s"},
				Action:  search,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "exact",
						Aliases: []string{"e"},
						Usage:   "perform an exact match of the package name provided",
					},
					&cli.StringFlag{
						Name:    "provider",
						Usage:   "limit search results to a specific provider",
						Aliases: []string{"p"},
					},
					&cli.StringFlag{
						Name:    "distribution",
						Usage:   "limit search results to a specific distribution (requires --provider)",
						Aliases: []string{"dist", "d"},
					},
				},
			},
			{
				Name:    "repository",
				Usage:   "install a package from an remote repository",
				Aliases: []string{"r"},
				Action:  repository,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "provider",
						Usage:   "limit search results to a specific provider",
						Aliases: []string{"p"},
					},
					&cli.StringFlag{
						Name:    "distribution",
						Usage:   "limit search results to a specific distribution (requires --provider)",
						Aliases: []string{"dist", "d"},
					},
				},
			},
			{
				Name:     "url",
				HelpName: "url [URL]",
				Usage:    "install a package from a URL directly",
				Aliases:  []string{"u"},
				Action:   url,
			},
			{
				Name:     "file",
				HelpName: "file [path]",
				Usage:    "install a package from a local DEB file",
				Aliases:  []string{"f"},
				Action:   file,
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
			{
				Name:    "version",
				Usage:   "print the version of this tool",
				Aliases: []string{"v"},
				Action:  version,
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
				Usage:   "set the temporary xdeb context root path",
				Value:   filepath.Join(os.TempDir(), "xdeb"),
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		xdeb.LogMessage(err.Error())
		os.Exit(1)
	}
}
