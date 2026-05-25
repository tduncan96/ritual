package internal

import (
	"os"
)

type JobDef struct {
	ID       int    `toml:"job.id"`
	JobName  string `toml:"job.name"`
	Schedule string `toml:"job.schedule"`
	Host     string `toml:"job.host"`
	JobType  string `toml:"job.type"`
	Commands string `toml:"job.commands"`
}

func GetTomlFiles() ([]string, error) {
	tomlPath := os.Getenv("RITUAL_TOML_DUMP")
	if tomlPath == "" {
		tomlPath = "./toml-dump"
	}

	files, err := os.ReadDir(tomlPath)
	if err != nil {
		return nil, err
	}

	var fileList []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileName := file.Name()
		fileList = append(fileList, fileName)
	}

	return fileList, nil

}

