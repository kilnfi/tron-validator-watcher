package blockwatcher

import (
	"strconv"

	"github.com/kilnfi/tron-validator-watcher/internal/tron"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// sunPerTRX is the number of SUN in one TRX (TRON's base unit).
const sunPerTRX = 1_000_000

type Collection struct {
	ProposedBlocks                     *prometheus.CounterVec
	MissedBlocks                       *prometheus.CounterVec
	ConsecutiveMissedBlocks            *prometheus.GaugeVec
	LatestBlockProcessedByBlockWatcher prometheus.Gauge
	BlockProducerInfo                  *prometheus.GaugeVec
	Rank                               *prometheus.GaugeVec
	Votes                              *prometheus.GaugeVec
	VotesPercentage                    *prometheus.GaugeVec
	VotesMarginToSR                    *prometheus.GaugeVec
	Balance                            *prometheus.GaugeVec
	RewardBalance                      *prometheus.GaugeVec
	RewardsClaimable                   *prometheus.GaugeVec
	AccountActivated                   *prometheus.GaugeVec
	FrozenBalance                      *prometheus.GaugeVec
	WitnessTotalProduced               *prometheus.GaugeVec
	WitnessTotalMissed                 *prometheus.GaugeVec
	Brokerage                          *prometheus.GaugeVec
	ActiveSRCount                      prometheus.Gauge
	NextMaintenanceTime                prometheus.Gauge
	NodeUp                             prometheus.Gauge
	NodeHeadBlock                      prometheus.Gauge
	RoundProgress                      prometheus.Gauge
	Epoch                              prometheus.Gauge
}

func NewCollection() *Collection {
	validatorLabels := []string{"validator_name", "validator_address"}
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
		Rank: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "rank",
			Help:      "The current ranking of the validator among all Super Representatives (1 = highest vote count). A validator is an active SR while its rank is within the active SR count.",
		}, validatorLabels),
		Votes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "votes",
			Help:      "The current number of votes received by the validator",
		}, validatorLabels),
		VotesPercentage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "votes_percentage",
			Help:      "The validator's share of the total network votes, in percent",
		}, validatorLabels),
		VotesMarginToSR: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "votes_margin_to_sr",
			Help:      "Votes separating the validator from the active SR cutoff: positive = cushion above the last SR seat, negative = votes needed to regain a seat",
		}, validatorLabels),
		Balance: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "balance",
			Help:      "The validator account balance, in TRX",
		}, validatorLabels),
		RewardBalance: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "reward_balance",
			Help:      "The validator's claimable (unwithdrawn) rewards, in TRX",
		}, validatorLabels),
		RewardsClaimable: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "rewards_claimable",
			Help:      "Whether the validator has claimable rewards (1) or not (0)",
		}, validatorLabels),
		AccountActivated: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "account_activated",
			Help:      "Whether the validator account is activated on-chain (1) or not (0)",
		}, validatorLabels),
		FrozenBalance: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "frozen_balance",
			Help:      "The validator's staked (frozen v2) TRX",
		}, validatorLabels),
		WitnessTotalProduced: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "witness_total_produced_blocks",
			Help:      "Lifetime number of blocks produced by the validator, as reported by the node",
		}, validatorLabels),
		WitnessTotalMissed: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "witness_total_missed_blocks",
			Help:      "Lifetime number of blocks missed by the validator, as reported by the node",
		}, validatorLabels),
		Brokerage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "brokerage",
			Help:      "The validator's brokerage (commission) rate, in percent",
		}, validatorLabels),
		ActiveSRCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "active_srs",
			Help:      "The number of active Super Representatives on the network",
		}),
		NextMaintenanceTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "next_maintenance_time",
			Help:      "Unix timestamp (seconds) of the next maintenance period, when the SR set is reshuffled",
		}),
		NodeUp: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "node_up",
			Help:      "Whether the TRON node queried by the watcher is reachable (1) or not (0)",
		}),
		NodeHeadBlock: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "tron_validator_watcher",
			Name:      "node_head_block",
			Help:      "The latest block number reported by the TRON node (chain head)",
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
		c.Rank,
		c.Votes,
		c.VotesPercentage,
		c.VotesMarginToSR,
		c.Balance,
		c.RewardBalance,
		c.RewardsClaimable,
		c.AccountActivated,
		c.FrozenBalance,
		c.WitnessTotalProduced,
		c.WitnessTotalMissed,
		c.Brokerage,
		c.ActiveSRCount,
		c.NextMaintenanceTime,
		c.NodeUp,
		c.NodeHeadBlock,
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
	c.ActiveSRCount.Set(0)
	c.NextMaintenanceTime.Set(0)
	c.NodeUp.Set(0)
	c.NodeHeadBlock.Set(0)

	for _, validator := range validators {
		c.ProposedBlocks.WithLabelValues(strconv.Itoa(epoch), validator.AccountName, validator.Address).Add(0)
		c.MissedBlocks.WithLabelValues(strconv.Itoa(epoch), validator.AccountName, validator.Address).Add(0)
		c.ConsecutiveMissedBlocks.WithLabelValues(strconv.Itoa(epoch), validator.AccountName, validator.Address).Set(0)
		c.BlockProducerInfo.WithLabelValues(validator.AccountName, validator.Address, strconv.Itoa(validator.WitnessInfo.Rank)).Set(0)
		// The metrics below are keyed only by validator (not by epoch or a value
		// label), so they are overwritten in place rather than reset, keeping the
		// series gap-free. They are seeded here and refreshed by the update methods.
		c.Rank.WithLabelValues(validator.AccountName, validator.Address).Set(float64(validator.WitnessInfo.Rank))
		c.Votes.WithLabelValues(validator.AccountName, validator.Address).Set(float64(validator.WitnessInfo.VoteCount))
		c.VotesPercentage.WithLabelValues(validator.AccountName, validator.Address).Set(0)
		c.VotesMarginToSR.WithLabelValues(validator.AccountName, validator.Address).Set(0)
		c.Balance.WithLabelValues(validator.AccountName, validator.Address).Set(float64(validator.Balance) / sunPerTRX)
		c.RewardBalance.WithLabelValues(validator.AccountName, validator.Address).Set(float64(validator.Allowance) / sunPerTRX)
		c.RewardsClaimable.WithLabelValues(validator.AccountName, validator.Address).Set(boolToFloat(validator.Allowance > 0))
		c.AccountActivated.WithLabelValues(validator.AccountName, validator.Address).Set(boolToFloat(validator.CreateTime > 0))
		c.FrozenBalance.WithLabelValues(validator.AccountName, validator.Address).Set(float64(validator.FrozenAmount()) / sunPerTRX)
		c.WitnessTotalProduced.WithLabelValues(validator.AccountName, validator.Address).Set(float64(validator.WitnessInfo.TotalProduced))
		c.WitnessTotalMissed.WithLabelValues(validator.AccountName, validator.Address).Set(float64(validator.WitnessInfo.TotalMissed))
		c.Brokerage.WithLabelValues(validator.AccountName, validator.Address).Set(0)
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

// UpdateRank records the validator's current ranking as a numeric gauge value.
// It is keyed only by validator so the series is stable across rank changes,
// and it is emitted for every monitored validator (including ones that dropped
// out of the active SR set) so a "lost SR status" alert can observe rank > N.
func (c *Collection) UpdateRank(validatorName, validatorAddress string, rank int) {
	c.Rank.WithLabelValues(validatorName, validatorAddress).Set(float64(rank))
}

// UpdateWitnessMetrics records the vote-related and lifetime block metrics that
// are derived from the network-wide witness list.
func (c *Collection) UpdateWitnessMetrics(validatorName, validatorAddress string, m WitnessMetrics) {
	c.Rank.WithLabelValues(validatorName, validatorAddress).Set(float64(m.Rank))
	c.Votes.WithLabelValues(validatorName, validatorAddress).Set(float64(m.Votes))
	c.VotesPercentage.WithLabelValues(validatorName, validatorAddress).Set(m.VotesPercentage)
	c.VotesMarginToSR.WithLabelValues(validatorName, validatorAddress).Set(float64(m.VotesMarginToSR))
	c.WitnessTotalProduced.WithLabelValues(validatorName, validatorAddress).Set(float64(m.TotalProduced))
	c.WitnessTotalMissed.WithLabelValues(validatorName, validatorAddress).Set(float64(m.TotalMissed))
}

// UpdateAccountMetrics records the balance and reward metrics sourced from the
// validator's on-chain account.
func (c *Collection) UpdateAccountMetrics(validatorName, validatorAddress string, account tron.Account) {
	c.Balance.WithLabelValues(validatorName, validatorAddress).Set(float64(account.Balance) / sunPerTRX)
	c.RewardBalance.WithLabelValues(validatorName, validatorAddress).Set(float64(account.Allowance) / sunPerTRX)
	c.RewardsClaimable.WithLabelValues(validatorName, validatorAddress).Set(boolToFloat(account.Allowance > 0))
	c.AccountActivated.WithLabelValues(validatorName, validatorAddress).Set(boolToFloat(account.CreateTime > 0))
	c.FrozenBalance.WithLabelValues(validatorName, validatorAddress).Set(float64(account.FrozenAmount()) / sunPerTRX)
}

func (c *Collection) UpdateBrokerage(validatorName, validatorAddress string, brokerage int) {
	c.Brokerage.WithLabelValues(validatorName, validatorAddress).Set(float64(brokerage))
}

func (c *Collection) UpdateActiveSRCount(count int) {
	c.ActiveSRCount.Set(float64(count))
}
func (c *Collection) UpdateNextMaintenanceTime(unixSeconds float64) {
	c.NextMaintenanceTime.Set(unixSeconds)
}
func (c *Collection) UpdateNodeUp(up bool) {
	c.NodeUp.Set(boolToFloat(up))
}
func (c *Collection) UpdateNodeHeadBlock(blockNumber float64) {
	c.NodeHeadBlock.Set(blockNumber)
}

func (c *Collection) UpdateEpoch(epoch float64) {
	c.Epoch.Set(epoch)
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
