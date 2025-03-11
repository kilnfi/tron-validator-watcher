package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kilnfi/tron-validator-watcher/cmd/watcher/app/config"
	clog "github.com/kilnfi/tron-validator-watcher/internal/logger"
	"github.com/kilnfi/tron-validator-watcher/internal/tron"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	logger     *logrus.Logger
	configFile string
	cfg        *config.Config
)

func init() {
	cobra.OnInitialize(initLogger)
	cobra.OnInitialize(initConfig)
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

	cmd.Flags().StringVarP(&configFile, "config-file", "", "", "config file (default is config.yml)")
	cmd.Flags().StringP("log-level", "", "info", "log level (debug, info, warn, error, fatal, panic)")
	cmd.Flags().BoolP("block-watcher-enabled", "", true, "enable block watcher")
	cmd.Flags().IntP("block-watcher-refresh-interval", "", 30, "block watcher refresh interval in seconds")
	cmd.Flags().StringP("rpc-endpoint", "", "", "block watcher API URL")
	cmd.Flags().StringP("http-server-host", "", "", "HTTP server host")
	cmd.Flags().IntP("http-server-port", "", 8080, "HTTP server port")

	if err := viper.BindPFlag("log-level", cmd.Flags().Lookup("log-level")); err != nil {
		logger.Fatalf("failed to bind log-level flag: %v", err)
	}
	if err := viper.BindPFlag("block-watcher.enabled", cmd.Flags().Lookup("block-watcher-enabled")); err != nil {
		logger.Fatalf("failed to bind block-watcher-enabled flag: %v", err)
	}
	if err := viper.BindPFlag("block-watcher.refresh-interval", cmd.Flags().Lookup("block-watcher-refresh-interval")); err != nil {
		logger.Fatalf("failed to bind block-watcher-refresh-interval flag: %v", err)
	}
	if err := viper.BindPFlag("rpc.endpoint", cmd.Flags().Lookup("rpc-endpoint")); err != nil {
		logger.Fatalf("failed to bind rpc-endpoint flag: %v", err)
	}
	if err := viper.BindPFlag("http-server.host", cmd.Flags().Lookup("http-server-host")); err != nil {
		logger.Fatalf("failed to bind http-server-host flag: %v", err)
	}
	if err := viper.BindPFlag("http-server.port", cmd.Flags().Lookup("http-server-port")); err != nil {
		logger.Fatalf("failed to bind http-server-port flag: %v", err)
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
		tron.WithBaseURL(viper.GetString("rpc.endpoint")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Tron client: %w", err)
	}
	return tronClient, nil
}

func initLogger() {
	logger = logrus.New()
	logger.SetFormatter(&clog.CustomTextFormatter{})

	switch viper.GetString("log-level") {
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

func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(1)
	}

	cfg = &config.Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		logger.Fatalf("failed to decode config: %v", err)
	}

	// validate the config
	if err := cfg.Validate(); err != nil {
		logger.Fatalf("failed to validate config: %v", err)
	}
}
