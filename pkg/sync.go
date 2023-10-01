package xdeb

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
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
	Providers []PackageListsProvider `yaml:"providers"`
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

		if err != nil {
			return nil, err
		}
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
		fmt.Printf("Syncing repository %s/%s @ %s\n", filepath.Base(directory), component, dist)

		filePath := filepath.Join(directory, dist, fmt.Sprintf("%s.yaml", component))
		bytes, err := yaml.Marshal(definition)

		if err != nil {
			return err
		}

		if err = writeFile(filePath, bytes); err != nil {
			return err
		}
	}

	return nil
}

func pullCustomRepository(directory string, urlPrefix string, dist string, component string) error {
	fmt.Printf("Syncing repository %s/%s @ %s\n", filepath.Base(urlPrefix), component, dist)

	url := fmt.Sprintf("%s/%s/%s", urlPrefix, dist, component)
	filePath := filepath.Join(directory, dist, component)

	_, err := DownloadFile(filepath.Dir(filePath), url, false)
	return err
}

func DumpRepository(directory string, url string, dist string, component string, architecture string, custom bool) error {
	if custom {
		return pullCustomRepository(directory, url, dist, component)
	}

	return pullAptRepository(directory, url, dist, component, architecture)
}
