package codec

import (
	goccy "github.com/goccy/go-yaml"
)

type YAMLCodec struct{}

func (y YAMLCodec) Marshal(defs []Definition) ([]byte, error) {
	yamlData, err := goccy.Marshal(dataFile{Rituals: defs})
	if err != nil {
		return nil, err
	}
	return yamlData, nil
}

func (y YAMLCodec) Unmarshal(blob []byte) ([]Definition, error) {
	var f dataFile
	if err := goccy.Unmarshal(blob, &f); err != nil {
		return nil, err
	}
	return f.Rituals, nil
}
