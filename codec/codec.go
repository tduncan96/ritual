package codec

import (
	"crypto/sha256"
	"fmt"
	"maps"
	"slices"
	"strings"
)

type Definition struct {
	Name     string            `toml:"name"`
	Schedule string            `toml:"schedule"`
	Host     string            `toml:"host"`
	Commands string            `toml:"commands"`
	Env      map[string]string `toml:"env"`
	Hash     string            `toml:"-"`
	Status   bool              `toml:"status"`
}

type Codec interface {
	Marshal([]Definition) ([]byte, error)     // to file
	Unmarshal([]byte) ([]Definition, error) // to struct
}

var Codecs = map[string]Codec{
	"cron": CronCodec{},
	"toml": TOMLCodec{},
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
