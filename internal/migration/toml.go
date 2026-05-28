package transfer

import (
	"os"
	"path/filepath"
	"time"

	"ritual/internal/db"

	sushi "github.com/BurntSushi/toml"
)

type JobDef struct {
	JobName  string `toml:"name"`
	Schedule string `toml:"schedule"`
	Host     string `toml:"host"`
	JobType  string `toml:"type"`
	Commands string `toml:"commands"`
}

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

	var def JobDef
	if err := sushi.Unmarshal(tomlData, &def); err != nil {
		return 0, err
	}

	job := db.Job{
		JobName:   def.JobName,
		Schedule:  def.Schedule,
		Host:      def.Host,
		JobType:   def.JobType,
		Commands:  def.Commands,
		JobStatus: "Active",
		Created:   time.Now().UTC().Format("2001-02-03 12:34:56"),
		Updated:   time.Now().UTC().Format("2001-02-03 12:34:56"),
		LastRun:   "Never",
		NextRun:   "Never",
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
		def := JobDef{
			JobName:  job.JobName,
			Schedule: job.Schedule,
			Host:     job.Host,
			JobType:  job.JobType,
			Commands: job.Commands,
		}

		tomlData, err := sushi.Marshal(def)
		if err != nil {
			return err
		}

		if err := os.WriteFile(filepath.Join(TomlPath, def.JobName + ".toml"), tomlData, 0644); err != nil {
			return err
		}
	}
	return nil
}
