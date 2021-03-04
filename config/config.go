package config

import (
	"io/ioutil"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v2"
)

// Config - the config object we pass around
type Config struct {
	S3       ResourceType `yaml:"s3"`
	IAMUsers ResourceType `yaml:"IAMUsers"`
}

type ResourceType struct {
	IncludeRule FilterRule `yaml:"include"`
	ExcludeRule FilterRule `yaml:"exclude"`
}

type FilterRule struct {
	NamesRE []Expression `yaml:"names_regex"`
}

type Expression struct {
	RE regexp.Regexp
}

// UnmarshalText - Internally used by yaml.Unmarshal to unmarshall an Expression field
func (expression *Expression) UnmarshalText(data []byte) error {
	var pattern string

	if err := yaml.Unmarshal(data, &pattern); err != nil {
		return err
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	expression.RE = *re

	return nil
}

// GetConfig - Unmarshall the config file and parse it into a config object.
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

	err = yaml.Unmarshal(yamlFile, &configObj)
	if err != nil {
		return nil, err
	}

	return &configObj, nil
}

func matchesInclude(name string, includeREs []*regexp.Regexp) bool {
	for _, re := range includeREs {
		if re.MatchString(name) {
			return true
		}
	}
	return false
}

func matchesExclude(name string, excludeREs []*regexp.Regexp) bool {
	for _, re := range excludeREs {
		if re.MatchString(name) {
			return false
		}
	}

	return true
}

// ShouldInclude - Checks if a name should be included according to the inclusion and exclusion rules
func ShouldInclude(name string, includeREs []*regexp.Regexp, excludeNamesREs []*regexp.Regexp) bool {
	shouldInclude := false

	if len(includeREs) > 0 {
		// If any include rules are specified,
		// only check to see if an exclude rule matches when an include rule matches the user
		if matchesInclude(name, includeREs) {
			shouldInclude = matchesExclude(name, excludeNamesREs)
		}
	} else if len(excludeNamesREs) > 0 {
		// Only check to see if an exclude rule matches when there are no include rules defined
		shouldInclude = matchesExclude(name, excludeNamesREs)
	} else {
		shouldInclude = true
	}

	return shouldInclude
}
