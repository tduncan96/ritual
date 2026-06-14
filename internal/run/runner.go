package run

import (
	"ritual/internal/db"
)

type Runner interface {
	ExecuteJob(job db.Job) error
}
