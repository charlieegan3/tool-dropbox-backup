package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kurin/blazer/b2"
)

// Check will make sure that the dropbox backup is still working
type Check struct {
	ScheduleOverride string
	BackblazeAccount string
	BackblazeKey     string
	BackblazeBucket  string
	BackblazePath    string
	AlertEndpoint    string
}

func (c *Check) Name() string {
	return "check"
}

func (c *Check) Run(ctx context.Context) error {
	doneCh := make(chan bool)
	errCh := make(chan error)

	go func() {
		var err error

		defer func() {
			if err != nil {
				requestError := alert(c.AlertEndpoint, "json-status: Check Error", err.Error())
				if requestError != nil {
					errCh <- fmt.Errorf("failed to alert on error %s: %w", err.Error(), requestError)
				}
				errCh <- err
			}
		}()

		ctx := context.Background()

		b2, err := b2.NewClient(ctx, c.BackblazeAccount, c.BackblazeKey)
		if err != nil {
			errCh <- fmt.Errorf("failed to authorize account: %w", err)
			return
		}

		if err != nil {
			errCh <- fmt.Errorf("error creating b2 client: %w", err)
			return
		}

		bucket, err := b2.Bucket(ctx, c.BackblazeBucket)
		if err != nil {
			errCh <- fmt.Errorf("error getting bucket: %w", err)
			return
		}

		obj := bucket.Object(c.BackblazePath)
		if err != nil {
			errCh <- fmt.Errorf("error downloading file: %w", err)
			return
		}
		reader := obj.NewReader(ctx)

		buf := new(strings.Builder)
		_, err = io.Copy(buf, reader)
		if err != nil {
			errCh <- fmt.Errorf("error reading file: %w", err)
			return
		}

		t, err := time.Parse(time.RFC3339, buf.String())
		if err != nil {
			errCh <- fmt.Errorf("error parsing time: %w", err)
			return
		}

		if time.Now().UTC().Sub(t) > 5*24*time.Hour {
			errCh <- fmt.Errorf("canary file too old")
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

func (c *Check) Timeout() time.Duration {
	return 15 * time.Second
}

func (c *Check) Schedule() string {
	if c.ScheduleOverride != "" {
		return c.ScheduleOverride
	}
	return "0 */5 * * * *"
}

func alert(webhookRSSEndpoint, title, message string) error {
	datab := []map[string]string{
		{
			"title": title,
			"body":  message,
			"url":   "",
		},
	}

	b, err := json.Marshal(datab)
	if err != nil {
		return fmt.Errorf("failed to form alert item JSON: %s", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", webhookRSSEndpoint, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("failed to build request for alert item: %s", err)
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request for alert item: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send request: non 200OK response")
	}

	return nil
}
