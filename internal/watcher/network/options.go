package networkwatcher

import (
	"github.com/kilnfi/tron-validator-watcher/internal/tron"
	"github.com/sirupsen/logrus"
)

type WatcherOptionFunc func(*NetworkWatcher) error

func WithTronClient(client *tron.Client) WatcherOptionFunc {
	return func(c *NetworkWatcher) error {
		c.tronClient = client
		return nil
	}
}

func WithLogger(logger *logrus.Logger) WatcherOptionFunc {
	return func(c *NetworkWatcher) error {
		c.logger = logger
		return nil
	}
}

func WithRefreshInterval(interval int) WatcherOptionFunc {
	return func(c *NetworkWatcher) error {
		c.refreshInterval = interval
		return nil
	}
}

func WithMetrics(metrics *Collection) WatcherOptionFunc {
	return func(c *NetworkWatcher) error {
		c.metrics = metrics
		return nil
	}
}
