package codec

import (
	sushi "github.com/BurntSushi/toml"
)

type TOMLCodec struct{}

type tomlFile struct {
	Rituals []Definition
}

func (t TOMLCodec) Marshal(defs []Definition) ([]byte, error) {
	tomlData, err := sushi.Marshal(tomlFile{Rituals: defs})
	if err != nil {
		return nil, err
	}
	return tomlData, nil
}

func (t TOMLCodec) Unmarshal(blob []byte) ([]Definition, error) {
	var defs []Definition
	var f tomlFile
	if err := sushi.Unmarshal(blob, &f); err != nil {
		return defs, err
	}
	return f.Rituals, nil
}
