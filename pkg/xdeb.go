package xdeb

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/yargevad/filepathx"
	"gopkg.in/yaml.v2"
)

const XDEB_URL = "https://github.com/toluschr/xdeb/releases"

type XdebPackageDefinition struct {
	Name         string `yaml:"name"`
	Version      string `yaml:"version"`
	Url          string `yaml:"url"`
	Sha256       string `yaml:"sha256"`
	Path         string `yaml:"path,omitempty"`
	Provider     string `yaml:"provider,omitempty"`
	Distribution string `yaml:"distribution,omitempty"`
	Component    string `yaml:"component,omitempty"`
}

type XdebProviderDefinition struct {
	Xdeb []XdebPackageDefinition `yaml:"xdeb"`
}

func ParseYamlDefinition(path string) (*XdebProviderDefinition, error) {
	yamlFile, err := os.ReadFile(path)

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

func FindPackage(name string, path string) (*XdebPackageDefinition, error) {
	globPattern := filepath.Join(path, "**", "*.yaml")
	globbed, err := filepathx.Glob(globPattern)

	if err != nil {
		return nil, err
	}

	for _, match := range globbed {
		definition, err := ParseYamlDefinition(match)

		if err != nil {
			return nil, err
		}

		for _, packageDefinition := range definition.Xdeb {
			if packageDefinition.Name == name {
				return PackageDefinitionWithMetadata(&packageDefinition, match), nil
			}
		}
	}

	return nil, fmt.Errorf("Could not find package %s", name)
}

func getXdebPath() (string, error) {
	xdebPath, err := exec.LookPath("xdeb")

	if err != nil {
		return "", fmt.Errorf("Package xdeb not found. Please install from %s.", XDEB_URL)
	}

	log.Printf("Package xdeb found: %s", xdebPath)
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
