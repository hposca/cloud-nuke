package config

import (
	"io/ioutil"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v2"
)

// RawConfig - used to unmarshall the raw config file
type RawConfig struct {
	S3       rawResourceType `yaml:"s3"`
	IAMUsers rawResourceType `yaml:"IAMUsers"`
}

type rawResourceType struct {
	IncludeRule rawFilterRule `yaml:"include"`
	ExcludeRule rawFilterRule `yaml:"exclude"`
}

type rawNamesRE []string

type rawFilterRule struct {
	NamesRE rawNamesRE `yaml:"names_regex"`
}

// Config - the config object we pass around
// that is a parsed version of RawConfig
type Config struct {
	S3       ResourceType
	IAMUsers ResourceType
}

// ResourceType - the include and exclude
// rules for a resource type
type ResourceType struct {
	IncludeRule FilterRule
	ExcludeRule FilterRule
}

type compiledNamesRE []*regexp.Regexp

// FilterRule - contains regular expressions or plain text patterns
// used to match against a resource type's properties
type FilterRule struct {
	NamesRE compiledNamesRE
}

// association - an ancillary internal struct that is used to represent a 1:1
// association between compiledNamesRE and a RawREs.
// This is used o indicate which compiled regular expression should be
// associated with the raw ones retrieved from the configuration file.
type association struct {
	CompiledREs *compiledNamesRE
	RawREs      rawNamesRE
}

// appendRegex - ancillary function to compile and append regular expressions
// into the ConfigObj's right location
func appendRegex(configNamesRE *compiledNamesRE, patterns rawNamesRE) error {
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}

		*configNamesRE = append(*configNamesRE, re)
	}

	return nil
}

// GetConfig - unmarshall the raw config file
// and parse it into a config object.
func GetConfig(filePath string) (*Config, error) {
	var configObj Config

	absolutePath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	yamlFile, err := ioutil.ReadFile(absolutePath)
	if err != nil {
		return nil, err
	}

	rawConfig := RawConfig{}

	err = yaml.Unmarshal(yamlFile, &rawConfig)
	if err != nil {
		return nil, err
	}

	associations := []association{
		association{&configObj.S3.IncludeRule.NamesRE, rawConfig.S3.IncludeRule.NamesRE},
		association{&configObj.S3.ExcludeRule.NamesRE, rawConfig.S3.ExcludeRule.NamesRE},
		association{&configObj.IAMUsers.IncludeRule.NamesRE, rawConfig.IAMUsers.IncludeRule.NamesRE},
		association{&configObj.IAMUsers.ExcludeRule.NamesRE, rawConfig.IAMUsers.ExcludeRule.NamesRE},
	}

	for _, association := range associations {
		if err := appendRegex(association.CompiledREs, association.RawREs); err != nil {
			return nil, err
		}
	}

	return &configObj, nil
}
