package codec

type Definition struct {
	Name     string            `toml:"name"`
	Schedule string            `toml:"schedule"`
	Host     string            `toml:"host"`
	Commands string            `toml:"commands"`
	Env      map[string]string `toml:"env"`
	Status   bool              `toml:"status"`
}

type Codec interface {
	Marshal(Definition) ([]byte, error)     // to file
	Unmarshal([]byte) ([]Definition, error) // to struct
}

var Codecs = map[string]Codec{
	"cron": CronCodec{},
	"toml": TOMLCodec{},
}
