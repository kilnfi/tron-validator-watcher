package networkwatcher

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func TestUpdateRoundProgress(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	collection := NewCollection()
	collection.MustRegister(registry)

	expectedMetrics, err := os.Open("testdata/round_progress.metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer expectedMetrics.Close()

	collection.UpdateRoundProgress(15)

	err = testutil.GatherAndCompare(registry, expectedMetrics, "tron_validator_watcher_round_progress")
	require.NoError(t, err)
}

func TestUpdateEpoch(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	collection := NewCollection()
	collection.MustRegister(registry)

	expectedMetrics, err := os.Open("testdata/epoch.metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer expectedMetrics.Close()

	collection.UpdateEpoch(100)

	err = testutil.GatherAndCompare(registry, expectedMetrics, "tron_validator_watcher_epoch")
	require.NoError(t, err)
}

func TestNewCollection(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	collection := NewCollection()
	collection.MustRegister(registry)

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
