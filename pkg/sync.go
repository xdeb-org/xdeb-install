package xdeb

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ulikunitz/xz"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"
)

type PackageListsProvider struct {
	Name          string   `yaml:"name"`
	Custom        bool     `yaml:"custom"`
	Url           string   `yaml:"url"`
	Architecture  string   `yaml:"architecture"`
	Components    []string `yaml:"components"`
	Distributions []string `yaml:"dists"`
}

type PackageListsDefinition struct {
	Path      string                 `yaml:"path,omitempty"`
	Providers []PackageListsProvider `yaml:"providers"`
}

func parsePackagesFile(urlPrefix string, packagesFile string) *XdebProviderDefinition {
	definition := XdebProviderDefinition{}
	packages := strings.Split(packagesFile, "\n\n")

	for _, packageData := range packages {
		if len(packageData) == 0 {
			continue
		}

		packageDefinition := XdebPackageDefinition{}
		scanner := bufio.NewScanner(strings.NewReader(packageData))

		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, "Package:") {
				packageDefinition.Name = line[strings.Index(line, ":")+2:]
				continue
			}

			if strings.HasPrefix(line, "Version:") {
				packageDefinition.Version = line[strings.Index(line, ":")+2:]
				continue
			}

			if strings.HasPrefix(line, "Filename:") {
				suffix := line[strings.Index(line, ":")+2:]
				packageDefinition.Url = fmt.Sprintf("%s/%s", urlPrefix, suffix)
				continue
			}

			if strings.HasPrefix(line, "SHA256:") {
				packageDefinition.Sha256 = line[strings.Index(line, ":")+2:]
				continue
			}
		}

		definition.Xdeb = append(definition.Xdeb, &packageDefinition)
	}

	return &definition
}

func pullPackagesFile(urlPrefix string, dist string, component string, architecture string) (*XdebProviderDefinition, error) {
	url := fmt.Sprintf(
		"%s/dists/%s/%s/binary-%s/Packages",
		urlPrefix, dist, component, architecture,
	)

	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		resp, err = http.Get(fmt.Sprintf("%s.xz", url))

		if err != nil {
			return nil, err
		}
	}

	if resp.StatusCode != 200 {
		resp, err = http.Get(fmt.Sprintf("%s.gz", url))

		if err != nil {
			return nil, err
		}
	}

	if resp.StatusCode != 200 {
		return nil, nil
	}

	defer resp.Body.Close()

	requestUrl := fmt.Sprintf(
		"%s://%s%s",
		resp.Request.URL.Scheme, resp.Request.URL.Host, resp.Request.URL.Path,
	)

	var reader io.Reader

	if strings.HasSuffix(requestUrl, ".xz") {
		reader, err = xz.NewReader(resp.Body)

		if err != nil {
			return nil, err
		}
	} else if strings.HasSuffix(requestUrl, ".gz") {
		reader, err = gzip.NewReader(resp.Body)

		if err != nil {
			return nil, err
		}
	} else {
		reader = resp.Body
	}

	output, err := io.ReadAll(reader)

	if err != nil {
		return nil, err
	}

	return parsePackagesFile(urlPrefix, string(output)), nil
}

func pullAptRepository(directory string, url string, dist string, component string, architecture string) error {
	definition, err := pullPackagesFile(url, dist, component, architecture)

	if err != nil {
		return err
	}

	if definition != nil && len(definition.Xdeb) > 0 {
		LogMessage("Syncing repository %s/%s: %s", filepath.Base(directory), dist, component)

		filePath := filepath.Join(directory, dist, fmt.Sprintf("%s.yaml", component))
		data, err := yaml.Marshal(definition)

		if err != nil {
			return err
		}

		if _, err = writeFileCompressed(filePath, data); err != nil {
			return err
		}
	}

	return nil
}

func pullCustomRepository(directory string, urlPrefix string, dist string, component string) error {
	LogMessage("Syncing repository %s/%s: %s", filepath.Base(urlPrefix), dist, component)

	url := fmt.Sprintf("%s/%s/%s", urlPrefix, dist, component)
	_, err := DownloadFile(filepath.Join(directory, dist), url, false, true)

	return err
}

func ParsePackageLists() (*PackageListsDefinition, error) {
	arch, err := FindArchitecture()

	if err != nil {
		return nil, err
	}

	path, err := RepositoryPath()

	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(XDEB_INSTALL_REPOSITORIES_URL, XDEB_INSTALL_REPOSITORIES_TAG, arch)
	LogMessage("Syncing lists: %s", url)

	listsFile, err := DownloadFile(path, url, true, true)

	if err != nil {
		return nil, err
	}

	yamlFile, err := decompressFile(listsFile)

	if err != nil {
		return nil, err
	}

	lists := &PackageListsDefinition{}
	err = yaml.Unmarshal(yamlFile, lists)

	if err != nil {
		return nil, err
	}

	lists.Path = path
	return lists, nil
}

func SyncRepositories(lists *PackageListsDefinition, providerNames ...string) error {
	availableProviderNames := []string{}

	for _, provider := range lists.Providers {
		availableProviderNames = append(availableProviderNames, provider.Name)
	}

	for _, providerName := range providerNames {
		if !slices.Contains(availableProviderNames, providerName) {
			return fmt.Errorf("Provider %s not supported. Omit or use any of %v", providerName, availableProviderNames)
		}
	}

	providers := []PackageListsProvider{}

	if len(providerNames) > 0 {
		for _, provider := range lists.Providers {
			if slices.Contains(providerNames, provider.Name) {
				providers = append(providers, provider)
			}
		}
	} else {
		providers = append(providers, lists.Providers...)
	}

	operations := len(providers)

	for _, provider := range providers {
		operations += len(provider.Distributions) * len(provider.Components)
	}

	for _, provider := range providers {
		errors := make(chan error, operations)
		var wg sync.WaitGroup

		for _, distribution := range provider.Distributions {
			for _, component := range provider.Components {
				wg.Add(1)

				go func(p PackageListsProvider, d string, c string) {
					defer wg.Done()
					directory := filepath.Join(lists.Path, p.Name)

					if p.Custom {
						errors <- pullCustomRepository(directory, p.Url, d, c)
					} else {
						errors <- pullAptRepository(directory, p.Url, d, c, p.Architecture)
					}
				}(provider, distribution, component)
			}
		}

		wg.Wait()
		close(errors)

		for i := 0; i < operations; i++ {
			err := <-errors

			if err != nil {
				return err
			}
		}
	}

	return nil
}
