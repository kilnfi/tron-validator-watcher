package app

import (
	"context"
	"fmt"
	"os"

	"github.com/kilnfi/tron-validator-watcher/internal/tron"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	logger *logrus.Logger
)

func init() {
	cobra.OnInitialize(initLogger)
}

func NewWatcherCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tron-validator-watcher",
		Short: "Tron Validator Watcher is a tool to monitor a set of validators on Tron",
		Long: `Tron Validator Watcher is a Prometheus exporter designed for monitoring validator nodes on the Tron network.
It collects and exposes key metrics related to validator performance, network health, and block production efficiency.

With this tool, you can track:
- The reliability and efficiency of validators in block production.
- Detailed insights into individual validators.
- Overall network activity and status.`,
		RunE:             start,
		SilenceUsage:     true,
		SilenceErrors:    true,
		TraverseChildren: true,
	}

	return cmd
}

func start(cmd *cobra.Command, args []string) error {
	fmt.Println("Starting Tron Validator Watcher...")

	tronClient, err := createTronClient()
	if err != nil {
		return fmt.Errorf("failed to create Tron client: %w", err)
	}

	block, err := tronClient.Network.GetLatestBlock(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to get latest block: %w", err)
	}
	logger.Infof("Latest Block Number: %d", block.BlockHeader.RawData.Number)
	logger.Infof("Latest Block Timestamp (ms): %d", block.BlockHeader.RawData.Timestamp)

	return nil
}

func createTronClient() (*tron.Client, error) {
	tronClient, err := tron.NewClient(
		tron.WithBaseURL(os.Getenv("TRON_BASE_URL")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Tron client: %w", err)
	}
	return tronClient, nil
}

func initLogger() {
	logger = logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{})

	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logger.SetLevel(logrus.FatalLevel)
	case "panic":
		logger.SetLevel(logrus.PanicLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}
}
