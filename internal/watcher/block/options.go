package blockwatcher

import (
	"github.com/kilnfi/tron-validator-watcher/internal/tron"
	"github.com/sirupsen/logrus"
)

type WatcherOptionFunc func(*BlockWatcher) error

func WithTronClient(client *tron.Client) WatcherOptionFunc {
	return func(c *BlockWatcher) error {
		c.tronClient = client
		return nil
	}
}

func WithValidators(validators tron.AccountList) WatcherOptionFunc {
	return func(c *BlockWatcher) error {
		c.validators = validators
		return nil
	}
}

func WithLogger(logger *logrus.Logger) WatcherOptionFunc {
	return func(c *BlockWatcher) error {
		c.logger = logger
		return nil
	}
}

func WithStartBlock(block *tron.Block) WatcherOptionFunc {
	return func(c *BlockWatcher) error {
		c.startBlock = block
		return nil
	}
}

func WithRefreshInterval(interval int) WatcherOptionFunc {
	return func(c *BlockWatcher) error {
		c.refreshInterval = interval
		return nil
	}
}

func WithMetrics(metrics *Collection) WatcherOptionFunc {
	return func(c *BlockWatcher) error {
		c.metrics = metrics
		return nil
	}
}
