package mageutil

import (
	"fmt"
	"os"
	"runtime"

	"gopkg.in/yaml.v3"
)

const (
	StartConfigFile = "start-config.yml"
)

var (
	serviceBinaries    map[string]int
	toolBinaries       []string
	MaxFileDescriptors int
)

type Config struct {
	ServiceBinaries    map[string]int `yaml:"serviceBinaries"`
	ToolBinaries       []string       `yaml:"toolBinaries"`
	MaxFileDescriptors int            `yaml:"maxFileDescriptors"`
}

func InitForSSC() error {
	yamlFile, err := os.ReadFile(StartConfigFile)
	if err != nil {
		PrintRed("error reading YAML file: " + err.Error())
		return fmt.Errorf("error reading YAML file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		PrintRed("error unmarshalling YAML: " + err.Error())
		return fmt.Errorf("error unmarshalling YAML: %w", err)
	}

	adjustedBinaries := make(map[string]int)
	for binary, count := range config.ServiceBinaries {
		if runtime.GOOS == "windows" {
			binary += ".exe"
		}
		adjustedBinaries[binary] = count
	}

	var adjustedToolsBinaries []string
	for _, tool := range config.ToolBinaries {
		if runtime.GOOS == "windows" {
			tool += ".exe"
		}
		adjustedToolsBinaries = append(adjustedToolsBinaries, tool)
	}
	serviceBinaries = adjustedBinaries
	toolBinaries = adjustedToolsBinaries
	MaxFileDescriptors = config.MaxFileDescriptors
	return nil
}
