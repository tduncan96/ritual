package jobio

import (
	"os"
	"path/filepath"
	"fmt"

	"ritual/internal/db"

	sushi "github.com/BurntSushi/toml"
)

var TomlPath string = getTomlPath()

func getTomlPath() (path string) {
	tomlPath := os.Getenv("RITUAL_TOML_DUMP")
	if tomlPath == "" {
		tomlPath = "./toml-dump/"
	}

	return tomlPath
}

func GetTomlFiles() ([]string, error) {

	files, err := os.ReadDir(TomlPath)
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

func TomlToJob(file string) error {
	tomlData, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	var job db.Job
	if err := sushi.Unmarshal(tomlData, &job); err != nil {
		return err
	}
	id, err := job.CreateJob()
	if err != nil {
		return err
	}
	fmt.Printf("Imported Job ID %d from file: ", id, file)
	return nil
}

func JobToToml(id int) error {
	job, err := db.GetJob(id)
	if err != nil {
		return err
	}

	tomlData, err := sushi.Marshal(job)
	if err != nil {
		return err
	}
	filename := filepath.Join(TomlPath, job.JobName+".toml")
	if err := os.WriteFile(filename, tomlData, 0644); err != nil {
		return err
	}
	fmt.Printf("Job #%d migrated to file: %v", id, filename)
	return nil
}
