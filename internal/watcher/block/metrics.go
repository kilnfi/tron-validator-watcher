package blockwatcher

import (
	"strconv"

	"github.com/kilnfi/tron-validator-watcher/internal/tron"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// Collection represents a collection of Prometheus metrics used by the block watcher.
type Collection struct {
	ProposedBlocks                     *prometheus.CounterVec
	MissedBlocks                       *prometheus.CounterVec
	ConsecutiveMissedBlocks            *prometheus.CounterVec
	LatestBlockProcessedByBlockWatcher prometheus.Gauge
	BlockProducerInfo                  *prometheus.GaugeVec
}

// NewCollection creates a new Collection with all the metrics needed by the block watcher.
func NewCollection() *Collection {
	return &Collection{
		ProposedBlocks: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "tron_validator_watcher",
			Name:      "proposed_blocks_total",
			Help:      "Total number of blocks proposed by the validator",
		}, []string{"epoch", "validator_name", "validator_address"}),
		MissedBlocks: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "tron_validator_watcher",
			Name:      "missed_blocks_total",
			Help:      "Total number of blocks missed by the validator",
		}, []string{"epoch", "validator_name", "validator_address"}),
		ConsecutiveMissedBlocks: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "tron_validator_watcher",
			Name:      "consecutive_missed_blocks_total",
			Help:      "Total number of consecutive blocks missed by the validator",
		}, []string{"epoch", "validator_name", "validator_address"}),
		LatestBlockProcessedByBlockWatcher: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "latest_block_processed_by_block_watcher",
			Help:      "The latest block processed by the block watcher",
		}),
		BlockProducerInfo: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "block_producer_info",
			Help:      "Block producer info",
		}, []string{"validator_name", "validator_address", "rank"}),
	}
}

// MustRegister registers all the metrics in the Collection with the provided Prometheus registry.
// This method should be called to ensure that all metrics are properly registered and can be scraped by Prometheus.
func (c *Collection) MustRegister(registry *prometheus.Registry) {
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		c.ProposedBlocks,
		c.MissedBlocks,
		c.ConsecutiveMissedBlocks,
		c.LatestBlockProcessedByBlockWatcher,
		c.BlockProducerInfo,
	)
}

// InitMetrics initializes the metrics for the given epoch and validators.
func (c *Collection) InitMetrics(epoch int, validators tron.AccountList) {
	c.ProposedBlocks.Reset()
	c.MissedBlocks.Reset()
	c.ConsecutiveMissedBlocks.Reset()
	c.BlockProducerInfo.Reset()
	c.LatestBlockProcessedByBlockWatcher.Set(0)

	for _, validator := range validators {
		c.ProposedBlocks.WithLabelValues(strconv.Itoa(epoch), validator.AccountName, validator.Address).Add(0)
		c.MissedBlocks.WithLabelValues(strconv.Itoa(epoch), validator.AccountName, validator.Address).Add(0)
		c.ConsecutiveMissedBlocks.WithLabelValues(strconv.Itoa(epoch), validator.AccountName, validator.Address).Add(0)
		c.BlockProducerInfo.WithLabelValues(validator.AccountName, validator.Address, strconv.Itoa(validator.WitnessInfo.Rank)).Set(0)
	}
}

// UpdateLatestBlockProcessed updates the latest block processed by the block watcher.
func (c *Collection) UpdateLatestBlockProcessed(blockNumber float64) {
	c.LatestBlockProcessedByBlockWatcher.Set(blockNumber)
}

// UpdateProposedBlock updates the number of proposed blocks by the validator.
func (c *Collection) UpdateProposedBlock(epoch int, validatorName, validatorAddress string) {
	c.ProposedBlocks.WithLabelValues(strconv.Itoa(epoch), validatorName, validatorAddress).Inc()
}

// UpdateMissedBlock updates the number of missed blocks by the validator.
func (c *Collection) UpdateMissedBlock(epoch int, validatorName, validatorAddress string) {
	c.MissedBlocks.WithLabelValues(strconv.Itoa(epoch), validatorName, validatorAddress).Inc()
}

// UpdateConsecutiveMissedBlock updates the number of consecutive missed blocks by the validator.
// If reset is true, the counter is reset to 0.
func (c *Collection) UpdateConsecutiveMissedBlock(epoch int, validatorName, validatorAddress string, reset bool) {
	if reset {
		c.ConsecutiveMissedBlocks.WithLabelValues(strconv.Itoa(epoch), validatorName, validatorAddress).Add(0)
		return
	}
	c.ConsecutiveMissedBlocks.WithLabelValues(strconv.Itoa(epoch), validatorName, validatorAddress).Inc()
}

// UpdateBlockProducerInfo updates the block producer info.
func (c *Collection) UpdateBlockProducerInfo(validatorName, validatorAddress string, rank int) {
	c.BlockProducerInfo.WithLabelValues(validatorName, validatorAddress, strconv.Itoa(rank)).Set(1)
}
