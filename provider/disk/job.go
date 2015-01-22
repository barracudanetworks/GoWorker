package disk

import "github.com/barracudanetworks/GoWorker/job"

// DiskJob a job that points back to a disk provider
type DiskJob struct {
	provider *Disk
	conf     *job.JobConfig
}

// Config return this jobs config
func (d *DiskJob) Config() *job.JobConfig {
	return d.conf
}

// JobConfirmer return the provider this job points back to
func (d *DiskJob) JobConfirmer() job.JobConfirmer {
	return d.provider
}
