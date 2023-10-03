package xdeb

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	version "github.com/knqyf263/go-deb-version"
	"gopkg.in/yaml.v2"
)

type XdebPackageDefinition struct {
	Name         string `yaml:"name"`
	Version      string `yaml:"version"`
	Url          string `yaml:"url"`
	Sha256       string `yaml:"sha256"`
	Path         string `yaml:"path,omitempty"`
	FilePath     string `yaml:"filepath,omitempty"`
	Provider     string `yaml:"provider,omitempty"`
	Distribution string `yaml:"distribution,omitempty"`
	Component    string `yaml:"component,omitempty"`
	IsConfigured bool   `yaml:"is_configured,omitempty"`
}

func (this *XdebPackageDefinition) setPaths(rootPath string) {
	this.Path = filepath.Join(rootPath, this.Name)

	if len(this.Url) > 0 {
		this.FilePath = filepath.Join(this.Path, filepath.Base(this.Url))
	} else {
		this.FilePath = filepath.Join(this.Path, fmt.Sprintf("%s.deb", this.Name))
	}
}

func (this *XdebPackageDefinition) setProvider() {
	if len(this.Provider) == 0 {
		this.Provider = "localhost"
	}
}

func (this *XdebPackageDefinition) setDistribution() {
	if this.Provider == "localhost" {
		this.Distribution = fmt.Sprintf("file:///%s", strings.TrimPrefix(this.FilePath, "/"))
	}
}

func (this *XdebPackageDefinition) setComponent() {
	if strings.HasSuffix(this.Component, ".yaml") {
		this.Component = TrimPathExtension(this.Component)
	}
}

func (this *XdebPackageDefinition) Configure(rootPath string) {
	if !this.IsConfigured {
		this.setPaths(rootPath)
		this.setProvider()
		this.setDistribution()
		this.setComponent()

		this.IsConfigured = true
	}
}

type XdebProviderDefinition struct {
	Xdeb []XdebPackageDefinition `yaml:"xdeb"`
}

func ParseYamlDefinition(path string) (*XdebProviderDefinition, error) {
	yamlFile, err := decompressFile(path)

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

func FindPackage(name string, path string, provider string, distribution string) ([]*XdebPackageDefinition, error) {
	LogMessage("Looking for package %s via provider %s and distribution %s ...", name, provider, distribution)

	globPattern := filepath.Join(path, provider, distribution, "*.yaml.zst")
	globbed, err := filepath.Glob(globPattern)

	if err != nil {
		return nil, err
	}

	if len(globbed) == 0 {
		return nil, fmt.Errorf("No repositories present on the system. Please sync repositories first.")
	}

	packageDefinitions := []*XdebPackageDefinition{}

	for _, match := range globbed {
		definition, err := ParseYamlDefinition(match)

		if err != nil {
			return nil, err
		}

		for _, packageDefinition := range definition.Xdeb {
			if packageDefinition.Name == name {
				packageDefinitions = append(packageDefinitions, PackageDefinitionWithMetadata(&packageDefinition, match))
			}
		}
	}

	if len(packageDefinitions) == 0 {
		return nil, fmt.Errorf("Could not find package %s", name)
	}

	sort.Slice(packageDefinitions, func(i int, j int) bool {
		versionA, err := version.NewVersion(packageDefinitions[i].Version)

		if err != nil {
			return false
		}

		versionB, err := version.NewVersion(packageDefinitions[j].Version)

		if err != nil {
			return false
		}

		return versionA.GreaterThan(versionB)
	})

	return packageDefinitions, nil
}

func getXdebPath() (string, error) {
	xdebPath, err := exec.LookPath("xdeb")

	if err != nil {
		return "", fmt.Errorf("Package xdeb not found. Please install from %s.", XDEB_URL)
	}

	LogMessage("Package xdeb found: %s", xdebPath)
	return xdebPath, nil
}

func convertPackage(path string, xdebArgs string) error {
	if strings.Contains(xdebArgs, "i") {
		xdebArgs = strings.ReplaceAll(xdebArgs, "i", "")
	}

	xdebPath, err := getXdebPath()

	if err != nil {
		return err
	}

	return ExecuteCommand(filepath.Dir(path), xdebPath, xdebArgs, path)
}
