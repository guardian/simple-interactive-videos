package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type TranscodeSet struct {
	PresetId  string `yaml:"presetId"`
	Suffix    string `yaml:"suffix"`
	Extension string `yaml:"extension"`
}

func LoadTranscodeSet(filepath *string) (*[]TranscodeSet, error) {
	f, err := os.Open(*filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rawContent, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var set []TranscodeSet
	err = yaml.Unmarshal(rawContent, &set)
	if err != nil {
		return nil, err
	}

	return &set, nil
}
