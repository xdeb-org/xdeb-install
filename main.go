package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	xdeb "github.com/thetredev/xdeb-install/v2/pkg"
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
		return nil, fmt.Errorf("no entries found, please sync the repositories first")
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
				"provider %s does not support distribution '%s', - omit or use any of %v",
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
			return "", fmt.Errorf("provider '%s' not supported, omit or use any of %v", provider, providers)
		}

		return provider, nil
	}

	return "*", nil
}

func deb(context *cli.Context) error {
	_, err := xdeb.FindXdeb()

	if err != nil {
		return err
	}

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
		return fmt.Errorf("no package provided to install")
	}

	packageDefinitions, err := xdeb.FindPackage(packageName, path, provider, distribution, true)

	if err != nil {
		return err
	}

	return xdeb.InstallPackage(packageDefinitions[0], context)
}

func file(context *cli.Context, filePath string) error {
	_, err := xdeb.FindXdeb()

	if err != nil {
		return err
	}

	var packageDefinition xdeb.XdebPackageDefinition

	fileUrl, err := url.Parse(filePath)
	isUrl := err == nil && fileUrl.Scheme != "" && fileUrl.Host != ""

	if !isUrl {
		if _, err := os.Stat(filePath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file '%s' does not exist", filePath)
			}

			return err
		}

		if !strings.HasSuffix(filePath, ".deb") {
			return fmt.Errorf("file '%s' is not a valid DEB package", filePath)
		}

		packageDefinition = xdeb.XdebPackageDefinition{
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
	} else {
		packageDefinition = xdeb.XdebPackageDefinition{
			Name: xdeb.TrimPathExtension(filepath.Base(filePath), 1),
			Url:  filePath,
		}
	}

	return xdeb.InstallPackage(&packageDefinition, context)
}

func search(context *cli.Context) error {
	packageName := context.Args().First()

	if len(packageName) == 0 {
		return fmt.Errorf("no package provided to search for")
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

	showDetails := context.Bool("details")

	for _, provider := range lists.Providers {
		fmt.Println(provider.Name)
		fmt.Printf("  architecture: %s\n", provider.Architecture)

		if provider.Custom {
			fmt.Printf("  url: %s/%s/%s\n", xdeb.XDEB_INSTALL_REPOSITORIES_URL, xdeb.XDEB_INSTALL_REPOSITORIES_TAG, provider.Url)
		} else {
			fmt.Printf("  url: %s\n", provider.Url)
		}

		if showDetails {
			for _, distribution := range provider.Distributions {
				fmt.Printf("    distribution: %s\n", distribution)

				for _, component := range provider.Components {
					fmt.Printf("      component: %s\n", component)
				}
			}
		}

		fmt.Println()
	}

	return nil
}

func prepare(context *cli.Context) error {
	version := context.Args().First()
	var urlPrefix string

	if len(version) == 0 {
		// install master
		urlPrefix = xdeb.XDEB_MASTER_URL
		version = "master"
	} else {
		if version == "latest" {
			// find latest release tag
			urlPrefix = fmt.Sprintf("%s/latest", xdeb.XDEB_URL)

			client := &http.Client{}
			resp, err := client.Get(urlPrefix)

			if err != nil {
				return fmt.Errorf("could not follow URL '%s'", urlPrefix)
			}

			// get latest release version
			releaseUrl := resp.Request.URL.String()
			version = releaseUrl[strings.LastIndex(releaseUrl, "/")+1:]
		}

		urlPrefix = fmt.Sprintf("%s/download/%s", xdeb.XDEB_URL, version)
	}

	requestUrl := fmt.Sprintf("%s/xdeb", urlPrefix)
	path := filepath.Join(os.TempDir(), "xdeb-download", "xdeb")

	xdeb.LogMessage("Downloading xdeb [%s] from %s to %s ...", version, requestUrl, path)
	xdebFile, err := xdeb.DownloadFile(path, requestUrl, false, false)

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
		return fmt.Errorf("please provide a temporary xdeb context root path")
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

func unixTime(epochString string) (*time.Time, error) {
	if epochString == "now" {
		now := time.Now().UTC()
		return &now, nil
	}

	parsed, err := strconv.ParseInt(epochString, 10, 64)

	if err != nil {
		return nil, err
	}

	epoch := time.Unix(parsed, 0)
	return &epoch, nil
}

var (
	VersionString   string = "dev"
	VersionCompiled string = "now"
	VersionAuthors  string = "me <me@localhost>"
)

func main() {
	authors := []*cli.Author{}

	for _, authorString := range strings.Split(VersionAuthors, ",") {
		authorItems := strings.Split(authorString, " <")
		authorEmail := ""

		if len(authorItems) > 1 {
			authorEmail = authorItems[1]
		}

		authorEmail = strings.TrimSuffix(authorEmail, ">")

		authors = append(authors, &cli.Author{
			Name:  authorItems[0],
			Email: authorEmail,
		})
	}

	compiled, err := unixTime(VersionCompiled)

	if err != nil {
		log.Fatalf("cannot parse compiled time from unix epoch: %s", err.Error())
	}

	app := &cli.App{
		Name:        xdeb.APPLICATION_NAME,
		Usage:       "Automation wrapper for the xdeb utility",
		UsageText:   fmt.Sprintf("%s [global options (except --file)] <package>\n%s [global options] command [command options] [arguments...]", xdeb.APPLICATION_NAME, xdeb.APPLICATION_NAME),
		Description: "Simple tool to automatically download, convert, and install DEB packages via the awesome xdeb utility.\nBasically just a wrapper to automate the process.",
		Version:     VersionString,
		Compiled:    *compiled,
		Authors:     authors,
		Suggest:     true,
		Action:      deb,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Usage:   "install a package from a local DEB file or remote URL",
				Aliases: []string{"f"},
				Action:  file,
			},
			&cli.StringFlag{
				Name:    "provider",
				Usage:   "limit search results to a specific provider when --file is not passed",
				Aliases: []string{"p"},
			},
			&cli.StringFlag{
				Name:    "distribution",
				Usage:   "limit search results to a specific distribution (requires --provider)",
				Aliases: []string{"dist", "d"},
			},
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
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "details",
						Usage: "display provider details (distributions and components)",
					},
				},
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
	}

	if err := app.Run(os.Args); err != nil {
		xdeb.LogMessage(err.Error())
		os.Exit(1)
	}
}
