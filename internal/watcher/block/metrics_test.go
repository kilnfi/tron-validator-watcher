package blockwatcher

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kilnfi/tron-validator-watcher/internal/tron"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func TestUpdateRoundProgress(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	collection := NewCollection()
	collection.MustRegister(registry)

	collection.UpdateRoundProgress(15)

	expectedMetrics, err := os.Open("testdata/round_progress.metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer expectedMetrics.Close()

	err = testutil.GatherAndCompare(registry, expectedMetrics, "tron_validator_watcher_round_progress")
	require.NoError(t, err)
}

func TestUpdateLatestBlockProcessed(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	collection := NewCollection()
	collection.MustRegister(registry)

	collection.UpdateLatestBlockProcessed(70089644)
	expectedMetrics, err := os.Open("testdata/latest_block_processed.metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer expectedMetrics.Close()

	err = testutil.GatherAndCompare(registry, expectedMetrics, "tron_validator_watcher_latest_block_processed_by_block_watcher")
	require.NoError(t, err)
}

func TestUpdateProposedBlock(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	collection := NewCollection()
	collection.MustRegister(registry)

	collection.ProposedBlocks.WithLabelValues("100", "test", "fakeaddress").Add(0)
	collection.UpdateProposedBlock(100, "test", "fakeaddress")
	expectedMetrics, err := os.Open("testdata/proposed_block.metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer expectedMetrics.Close()

	err = testutil.GatherAndCompare(registry, expectedMetrics, "tron_validator_watcher_proposed_blocks_total")
	require.NoError(t, err)
}

func TestUpdateMissedBlock(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	collection := NewCollection()
	collection.MustRegister(registry)

	collection.MissedBlocks.WithLabelValues("100", "test", "fakeaddress").Add(0)
	collection.UpdateMissedBlock(100, "test", "fakeaddress")
	expectedMetrics, err := os.Open("testdata/missed_block.metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer expectedMetrics.Close()

	err = testutil.GatherAndCompare(registry, expectedMetrics, "tron_validator_watcher_missed_blocks_total")
	require.NoError(t, err)
}

func TestUpdateConsecutiveMissedBlock(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	collection := NewCollection()
	collection.MustRegister(registry)

	collection.ConsecutiveMissedBlocks.WithLabelValues("100", "test", "fakeaddress").Add(0)
	collection.UpdateConsecutiveMissedBlock(100, "test", "fakeaddress", false)
	expectedMetrics, err := os.Open("testdata/consecutive_missed_blocks.metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer expectedMetrics.Close()

	err = testutil.GatherAndCompare(registry, expectedMetrics, "tron_validator_watcher_consecutive_missed_blocks_total")
	require.NoError(t, err)
}

func TestUpdateBlockProducerInfo(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	collection := NewCollection()
	collection.MustRegister(registry)

	collection.UpdateBlockProducerInfo("test", "fakeaddress", 10)
	expectedMetrics, err := os.Open("testdata/block_producer_info.metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer expectedMetrics.Close()

	err = testutil.GatherAndCompare(registry, expectedMetrics, "tron_validator_watcher_block_producer_info")
	require.NoError(t, err)
}

func TestUpdateEpoch(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	collection := NewCollection()
	collection.MustRegister(registry)

	collection.UpdateEpoch(100)
	expectedMetrics, err := os.Open("testdata/epoch.metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer expectedMetrics.Close()

	err = testutil.GatherAndCompare(registry, expectedMetrics, "tron_validator_watcher_epoch")
	require.NoError(t, err)
}

func TestInitMetrics(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	collection := NewCollection()
	collection.MustRegister(registry)

	validators := tron.AccountList{
		{
			Address:     "val1.address",
			AccountName: "val1",
			WitnessInfo: &tron.Witness{
				Rank: 10,
			},
		},
		{
			Address:     "val2.address",
			AccountName: "val2",
			WitnessInfo: &tron.Witness{
				Rank: 20,
			},
		},
	}

	collection.InitMetrics(100, validators)

	metrics, err := registry.Gather()
	require.NoError(t, err)

	var metricNames []string
	for _, m := range metrics {
		if strings.HasPrefix(m.GetName(), "tron_validator_watcher") {
			metricNames = append(metricNames, m.GetName())
		}
	}

	expectedMetrics, err := os.Open("testdata/init_metrics.metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer expectedMetrics.Close()

	err = testutil.GatherAndCompare(registry, expectedMetrics, metricNames...)
	require.NoError(t, err)
}

func TestNewCollection(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	collection := NewCollection()
	collection.MustRegister(registry)

	validators := tron.AccountList{
		{
			Address:     "val1.address",
			AccountName: "val1",
			WitnessInfo: &tron.Witness{
				Rank: 10,
			},
		},
		{
			Address:     "val2.address",
			AccountName: "val2",
			WitnessInfo: &tron.Witness{
				Rank: 20,
			},
		},
	}
	collection.InitMetrics(100, validators)

	metrics, err := registry.Gather()
	if err != nil {
		t.Fatal(err)
	}

	var metricNames []string
	for _, m := range metrics {
		if strings.HasPrefix(m.GetName(), "tron_validator_watcher") {
			metricNames = append(metricNames, m.GetName())
		}
	}

	expectedMetricsCount := reflect.TypeOf(Collection{}).NumField()
	require.Len(t, metricNames, expectedMetricsCount)
}
