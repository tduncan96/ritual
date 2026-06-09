package codec

import (
	sushi "github.com/BurntSushi/toml"
)

type TOMLCodec struct{}

func (t TOMLCodec) Marshal(def Definition) ([]byte, error) {
	tomlData, err := sushi.Marshal(def)
	if err != nil {
		return nil, err
	}
	return tomlData, nil
}

func (t TOMLCodec) Unmarshal(blob []byte) ([]Definition, error) {
	var defs []Definition
	var def Definition
	if err := sushi.Unmarshal(blob, &def); err != nil {
		return defs, err
	}
	return append(defs, def), nil
}
