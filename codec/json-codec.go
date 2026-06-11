package codec

import (
	"encoding/json"
)

type JSONCodec struct {}

type jsonFile struct {
	Rituals []Definition
}

func (j JSONCodec) Marshal(defs []Definition) ([]byte, error) {
	jsonData, err := json.Marshal(jsonFile{Rituals: defs})
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func (j JSONCodec) Unmarshal(blob []byte) ([]Definition, error) {
	var f jsonFile
	if err := json.Unmarshal(blob, &f); err != nil {
		return []Definition{}, err
	}
	return f.Rituals, nil
}