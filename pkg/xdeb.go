package xdeb

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/adrg/xdg"
	version "github.com/knqyf263/go-deb-version"
	"gopkg.in/yaml.v2"
)

type XdebPackagePostInstallCommandDefinition struct {
	Root    bool   `yaml:"root"`
	Command string `yaml:"command"`
}

type XdebPackagePostInstallDefinition struct {
	Name     string                                    `yaml:"name"`
	Commands []XdebPackagePostInstallCommandDefinition `yaml:"commands"`
}

type XdebPackageDefinition struct {
	Name         string                             `yaml:"name"`
	Version      string                             `yaml:"version"`
	Url          string                             `yaml:"url"`
	Sha256       string                             `yaml:"sha256"`
	PostInstall  []XdebPackagePostInstallDefinition `yaml:"post-install,omitempty"`
	Path         string                             `yaml:"path,omitempty"`
	FilePath     string                             `yaml:"filepath,omitempty"`
	Provider     string                             `yaml:"provider,omitempty"`
	Distribution string                             `yaml:"distribution,omitempty"`
	Component    string                             `yaml:"component,omitempty"`
	IsConfigured bool                               `yaml:"is_configured,omitempty"`
}

func (this *XdebPackageDefinition) setProvider() {
	if len(this.Provider) == 0 {
		if len(this.Url) == 0 {
			this.Provider = "localhost"
		} else {
			this.Provider = "remote"
		}
	}
}

func (this *XdebPackageDefinition) setDistribution() {
	if this.Provider == "localhost" || this.Provider == "remote" {
		this.Distribution = "file"
	}
}

func (this *XdebPackageDefinition) setComponent() {
	if strings.HasSuffix(this.Component, ".yaml") {
		this.Component = TrimPathExtension(this.Component, 1)
	}
}

func (this *XdebPackageDefinition) setPaths(rootPath string) {
	this.Path = filepath.Join(rootPath, this.Provider, this.Distribution, this.Component, this.Name)

	if len(this.Url) > 0 {
		this.FilePath = filepath.Join(this.Path, filepath.Base(this.Url))
	} else {
		this.FilePath = filepath.Join(this.Path, fmt.Sprintf("%s.deb", this.Name))
	}
}

func (this *XdebPackageDefinition) Configure(rootPath string) {
	if !this.IsConfigured {
		this.setProvider()
		this.setDistribution()
		this.setComponent()
		this.setPaths(rootPath)

		this.IsConfigured = true
	}
}

func (this *XdebPackageDefinition) runPostInstallHooks() error {
	for _, postInstallHook := range this.PostInstall {
		for _, command := range postInstallHook.Commands {
			args := []string{}

			if command.Root && os.Getuid() > 0 {
				args = append(args, "sudo")
			}

			args = append(args, strings.Split(command.Command, " ")...)

			LogMessage("Running post-install hook: %s", postInstallHook.Name)

			if err := ExecuteCommand(this.Path, args...); err != nil {
				return err
			}
		}
	}

	return nil
}

type XdebProviderDefinition struct {
	Xdeb []*XdebPackageDefinition `yaml:"xdeb"`
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

func FindPackage(name string, path string, provider string, distribution string, exact bool) ([]*XdebPackageDefinition, error) {
	LogMessage("Looking for package %s (exact: %s) via provider %s and distribution %s ...", name, strconv.FormatBool(exact), provider, distribution)

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

		for index := range definition.Xdeb {
			if (exact && definition.Xdeb[index].Name == name) || (!exact && strings.HasPrefix(definition.Xdeb[index].Name, name)) {
				distPath := filepath.Dir(match)

				definition.Xdeb[index].Component = TrimPathExtension(filepath.Base(match), 2)
				definition.Xdeb[index].Distribution = filepath.Base(distPath)
				definition.Xdeb[index].Provider = filepath.Base(filepath.Dir(distPath))

				packageDefinitions = append(packageDefinitions, definition.Xdeb[index])
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

func RepositoryPath() (string, error) {
	arch, err := FindArchitecture()

	if err != nil {
		return "", err
	}

	return filepath.Join(xdg.ConfigHome, APPLICATION_NAME, "repositories", arch), nil
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
