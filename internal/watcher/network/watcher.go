package networkwatcher

import (
	"context"
	"fmt"
	"time"

	"github.com/kilnfi/tron-validator-watcher/internal/tron"
	"github.com/sirupsen/logrus"
)

type NetworkWatcher struct {
	tronClient      *tron.Client
	logger          *logrus.Logger
	refreshInterval int
	metrics         *Collection
}

func NewNetworkWatcher(options ...WatcherOptionFunc) (*NetworkWatcher, error) {
	w := &NetworkWatcher{}
	for _, fn := range options {
		if fn == nil {
			continue
		}
		if err := fn(w); err != nil {
			return nil, err
		}
	}

	return w, nil
}

func (w *NetworkWatcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(w.refreshInterval) * time.Second)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("stopping watcher")
			return fmt.Errorf("context done in watcher: %w", ctx.Err())
		case <-ticker.C:
			if err := w.start(ctx); err != nil {
				return fmt.Errorf("NetworkWatcher: failed to collect network stats: %w", err)
			}
			w.logger.WithField("service", "network-watcher").Infof("😴 Sleeping %ds before next iteration...", w.refreshInterval)
		}
	}
}

func (w *NetworkWatcher) start(ctx context.Context) error {
	if err := w.collectLatestBlock(ctx); err != nil {
		return fmt.Errorf("NetworkWatcher: failed to collect latest block: %w", err)
	}

	if err := w.collectRoundInfo(ctx); err != nil {
		return fmt.Errorf("NetworkWatcher: failed to collect round info: %w", err)
	}

	w.collectValidatorInfo()

	return nil
}

func (w *NetworkWatcher) collectLatestBlock(ctx context.Context) error {
	block, err := w.tronClient.Network.GetLatestBlock(ctx)
	if err != nil {
		return fmt.Errorf("NetworkWatcher: failed to get latest block: %w", err)
	}
	w.metrics.UpdateLatestBlock(float64(block.BlockHeader.RawData.Number))
	return nil
}

func (w *NetworkWatcher) collectRoundInfo(ctx context.Context) error {
	block, err := w.tronClient.Network.GetLatestBlock(ctx)
	if err != nil {
		return fmt.Errorf("NetworkWatcher: failed to get latest block: %w", err)
	}
	epoch := tron.GetEpochID(block)
	w.metrics.UpdateEpoch(float64(epoch))

	nextRoundTimestamp := tron.GetNextRound(block)
	w.metrics.UpdateNextRoundTimestamp(float64(nextRoundTimestamp))

	roundProgress := int((block.BlockHeader.RawData.Timestamp / 1000) % tron.RoundDuration)
	w.metrics.UpdateRoundProgress(float64(roundProgress))

	w.metrics.UpdateRoundDuration(float64(tron.RoundDuration))

	return nil
}

func (w NetworkWatcher) collectValidatorInfo() {
	w.metrics.UpdateSuperRepresentativesNumber(tron.NumberOfValidators)
	w.metrics.UpdateMaxBlockPerRound(float64(tron.BlocksPerRound))
}
