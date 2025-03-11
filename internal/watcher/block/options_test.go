package blockwatcher

import (
	"testing"

	"github.com/kilnfi/tron-validator-watcher/internal/tron"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestWithTronClient(t *testing.T) {
	t.Parallel()

	client := &tron.Client{}
	option := WithTronClient(client)

	watcher := &BlockWatcher{}
	err := option(watcher)

	require.NoError(t, err)
	require.Equal(t, client, watcher.tronClient)
}

func TestWithValidators(t *testing.T) {
	t.Parallel()

	validators := tron.AccountList{
		{
			Address:     "TFKkZiRTYxMud3BrJGQxZxWufwcRrMHKko",
			AccountName: "test",
			WitnessInfo: &tron.Witness{
				Rank:    10,
				Address: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
				IsJobs:  true,
			},
		},
	}
	option := WithValidators(validators)

	watcher := &BlockWatcher{}
	err := option(watcher)

	require.NoError(t, err)
	require.NotNil(t, watcher.validators)
	require.Len(t, watcher.validators, 1)
}

func TestWithLogger(t *testing.T) {
	t.Parallel()

	logger := logrus.New()
	option := WithLogger(logger)

	watcher := &BlockWatcher{}
	err := option(watcher)

	require.NoError(t, err)
	require.Equal(t, logger, watcher.logger)
}

func TestWithStartBlock(t *testing.T) {
	t.Parallel()

	block := &tron.Block{
		BlockHeader: tron.BlockHeader{
			RawData: tron.BlockHeaderRawData{
				Number: 123,
			},
		},
	}

	option := WithStartBlock(block)

	watcher := &BlockWatcher{}
	err := option(watcher)

	require.NoError(t, err)
	require.Equal(t, block, watcher.startBlock)
}

func TestWithRefreshInterval(t *testing.T) {
	t.Parallel()

	option := WithRefreshInterval(10)

	watcher := &BlockWatcher{}
	err := option(watcher)

	require.NoError(t, err)
	require.Equal(t, 10, watcher.refreshInterval)
}

func TestWithMetrics(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	metrics := NewCollection()
	metrics.MustRegister(registry)

	option := WithMetrics(metrics)

	watcher := &BlockWatcher{}
	err := option(watcher)

	require.NoError(t, err)
	require.Equal(t, metrics, watcher.metrics)
}
