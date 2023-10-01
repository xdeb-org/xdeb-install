package xdeb

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
	"gopkg.in/yaml.v2"
)

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

func pullPackagFile(url string, dist string, component string, architecture string) (*XdebProviderDefinition, error) {
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

func DumpPackages(directory string, url string, dist string, component string, architecture string) error {
	definition, err := pullPackagFile(url, dist, component, architecture)

	if err != nil {
		return err
	}

	if definition != nil && len(definition.Xdeb) > 0 {
		filePath := filepath.Join(directory, dist, fmt.Sprintf("%s.yaml", component))
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
