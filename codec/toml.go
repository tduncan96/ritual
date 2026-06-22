package codec

import (
	sushi "github.com/BurntSushi/toml"
)

type TOMLCodec struct{}

func (t TOMLCodec) Marshal(defs []Definition) ([]byte, error) {
	tomlData, err := sushi.Marshal(dataFile{Rituals: defs})
	if err != nil {
		return nil, err
	}
	return tomlData, nil
}

func (t TOMLCodec) Unmarshal(blob []byte) ([]Definition, error) {
	var f dataFile
	if err := sushi.Unmarshal(blob, &f); err != nil {
		return nil, err
	}
	return f.Rituals, nil
}
