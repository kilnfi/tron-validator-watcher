package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kilnfi/tron-validator-watcher/cmd/watcher/app/config"
	clog "github.com/kilnfi/tron-validator-watcher/internal/logger"
	httpserver "github.com/kilnfi/tron-validator-watcher/internal/server/http"
	"github.com/kilnfi/tron-validator-watcher/internal/tron"
	blockwatcher "github.com/kilnfi/tron-validator-watcher/internal/watcher/block"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

var (
	configFile string
	cfg        *config.Config
	logger     *logrus.Logger
	server     *httpserver.Server
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

// start is the entry point for the watcher command.
func start(cmd *cobra.Command, args []string) error {
	// Initialize context and cancel function
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize signal channel for handling interrupts
	ctx, cancel = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)

	// Create a new Prometheus registry and register the metrics
	registry := prometheus.NewRegistry()
	blockMetrics := blockwatcher.NewCollection()
	blockMetrics.MustRegister(registry)

	// Create Tron client
	tronClient, err := createTronClient()
	if err != nil {
		return fmt.Errorf("failed to create Tron client: %w", err)
	}

	// Get the latest block
	block, err := tronClient.Network.GetLatestBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block: %w", err)
	}
	logger.Infof("Latest Block Number: %d", block.BlockHeader.RawData.Number)
	logger.Infof("Latest Block Timestamp (ms): %d", block.BlockHeader.RawData.Timestamp)

	// Get the first block in the round
	firtBlockInRound, err := tron.GetFirstBlockNum(ctx, tronClient, block.BlockHeader.RawData.Number, block.BlockHeader.RawData.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to get first block: %w", err)
	}
	logger.Infof("First Block Timestamp (ms): %d", firtBlockInRound.BlockHeader.RawData.Timestamp)
	logger.Infof("First Block Number: %d", firtBlockInRound.BlockHeader.RawData.Number)

	// Get the account info for each validator
	validators := fetchAccountInfo(tronClient)

	// Starts HTTP server
	if err := startHTTPServer(eg, registry); err != nil {
		return fmt.Errorf("failed to start http server: %w", err)
	}

	// Starts Block watcher
	if cfg.BlockWatcher.Enabled {
		bw, err := blockwatcher.NewBlockWatcher(
			blockwatcher.WithStartBlock(block),
			blockwatcher.WithValidators(validators),
			blockwatcher.WithTronClient(tronClient),
			blockwatcher.WithLogger(logger),
			blockwatcher.WithRefreshInterval(
				cfg.BlockWatcher.RefreshInterval,
			),
			blockwatcher.WithMetrics(blockMetrics),
		)
		if err != nil {
			return fmt.Errorf("failed to create BlockWatcher: %v", err)
		}
		startBlockWatcher(ctx, eg, bw)
	}

	<-ctx.Done()
	logger.Info("shutting down")

	// shutting down HTTP server
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	logger.Info("stopping http server")
	if err := server.Stop(ctx); err != nil {
		logger.Errorf("unable to stop http service: %s", err.Error())
	}

	if err := eg.Wait(); err != nil {
		if errors.Is(err, context.Canceled) {
			logger.Info("Program interrupted by user")
			return nil
		}
		return fmt.Errorf("error during execution: %w", err)
	}

	return nil
}

// createTronClient creates a new Tron client with the given configuration.
func createTronClient() (*tron.Client, error) {
	tronClient, err := tron.NewClient(
		tron.WithBaseURL(viper.GetString("rpc.endpoint")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Tron client: %w", err)
	}
	return tronClient, nil
}

// fetchAccountInfo fetches account information for each validator
// and returns a slice of Account.
func fetchAccountInfo(tronClient *tron.Client) []tron.Account {
	validators := make([]tron.Account, 0, len(cfg.Validators))
	for _, validator := range cfg.Validators {
		account, err := tronClient.Account.GetAccount(validator.Address)
		if err != nil {
			logger.WithField("error", err).Error("Error getting account info")
			os.Exit(1)
		}

		logger.Debugf("Account Address: %v", account.Address)
		logger.Debugf("Account Name: %v", account.AccountName)
		logger.Debugf("Account Witness Address: %v", account.WitnessInfo.Address)
		logger.Debugf("Account Rank: %v", account.WitnessInfo.Rank)

		validators = append(validators, *account)
	}

	return validators
}

// startHTTPServer starts the HTTP server.
func startHTTPServer(eg *errgroup.Group, registry *prometheus.Registry) error {
	var err error

	server, err = httpserver.New(
		registry,
		httpserver.WithHost(cfg.HTTPServer.Host),
		httpserver.WithPort(cfg.HTTPServer.Port),
	)
	if err != nil {
		return fmt.Errorf("unable to create http server: %w", err)
	}

	eg.Go(func() error {
		logger.Infof("starting http server on %s:%d", cfg.HTTPServer.Host, cfg.HTTPServer.Port)

		if err := server.Start(); err != nil {
			return fmt.Errorf("unable to start http server: %w", err)
		}
		return nil
	})

	return nil
}

// startBlockWatcher starts the block watcher.
func startBlockWatcher(ctx context.Context, eg *errgroup.Group, watcher *blockwatcher.BlockWatcher) {
	eg.Go(func() error {
		logger.Info("starting block watcher")

		if err := watcher.Start(ctx); err != nil {
			return fmt.Errorf("unable to start block watcher: %w", err)
		}
		return nil
	})
}

// initLogger initializes the logger with the configuration.
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

// initConfig load the configuration from the file and environment variables.
// It also validates the configuration.
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
