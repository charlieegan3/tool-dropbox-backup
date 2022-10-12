package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/charlieegan3/toolbelt/pkg/apis"
)

// Refresh will refresh the contents of the backblaze backup
type Refresh struct {
	ScheduleOverride string

	RunExternalJobFunc func(job apis.ExternalJob) error
	ExternalJobConfig  map[string]interface{}
}

func (r *Refresh) Name() string {
	return "refresh"
}

func (r *Refresh) Run(ctx context.Context) error {
	doneCh := make(chan bool)
	errCh := make(chan error)

	go func() {
		err := r.RunExternalJobFunc(&refreshOnNorthflank{
			cfg: r.ExternalJobConfig,
		})
		if err != nil {
			errCh <- fmt.Errorf("failed to run external job: %w", err)
		}

		doneCh <- true
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-errCh:
		return fmt.Errorf("job failed with error: %s", e)
	case <-doneCh:
		return nil
	}
}

func (r *Refresh) Timeout() time.Duration {
	return 15 * time.Second
}

func (r *Refresh) Schedule() string {
	if r.ScheduleOverride != "" {
		return r.ScheduleOverride
	}
	return "0 */5 * * * *"
}

type refreshOnNorthflank struct {
	cfg map[string]any
}

func (r *refreshOnNorthflank) Name() string {
	return "refresh"
}

func (r *refreshOnNorthflank) RunnerName() string {
	return "northflank"
}

func (r *refreshOnNorthflank) Config() map[string]any {
	return r.cfg
}
