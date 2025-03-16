package networkwatcher

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Collection represents a collection of Prometheus metrics that can be used to monitor the Tron network.
type Collection struct {
	LatestBlock        prometheus.Gauge
	RoundProgress      prometheus.Gauge
	Epoch              prometheus.Gauge
	NextRoundTimestamp prometheus.Gauge
	NumberOfValidators prometheus.Gauge
	RoundDuration      prometheus.Gauge
	MaxBlockPerRound   prometheus.Gauge
}

// NewCollection creates a new Collection with all the metrics initialized.
func NewCollection() *Collection {
	return &Collection{
		LatestBlock: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "latest_block",
			Help:      "The latest block number proposed by the network",
		}),
		RoundProgress: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "round_progress",
			Help:      "The current progress of the round",
		}),
		Epoch: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "epoch",
			Help:      "The current epoch",
		}),
		NextRoundTimestamp: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "next_round_timestamp",
			Help:      "The timestamp of the next round",
		}),
		NumberOfValidators: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "number_of_validators",
			Help:      "The number of active SR in the network",
		}),
		RoundDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "round_duration",
			Help:      "The duration of a round",
		}),
		MaxBlockPerRound: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "max_block_per_round",
			Help:      "The maximum number of blocks per round",
		}),
	}
}

// MustRegister registers all the metrics in the Collection with the provided Prometheus registry.
// This method should be called to ensure that all metrics are properly registered and can be scraped by Prometheus.
func (c *Collection) MustRegister(registry *prometheus.Registry) {
	registry.MustRegister(
		c.LatestBlock,
		c.RoundProgress,
		c.Epoch,
		c.NextRoundTimestamp,
		c.NumberOfValidators,
		c.RoundDuration,
		c.MaxBlockPerRound,
	)
}

// UpdateRoundProgress updates the progress of the current round.
func (c *Collection) UpdateRoundProgress(progress float64) {
	c.RoundProgress.Set(progress)
}

// UpdateSuperRepresentativesNumber updates the number of super representatives.
func (c *Collection) UpdateSuperRepresentativesNumber(number float64) {
	c.NumberOfValidators.Set(number)
}

// UpdateNextRoundTimestamp updates the timestamp of the next round.
func (c *Collection) UpdateNextRoundTimestamp(timestamp float64) {
	c.NextRoundTimestamp.Set(timestamp)
}

// UpdateRoundDuration updates the duration of a round.
func (c *Collection) UpdateRoundDuration(duration float64) {
	c.RoundDuration.Set(duration)
}

// UpdateMaxBlockPerRound updates the maximum number of blocks per round.
func (c *Collection) UpdateMaxBlockPerRound(maxBlock float64) {
	c.MaxBlockPerRound.Set(maxBlock)
}

// UpdateEpoch updates the current epoch.
func (c *Collection) UpdateEpoch(epoch float64) {
	c.Epoch.Set(epoch)
}

// UpdateLatestBlock updates the latest block number.
func (c *Collection) UpdateLatestBlock(block float64) {
	c.LatestBlock.Set(block)
}
