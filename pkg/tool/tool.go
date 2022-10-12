package tool

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/charlieegan3/tool-dropbox-backup/pkg/tool/jobs"

	"github.com/Jeffail/gabs/v2"

	"github.com/charlieegan3/toolbelt/pkg/apis"
	"github.com/gorilla/mux"
)

// DropboxBackup is a tool for generating a personal JSON status for my public activities
type DropboxBackup struct {
	config *gabs.Container
	db     *sql.DB

	externalJobsFunc func(job apis.ExternalJob) error
}

func (d *DropboxBackup) Name() string {
	return "dropbox-backup"
}

func (d *DropboxBackup) FeatureSet() apis.FeatureSet {
	return apis.FeatureSet{
		Config:       true,
		Jobs:         true,
		ExternalJobs: true,
	}
}

func (d *DropboxBackup) SetConfig(config map[string]any) error {
	d.config = gabs.Wrap(config)

	return nil
}
func (d *DropboxBackup) Jobs() ([]apis.Job, error) {
	if d.externalJobsFunc == nil {
		return []apis.Job{}, fmt.Errorf("externalJobsFunc not set")
	}

	var j []apis.Job
	var path string
	var ok bool

	// load config to be passed to external job
	path = "jobs.refresh"
	jobConfig, ok := d.config.Path(path).Data().(map[string]interface{})
	if !ok {
		return j, fmt.Errorf("missing required config path: %s", path)
	}

	// load all config
	path = "jobs.refresh.schedule"
	schedule, ok := d.config.Path(path).Data().(string)
	if !ok {
		return j, fmt.Errorf("missing required config path: %s", path)
	}

	return []apis.Job{
		&jobs.Refresh{
			ScheduleOverride:   schedule,
			RunExternalJobFunc: d.externalJobsFunc,
			ExternalJobConfig:  jobConfig,
		},
	}, nil
}

func (d *DropboxBackup) ExternalJobsFuncSet(f func(job apis.ExternalJob) error) {
	d.externalJobsFunc = f
}

func (d *DropboxBackup) DatabaseMigrations() (*embed.FS, string, error) {
	return nil, "migrations", nil
}
func (d *DropboxBackup) DatabaseSet(db *sql.DB)              {}
func (d *DropboxBackup) HTTPPath() string                    { return "json-status" }
func (d *DropboxBackup) HTTPAttach(router *mux.Router) error { return nil }
