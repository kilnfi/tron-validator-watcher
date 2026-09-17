package blockwatcher

import (
	"context"
	"fmt"
	"time"

	"github.com/kilnfi/tron-validator-watcher/internal/status"
	"github.com/kilnfi/tron-validator-watcher/internal/tron"

	"github.com/sirupsen/logrus"
)

type BlockWatcher struct {
	startBlock             *tron.Block
	tronClient             *tron.Client
	validators             tron.AccountList
	filteredValidators     tron.AccountList
	logger                 *logrus.Logger
	refreshInterval        int
	roundProgress          int
	epoch                  int
	totalSRs               int
	metrics                *Collection
	store                  *status.Store
	lastProcessedTimestamp int64
}

func NewBlockWatcher(options ...WatcherOptionFunc) (*BlockWatcher, error) {
	bw := &BlockWatcher{}
	for _, fn := range options {
		if fn == nil {
			continue
		}
		if err := fn(bw); err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	blockInfo, err := bw.tronClient.Network.GetBlockByNumber(ctx, bw.startBlock.BlockHeader.RawData.Number)
	if err != nil {
		return nil, fmt.Errorf("BlockWatcher: failed to retrieve block with number %d: %w", bw.startBlock.BlockHeader.RawData.Number, err)
	}

	blockTime := time.UnixMilli(blockInfo.BlockHeader.RawData.Timestamp)
	currentTime := time.Now().UTC()
	diff := blockTime.Sub(currentTime)

	if diff.Abs() > (6*time.Hour + time.Minute) {
		bw.logger.WithFields(logrus.Fields{
			"diff":    diff,
			"service": "block-watcher",
		}).Warn("BlockWatcher: block time is too old. Using latest block...")

		block, err := bw.tronClient.Network.GetLatestBlock(ctx)
		if err != nil {
			return nil, fmt.Errorf("BlockWatcher: failed to get latest block: %w", err)
		}
		bw.startBlock = block
	}

	bw.epoch = tron.GetEpochID(bw.startBlock)
	bw.lastProcessedTimestamp = bw.startBlock.BlockHeader.RawData.Timestamp

	return bw, nil
}

func (bw *BlockWatcher) Start(ctx context.Context) error {
	// filter out only block producers validators
	bw.filteredValidators = bw.getBlockProducers()
	bw.metrics.InitMetrics(bw.epoch, bw.filteredValidators)
	for _, validator := range bw.filteredValidators {
		bw.logger.Infof("🥇 Validator %s is ranked #%d", validator.AccountName, validator.WitnessInfo.Rank)
		bw.metrics.UpdateBlockProducerInfo(validator.AccountName, validator.Address, validator.WitnessInfo.Rank)
	}
	// Populate every metric up front so the exporter serves data immediately at
	// startup instead of waiting for the first round change. The account data was
	// already fetched into bw.validators before the watcher was created.
	bw.refreshWitnessMetrics()
	bw.refreshAccountMetrics(ctx)

	if witnesses, err := bw.tronClient.Account.ListWitnesses(); err == nil {
		activeCount := 0
		for _, w := range witnesses.Witnesses {
			if w.IsJobs {
				activeCount++
			}
		}
		bw.totalSRs = activeCount
		bw.logger.Infof("🌐 Active SRs: %d", activeCount)
		if bw.store != nil {
			bw.store.SetTotalSRs(activeCount)
		}
	}

	if bw.store != nil {
		bw.store.SetValidators(bw.filteredValidators)
		bw.store.ResetEpoch(bw.epoch)
	}

	bw.roundProgress = int((bw.startBlock.BlockHeader.RawData.Timestamp / 1000) % tron.RoundDuration)
	ticker := time.NewTicker(time.Duration(bw.refreshInterval) * time.Second)

	for {
		select {
		case <-ctx.Done():
			bw.logger.Info("stopping watcher")
			return fmt.Errorf("context done in watcher: %w", ctx.Err())
		case <-ticker.C:
			if err := bw.start(ctx); err != nil {
				return fmt.Errorf("BlockWatcher: failed to start: %w", err)
			}
			bw.logger.Infof("😴 Sleeping %ds before next iteration...", bw.refreshInterval)
		}
	}
}

func (bw *BlockWatcher) start(ctx context.Context) error {
	block, err := bw.tronClient.Network.GetLatestBlock(ctx)
	if err != nil {
		bw.metrics.UpdateNodeUp(false)
		return fmt.Errorf("BlockWatcher: failed to fetch latest block: %w", err)
	}
	bw.metrics.UpdateNodeUp(true)
	bw.metrics.UpdateNodeHeadBlock(float64(block.BlockHeader.RawData.Number))

	// Keep the vote/rank metrics fresh on every tick (they change continuously),
	// unlike the account metrics which are only refreshed on round changes.
	bw.refreshWitnessMetrics()

	roundChanged := false
	progress := bw.roundProgress

	start := bw.startBlock.BlockHeader.RawData.Number
	end := block.BlockHeader.RawData.Number
	for bid := start; bid < end; bid++ {
		bw.logger.WithFields(
			logrus.Fields{
				"block":   bid,
				"epoch":   bw.epoch,
				"service": "block-watcher",
			},
		).Infof("📦 Processing block #%d (progress=%d/%d slots)", bid, progress, tron.RoundDuration)

		currentBlock, err := bw.tronClient.Network.GetBlockByNumber(ctx, bid)
		if err != nil {
			return fmt.Errorf("BlockWatcher: failed to retrieve block with number %d: %w", bid, err)
		}

		// Check if a new round has started
		roundChanged, progress = bw.isNewRound(currentBlock.BlockHeader.RawData.Timestamp)
		if roundChanged {
			if err := bw.handleRoundChanged(ctx, currentBlock); err != nil {
				return fmt.Errorf("BlockWatcher: failed to handle round change: %w", err)
			}
		}

		// Detect missed slots: if the gap between the last processed block and the current block
		// is greater than BlockTime, some slots were skipped (missed blocks).
		currentTimestamp := currentBlock.BlockHeader.RawData.Timestamp
		for ts := bw.lastProcessedTimestamp + tron.BlockTime*1000; ts < currentTimestamp; ts += tron.BlockTime * 1000 {
			isLeaderForSlot, slotAccount := bw.isLeader(ts)
			if isLeaderForSlot {
				bw.logger.WithFields(logrus.Fields{
					"validator_name": slotAccount.AccountName,
					"expected_slot":  ts / 1000,
					"service":        "block-watcher",
				}).Infof("❌ Our Validator %s missed a block (skipped slot)", slotAccount.AccountName)
				bw.metrics.UpdateMissedBlock(bw.epoch, slotAccount.AccountName, slotAccount.Address)
				bw.metrics.UpdateConsecutiveMissedBlock(bw.epoch, slotAccount.AccountName, slotAccount.Address, false)
				if bw.store != nil {
					bw.store.AddMissedBlock(status.RecentBlock{
						Number:    currentBlock.BlockHeader.RawData.Number,
						Proposer:  slotAccount.AccountName,
						Address:   slotAccount.Address,
						IsOurs:    true,
						Missed:    true,
						Timestamp: ts,
					})
				}
			}
		}
		bw.lastProcessedTimestamp = currentTimestamp

		proposerAddress, err := tron.ConvertAddressToBase58(currentBlock.BlockHeader.RawData.WitnessAddress)
		if err != nil {
			return fmt.Errorf("BlockWatcher: failed to convert address to base58: %w", err)
		}
		proposerInfo, err := bw.tronClient.Account.GetAccount(proposerAddress)
		if err != nil {
			return fmt.Errorf("BlockWatcher: failed to get proposer account: %w", err)
		}

		// Sometimes it can happen that the proposer does not have a name.
		// In this case, we use the address as the name.
		if proposerInfo.AccountName == "" {
			proposerInfo.AccountName = proposerInfo.Address
		}

		// Check if the validator is the leader
		isLeader, account := bw.isLeader(currentBlock.BlockHeader.RawData.Timestamp)
		if isLeader {
			bw.handleSlotLeader(currentBlock, proposerAddress, account)
		} else {
			// isLeader() may miss our validator if NumberOfValidators doesn't match
			// the real network size. As a fallback, always check if the proposer
			// is one of our validators by address comparison.
			ours := false
			for _, v := range bw.filteredValidators {
				if proposerAddress == v.Address {
					ours = true
					bw.logger.WithFields(logrus.Fields{
						"validator_name": v.AccountName,
						"block":          currentBlock.BlockHeader.RawData.Number,
						"block_time":     (currentBlock.BlockHeader.RawData.Timestamp / 1000),
						"service":        "block-watcher",
					}).Infof("✅ Our Validator %s proposed a block (fallback detection)", v.AccountName)
					bw.metrics.UpdateProposedBlock(bw.epoch, v.AccountName, v.Address)
					bw.metrics.UpdateConsecutiveMissedBlock(bw.epoch, v.AccountName, v.Address, true)
					if bw.store != nil {
						bw.store.AddProposedBlock(status.RecentBlock{
							Number:    currentBlock.BlockHeader.RawData.Number,
							Proposer:  v.AccountName,
							Address:   v.Address,
							IsOurs:    true,
							Missed:    false,
							Timestamp: currentBlock.BlockHeader.RawData.Timestamp,
						})
					}
					break
				}
			}
			if !ours {
				bw.logger.WithFields(logrus.Fields{
					"validator_name": proposerInfo.AccountName,
					"block":          currentBlock.BlockHeader.RawData.Number,
					"block_time":     (currentBlock.BlockHeader.RawData.Timestamp / 1000),
					"block_slot":     progress,
					"service":        "block-watcher",
				}).Infof("🏆 Validator %s proposed a block", proposerInfo.AccountName)
				if bw.store != nil {
					bw.store.AddOtherBlock(status.RecentBlock{
						Number:    currentBlock.BlockHeader.RawData.Number,
						Proposer:  proposerInfo.AccountName,
						Address:   proposerAddress,
						IsOurs:    false,
						Missed:    false,
						Timestamp: currentBlock.BlockHeader.RawData.Timestamp,
					})
				}
			}
		}
	}

	if !roundChanged {
		bw.startBlock = block
		bw.roundProgress = progress
	}

	bw.metrics.UpdateLatestBlockProcessed(float64(block.BlockHeader.RawData.Number))
	bw.metrics.UpdateRoundProgress(float64(progress))

	return nil
}

func (bw *BlockWatcher) getBlockProducers() tron.AccountList {
	blockProducers := bw.validators.GetBlockProducerValidators()

	// Create a set for fast lookup
	producerSet := make(map[string]bool, len(blockProducers))
	for _, producer := range blockProducers {
		producerSet[producer.Address] = true
	}

	// Filter only block producers
	var filtered []tron.Account
	for _, validator := range bw.validators {
		if producerSet[validator.Address] {
			filtered = append(filtered, validator)
		} else {
			bw.logger.Warnf("🔕 Validator %s ranked %d is not a block producer, ignoring...", validator.AccountName, validator.WitnessInfo.Rank)
		}
	}

	return filtered
}

func (bw *BlockWatcher) isLeader(timeSlot int64) (bool, tron.Account) {
	var witnessID int64
	slot := (timeSlot - tron.GenesisBlockTime) / (tron.BlockTime * 1000)

	// During the maintenance period, no blocks are produced for 6 seconds (2 slots).
	// To prevent misalignment with the round-robin rotation, we must account for this period
	// and adjust the witness ID accordingly. Otherwise, we might select the wrong validator.
	n := int64(bw.totalSRs)
	if n == 0 {
		return false, tron.Account{}
	}
	if (timeSlot/1000)%tron.RoundDuration == 0 {
		bw.logger.Debug("isLeader: First slot of the round detected, considering maintenance period")
		witnessID = ((slot - tron.MaintenanceSkipSlots) % (n * tron.SingleRepeat)) / tron.SingleRepeat
	} else {
		witnessID = (slot % (n * tron.SingleRepeat)) / tron.SingleRepeat
	}

	for _, account := range bw.filteredValidators {
		if account.WitnessInfo.Rank == int(witnessID+1) {
			return true, account
		}
	}
	return false, tron.Account{}
}

func (bw *BlockWatcher) handleSlotLeader(block *tron.Block, proposer string, account tron.Account) {
	bw.logger.WithFields(logrus.Fields{
		"validator_name": account.AccountName,
		"block":          block.BlockHeader.RawData.Number,
		"block_time":     (block.BlockHeader.RawData.Timestamp / 1000),
		"service":        "block-watcher",
	}).Infof("👑 Our validator %s is leader", account.AccountName)

	if proposer == account.Address {
		bw.logger.WithFields(logrus.Fields{
			"validator_name": account.AccountName,
			"block":          block.BlockHeader.RawData.Number,
			"block_time":     (block.BlockHeader.RawData.Timestamp / 1000),
			"service":        "block-watcher",
		}).Infof("✅ Our Validator %s proposed a block", account.AccountName)

		bw.metrics.UpdateProposedBlock(bw.epoch, account.AccountName, account.Address)
		bw.metrics.UpdateConsecutiveMissedBlock(bw.epoch, account.AccountName, account.Address, true)
		if bw.store != nil {
			bw.store.AddProposedBlock(status.RecentBlock{
				Number:    block.BlockHeader.RawData.Number,
				Proposer:  account.AccountName,
				Address:   account.Address,
				IsOurs:    true,
				Missed:    false,
				Timestamp: block.BlockHeader.RawData.Timestamp,
			})
		}
	} else {
		bw.logger.WithFields(logrus.Fields{
			"validator_name": account.AccountName,
			"block":          block.BlockHeader.RawData.Number,
			"block_time":     (block.BlockHeader.RawData.Timestamp / 1000),
			"service":        "block-watcher",
		}).Infof("❌ Our Validator %s missed a block", account.AccountName)

		bw.metrics.UpdateMissedBlock(bw.epoch, account.AccountName, account.Address)
		bw.metrics.UpdateConsecutiveMissedBlock(bw.epoch, account.AccountName, account.Address, false)
		if bw.store != nil {
			bw.store.AddMissedBlock(status.RecentBlock{
				Number:    block.BlockHeader.RawData.Number,
				Proposer:  account.AccountName,
				Address:   account.Address,
				IsOurs:    true,
				Missed:    true,
				Timestamp: block.BlockHeader.RawData.Timestamp,
			})
		}
	}
}

func (bw *BlockWatcher) handleRoundChanged(ctx context.Context, block *tron.Block) error {
	bw.logger.WithFields(logrus.Fields{
		"block":      block.BlockHeader.RawData.Number,
		"block_time": (block.BlockHeader.RawData.Timestamp / 1000),
		"service":    "block-watcher",
	}).Info("🎉 A new round has started. Refreshing validator data...")

	// Refresh account data
	if err := bw.refresh(); err != nil {
		return fmt.Errorf("handleRoundChanged: failed to refresh account data: %w", err)
	}

	// InitMetrics resets the per-epoch metrics (including block_producer_info),
	// so it must run before we set the current values, otherwise the values set
	// below would be wiped out immediately.
	bw.epoch = tron.GetEpochID(block)
	bw.metrics.InitMetrics(bw.epoch, bw.filteredValidators)

	for _, validator := range bw.filteredValidators {
		bw.logger.Infof("🥇 Validator %s is ranked #%d", validator.AccountName, validator.WitnessInfo.Rank)
		bw.metrics.UpdateBlockProducerInfo(validator.AccountName, validator.Address, validator.WitnessInfo.Rank)
	}
	// Refresh vote/rank metrics (all validators) and the slower account/brokerage
	// metrics now that we have fresh account data for this round.
	bw.refreshWitnessMetrics()
	bw.refreshAccountMetrics(ctx)

	if witnesses, err := bw.tronClient.Account.ListWitnesses(); err == nil {
		activeCount := 0
		for _, w := range witnesses.Witnesses {
			if w.IsJobs {
				activeCount++
			}
		}
		bw.totalSRs = activeCount
		bw.logger.Infof("🌐 Active SRs: %d", activeCount)
		if bw.store != nil {
			bw.store.SetTotalSRs(activeCount)
		}
	}

	if bw.store != nil {
		bw.store.SetValidators(bw.filteredValidators)
		bw.store.ResetEpoch(bw.epoch)
	}

	time.Sleep(3 * time.Second)
	nextBlock, err := bw.tronClient.Network.GetBlockByNumber(ctx, block.BlockHeader.RawData.Number+1)
	if err != nil {
		return fmt.Errorf("BlockWatcher: failed to get next block: %v", err)
	}
	bw.startBlock = nextBlock
	bw.lastProcessedTimestamp = block.BlockHeader.RawData.Timestamp
	bw.roundProgress = 0

	return nil
}

func (bw *BlockWatcher) isNewRound(timeSlot int64) (bool, int) {
	progress := (timeSlot / 1000) % tron.RoundDuration
	if progress == 0 {
		return true, int(progress)
	}
	return false, int(progress)
}

func (bw *BlockWatcher) refresh() error {
	accounts := []tron.Account{}

	for _, account := range bw.validators {
		account, err := bw.tronClient.Account.GetAccount(account.Address)

		if err != nil {
			return fmt.Errorf("refresh: failed to get account info: %w", err)
		}
		accounts = append(accounts, *account)
	}

	bw.validators = accounts
	bw.filteredValidators = bw.getBlockProducers()
	return nil
}

// refreshWitnessMetrics recomputes the vote/rank metrics for every monitored
// validator from the network-wide witness list. It is best-effort: a failure is
// logged and never interrupts block processing.
func (bw *BlockWatcher) refreshWitnessMetrics() {
	witnesses, err := bw.tronClient.Account.ListWitnesses()
	if err != nil {
		bw.logger.Warnf("failed to refresh witness metrics: %v", err)
		return
	}

	// A successful witness fetch means the node answered, so report it as up
	// (relevant at startup, before the first block-processing tick runs).
	bw.metrics.UpdateNodeUp(true)

	board := newWitnessBoard(witnesses.Witnesses)
	bw.metrics.UpdateActiveSRCount(board.activeSRCount)

	margins := make(map[string]int64, len(bw.validators))
	for _, validator := range bw.validators {
		metrics, ok := board.metricsFor(tron.ConvertAddressToHex(validator.Address))
		if !ok {
			continue
		}
		bw.metrics.UpdateWitnessMetrics(validator.AccountName, validator.Address, metrics)
		margins[validator.Address] = metrics.VotesMarginToSR
	}
	if bw.store != nil {
		bw.store.SetVotesMargin(margins)
	}
}

// refreshAccountMetrics records the slower-moving account metrics (balance,
// rewards, stake, brokerage) and the next maintenance time. It relies on the
// account data already refreshed into bw.validators and is best-effort.
func (bw *BlockWatcher) refreshAccountMetrics(ctx context.Context) {
	for _, validator := range bw.validators {
		bw.metrics.UpdateAccountMetrics(validator.AccountName, validator.Address, validator)

		brokerage, err := bw.tronClient.Account.GetBrokerage(validator.Address)
		if err != nil {
			bw.logger.Warnf("failed to fetch brokerage for %s: %v", validator.AccountName, err)
		} else {
			bw.metrics.UpdateBrokerage(validator.AccountName, validator.Address, brokerage)
		}
	}

	nextMaintenance, err := bw.tronClient.Network.GetNextMaintenanceTime(ctx)
	if err != nil {
		bw.logger.Warnf("failed to fetch next maintenance time: %v", err)
		return
	}
	// The node reports milliseconds; expose seconds to match Prometheus conventions.
	bw.metrics.UpdateNextMaintenanceTime(float64(nextMaintenance) / 1000)
}
