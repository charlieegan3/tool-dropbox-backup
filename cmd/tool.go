package main

import (
	"context"
	"fmt"
	"log"

	"github.com/charlieegan3/toolbelt-external-job-runner-northflank/pkg/runner"
	"github.com/charlieegan3/toolbelt/pkg/tool"
	"github.com/spf13/viper"

	dbxBackupTool "github.com/charlieegan3/tool-dropbox-backup/pkg/tool"
)

func main() {
	var err error
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config file: %w \n", err)
	}

	tb := tool.NewBelt()
	tb.SetConfig(viper.GetStringMap("tools"))
	tb.AddExternalJobRunner(&runner.Northflank{
		APIToken: viper.GetString("northflank.token"),
	})

	fmt.Println(tb.ExternalJobsFunc())

	err = tb.AddTool(&dbxBackupTool.DropboxBackup{})
	if err != nil {
		log.Fatalf("failed to add tool: %v", err)
	}

	tb.RunJobs(context.Background())
}
