package jobio

import (
	"os"
	"path/filepath"

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

func TomlToJob(file string) (int64, error) {
	tomlData, err := os.ReadFile(file)
	if err != nil {
		return 0, err
	}
	var job db.Job
	if err := sushi.Unmarshal(tomlData, &job); err != nil {
		return 0, err
	}
	id, err := job.CreateJob()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func jobsToToml(ids []int) error {
	jobs, err := db.GetJobs(ids)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		tomlData, err := sushi.Marshal(job)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(TomlPath, job.JobName+".toml"), tomlData, 0644); err != nil {
			return err
		}
	}
	return nil
}
