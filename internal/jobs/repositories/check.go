package repositories

// Compile-time interface satisfaction checks.
// These ensure concrete types implement the repository interfaces.

import (
	"example.com/axiomnizam/internal/jobs"
)

var _ JobRepository = (*jobs.JobManagerImpl)(nil)
