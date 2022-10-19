package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

// Canary will leave a file in dropbox to allow the check job to verify that the
// backup is functioning
type Canary struct {
	ScheduleOverride string
	DropboxToken     string
	DropboxPath      string
}

func (c *Canary) Name() string {
	return "canary"
}

func (c *Canary) Run(ctx context.Context) error {
	doneCh := make(chan bool)
	errCh := make(chan error)

	go func() {

		config := dropbox.Config{
			Token: c.DropboxToken,
		}
		dbx := files.New(config)

		_, err := dbx.Upload(
			&files.UploadArg{
				CommitInfo: files.CommitInfo{
					Mode: &files.WriteMode{
						Tagged: dropbox.Tagged{
							Tag: files.WriteModeOverwrite,
						},
					},
					Path: c.DropboxPath,
				},
			},
			strings.NewReader(time.Now().UTC().Format(time.RFC3339)),
		)
		if err != nil {
			errCh <- fmt.Errorf("error writing canary file: %w", err)
			return
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

func (c *Canary) Timeout() time.Duration {
	return 15 * time.Second
}

func (c *Canary) Schedule() string {
	if c.ScheduleOverride != "" {
		return c.ScheduleOverride
	}
	return "0 */5 * * * *"
}
