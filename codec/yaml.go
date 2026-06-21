package codec

import (
	goccy "github.com/goccy/go-yaml"
)

type YAMLCodec struct{}

type yamlFile struct {
	Rituals []Definition
}

func (y YAMLCodec) Marshal(defs []Definition) ([]byte, error) {
	yamlData, err := goccy.Marshal(yamlFile{Rituals: defs})
	if err != nil {
		return nil, err
	}
	return yamlData, nil
}

func (y YAMLCodec) Unmarshal(blob []byte) ([]Definition, error) {
	var f yamlFile
	if err := goccy.Unmarshal(blob, &f); err != nil {
		return []Definition{}, err
	}
	return f.Rituals, nil
}
