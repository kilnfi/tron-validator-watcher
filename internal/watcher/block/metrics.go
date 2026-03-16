package blockwatcher

import (
	"strconv"

	"github.com/kilnfi/tron-validator-watcher/internal/tron"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type Collection struct {
	ProposedBlocks                     *prometheus.CounterVec
	MissedBlocks                       *prometheus.CounterVec
	ConsecutiveMissedBlocks            *prometheus.GaugeVec
	LatestBlockProcessedByBlockWatcher prometheus.Gauge
	BlockProducerInfo                  *prometheus.GaugeVec
	RoundProgress                      prometheus.Gauge
	Epoch                              prometheus.Gauge
}

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
		ConsecutiveMissedBlocks: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "consecutive_missed_blocks",
			Help:      "Current number of consecutive blocks missed by the validator (resets to 0 when a block is proposed)",
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
	}
}

func (c *Collection) MustRegister(registry *prometheus.Registry) {
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		c.ProposedBlocks,
		c.MissedBlocks,
		c.ConsecutiveMissedBlocks,
		c.LatestBlockProcessedByBlockWatcher,
		c.BlockProducerInfo,
		c.RoundProgress,
		c.Epoch,
	)
}

func (c *Collection) InitMetrics(epoch int, validators tron.AccountList) {
	c.ProposedBlocks.Reset()
	c.MissedBlocks.Reset()
	c.ConsecutiveMissedBlocks.Reset()
	c.BlockProducerInfo.Reset()
	c.RoundProgress.Set(0)
	c.Epoch.Set(float64(epoch))
	c.LatestBlockProcessedByBlockWatcher.Set(0)

	for _, validator := range validators {
		c.ProposedBlocks.WithLabelValues(strconv.Itoa(epoch), validator.AccountName, validator.Address).Add(0)
		c.MissedBlocks.WithLabelValues(strconv.Itoa(epoch), validator.AccountName, validator.Address).Add(0)
		c.ConsecutiveMissedBlocks.WithLabelValues(strconv.Itoa(epoch), validator.AccountName, validator.Address).Set(0)
		c.BlockProducerInfo.WithLabelValues(validator.AccountName, validator.Address, strconv.Itoa(validator.WitnessInfo.Rank)).Set(0)
	}
}

func (c *Collection) UpdateRoundProgress(progress float64) {
	c.RoundProgress.Set(progress)
}
func (c *Collection) UpdateLatestBlockProcessed(blockNumber float64) {
	c.LatestBlockProcessedByBlockWatcher.Set(blockNumber)
}
func (c *Collection) UpdateProposedBlock(epoch int, validatorName, validatorAddress string) {
	c.ProposedBlocks.WithLabelValues(strconv.Itoa(epoch), validatorName, validatorAddress).Inc()
}
func (c *Collection) UpdateMissedBlock(epoch int, validatorName, validatorAddress string) {
	c.MissedBlocks.WithLabelValues(strconv.Itoa(epoch), validatorName, validatorAddress).Inc()
}
func (c *Collection) UpdateConsecutiveMissedBlock(epoch int, validatorName, validatorAddress string, reset bool) {
	if reset {
		c.ConsecutiveMissedBlocks.WithLabelValues(strconv.Itoa(epoch), validatorName, validatorAddress).Set(0)
		return
	}
	c.ConsecutiveMissedBlocks.WithLabelValues(strconv.Itoa(epoch), validatorName, validatorAddress).Inc()
}
func (c *Collection) UpdateBlockProducerInfo(validatorName, validatorAddress string, rank int) {
	c.BlockProducerInfo.WithLabelValues(validatorName, validatorAddress, strconv.Itoa(rank)).Set(1)
}

func (c *Collection) UpdateEpoch(epoch float64) {
	c.Epoch.Set(epoch)
}
