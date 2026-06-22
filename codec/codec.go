package codec

import (
	"crypto/sha256"
	"fmt"
	"maps"
	"slices"
	"strings"
)

type Definition struct {
	Name     string            `toml:"name" yaml:"name" json:"name"`
	Schedule string            `toml:"schedule" yaml:"schedule" json:"schedule"`
	Host     string            `toml:"host" yaml:"host" json:"host"`
	Commands string            `toml:"commands" yaml:"commands" json:"commands"`
	Env      map[string]string `toml:"env" yaml:"env" json:"env"`
	Hash     string            `toml:"-" yaml:"-" json:"-"`
	Status   bool              `toml:"status" yaml:"status" json:"status"`
}

type Codec interface {
	Marshal([]Definition) ([]byte, error)   // to file
	Unmarshal([]byte) ([]Definition, error) // to []struct
}

var Codecs = map[string]Codec{
	"cron": CronCodec{},
	"toml": TOMLCodec{},
	"yaml": YAMLCodec{},
	"json": JSONCodec{},
}

func GetHash(host, schedule, commands string, lineEnv map[string]string) string {
	h := sha256.New()

	h.Write([]byte(host))
	h.Write([]byte{0})
	h.Write([]byte(schedule))
	h.Write([]byte{0})
	h.Write([]byte(commands))
	h.Write([]byte{0})

	var envStrings []string
	if len(lineEnv) > 0 {
		for _, key := range slices.Sorted(maps.Keys(lineEnv)) {
			envLine := strings.Join([]string{key, lineEnv[key]}, "=")
			envStrings = append(envStrings, envLine)
		}
	} else {
		envStrings = append(envStrings, "")
	}
	for _, line := range envStrings {
		h.Write([]byte(line))
		h.Write([]byte{0})
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}
