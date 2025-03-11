package main

import (
	cmd "github.com/kilnfi/tron-validator-watcher/cmd/watcher/app"

	clog "github.com/kilnfi/tron-validator-watcher/internal/logger"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&clog.CustomTextFormatter{})
	logger.Level = logrus.DebugLevel

	cmd := cmd.NewWatcherCommand()

	if err := cmd.Execute(); err != nil {
		logger.Fatalf("Error executing command: %v", err)
	}
}
