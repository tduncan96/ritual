package codec

import (
	"encoding/json"
)

type JSONCodec struct{}

func (j JSONCodec) Marshal(defs []Definition) ([]byte, error) {
	jsonData, err := json.Marshal(dataFile{Rituals: defs})
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func (j JSONCodec) Unmarshal(blob []byte) ([]Definition, error) {
	var f dataFile
	if err := json.Unmarshal(blob, &f); err != nil {
		return nil, err
	}
	return f.Rituals, nil
}
