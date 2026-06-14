package run

import "ritual/internal/db"

type RemoteRunner struct {}

func (r RemoteRunner) ExecuteJob(job db.Job) error {
	return nil
}