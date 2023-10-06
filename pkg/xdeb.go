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

func (packageDefinition *XdebPackageDefinition) setProvider() {
	if len(packageDefinition.Provider) == 0 {
		if len(packageDefinition.Url) == 0 {
			packageDefinition.Provider = "localhost"
		} else {
			packageDefinition.Provider = "remote"
		}
	}
}

func (packageDefinition *XdebPackageDefinition) setDistribution() {
	if packageDefinition.Provider == "localhost" || packageDefinition.Provider == "remote" {
		packageDefinition.Distribution = "file"
	}
}

func (packageDefinition *XdebPackageDefinition) setComponent() {
	if strings.HasSuffix(packageDefinition.Component, ".yaml") {
		packageDefinition.Component = TrimPathExtension(packageDefinition.Component, 1)
	}
}

func (packageDefinition *XdebPackageDefinition) setPaths(rootPath string) {
	packageDefinition.Path = filepath.Join(rootPath, packageDefinition.Provider, packageDefinition.Distribution, packageDefinition.Component, packageDefinition.Name)

	if len(packageDefinition.Url) > 0 {
		packageDefinition.FilePath = filepath.Join(packageDefinition.Path, filepath.Base(packageDefinition.Url))
	} else {
		packageDefinition.FilePath = filepath.Join(packageDefinition.Path, fmt.Sprintf("%s.deb", packageDefinition.Name))
	}
}

func (packageDefinition *XdebPackageDefinition) Configure(rootPath string) {
	if !packageDefinition.IsConfigured {
		packageDefinition.setProvider()
		packageDefinition.setDistribution()
		packageDefinition.setComponent()
		packageDefinition.setPaths(rootPath)

		packageDefinition.IsConfigured = true
	}
}

func (packageDefinition *XdebPackageDefinition) runPostInstallHooks() error {
	for _, postInstallHook := range packageDefinition.PostInstall {
		for _, command := range postInstallHook.Commands {
			args := []string{}

			if command.Root && os.Getuid() > 0 {
				args = append(args, "sudo")
			}

			args = append(args, strings.Split(command.Command, " ")...)

			LogMessage("Running post-install hook: %s", postInstallHook.Name)

			if err := ExecuteCommand(packageDefinition.Path, args...); err != nil {
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
		return nil, fmt.Errorf("no repositories present on the system, please sync repositories first")
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
		return nil, fmt.Errorf("could not find package '%s'", name)
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

func FindXdeb() (string, error) {
	xdebPath, err := exec.LookPath("xdeb")

	if err != nil {
		return "", fmt.Errorf("xdeb is not installed, please install it via 'xdeb-install xdeb' or manually from '%s'", XDEB_URL)
	}

	LogMessage("Package xdeb found: %s", xdebPath)
	return xdebPath, nil
}

func convertPackage(path string, xdebArgs string) error {
	if strings.Contains(xdebArgs, "i") {
		xdebArgs = strings.ReplaceAll(xdebArgs, "i", "")
	}

	xdebPath, _ := FindXdeb()
	return ExecuteCommand(filepath.Dir(path), xdebPath, xdebArgs, path)
}
