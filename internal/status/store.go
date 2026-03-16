package status

import (
	"sync"
	"time"

	"github.com/kilnfi/tron-validator-watcher/internal/tron"
)

const maxRecentBlocks = 20

type RecentBlock struct {
	Number    int64  `json:"number"`
	Proposer  string `json:"proposer"`
	Address   string `json:"address"`
	IsOurs    bool   `json:"is_ours"`
	Missed    bool   `json:"missed"`
	Timestamp int64  `json:"timestamp"`
}

type ValidatorStatus struct {
	Name         string `json:"name"`
	Address      string `json:"address"`
	Rank         int    `json:"rank"`
	Proposed     int    `json:"proposed"`
	Missed       int    `json:"missed"`
	ConsecMissed int    `json:"consec_missed"`
	IsActive     bool   `json:"is_active"`
	VoteCount    int64  `json:"vote_count"`
	NextSlotInMs int64  `json:"next_slot_in_ms"` // -1 if unknown
}

type Status struct {
	Epoch             int               `json:"epoch"`
	RoundProgress     int               `json:"round_progress"`
	RoundDuration     int               `json:"round_duration"`
	NextRoundInMs     int64             `json:"next_round_in_ms"`
	IsOurSlot         bool              `json:"is_our_slot"`
	CurrentLeader     string            `json:"current_leader"`
	ProposedTotal     int               `json:"proposed_total"`
	MissedTotal       int               `json:"missed_total"`
	ConsecMissed      int               `json:"consec_missed"`
	RecentBlocks      []RecentBlock     `json:"recent_blocks"`
	NextOurSlotInMs   int64             `json:"next_our_slot_in_ms"` // -1 if unknown
	NextOurSlotLeader string            `json:"next_our_slot_leader"`
	Validators        []ValidatorStatus `json:"validators"`
	TotalSRs          int               `json:"total_srs"`
}

type validatorStat struct {
	proposed     int
	missed       int
	consecMissed int
}

type Store struct {
	mu             sync.RWMutex
	recentBlocks   []RecentBlock
	epoch          int
	validators     tron.AccountList
	totalSRs       int
	validatorStats map[string]*validatorStat // keyed by address
}

func NewStore() *Store {
	return &Store{
		recentBlocks:   make([]RecentBlock, 0, maxRecentBlocks),
		validatorStats: make(map[string]*validatorStat),
	}
}

func (s *Store) SetValidators(validators tron.AccountList) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.validators = validators
	// initialize stats for any new validator (preserve existing counts)
	for _, v := range validators {
		if _, ok := s.validatorStats[v.Address]; !ok {
			s.validatorStats[v.Address] = &validatorStat{}
		}
	}
}

func (s *Store) SetTotalSRs(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalSRs = n
}

func (s *Store) ResetEpoch(epoch int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.epoch = epoch
	for addr := range s.validatorStats {
		s.validatorStats[addr] = &validatorStat{}
	}
	s.recentBlocks = s.recentBlocks[:0]
}

func (s *Store) AddProposedBlock(block RecentBlock) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st, ok := s.validatorStats[block.Address]; ok {
		st.proposed++
		st.consecMissed = 0
	}
	s.prepend(block)
}

func (s *Store) AddMissedBlock(block RecentBlock) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st, ok := s.validatorStats[block.Address]; ok {
		st.missed++
		st.consecMissed++
	}
	s.prepend(block)
}

func (s *Store) AddOtherBlock(block RecentBlock) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prepend(block)
}

func (s *Store) prepend(block RecentBlock) {
	s.recentBlocks = append([]RecentBlock{block}, s.recentBlocks...)
	if len(s.recentBlocks) > maxRecentBlocks {
		s.recentBlocks = s.recentBlocks[:maxRecentBlocks]
	}
}

func (s *Store) Get() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now().UnixMilli()
	roundProgress := int((now / 1000) % tron.RoundDuration)
	nextRoundInMs := (int64(s.epoch)+1)*int64(tron.RoundDuration*1000) - now

	isOurSlot := false
	currentLeader := ""
	slot := now / (tron.BlockTime * 1000)
	witnessID := (slot % int64(tron.NumberOfValidators*tron.SingleRepeat)) / int64(tron.SingleRepeat)
	for _, v := range s.validators {
		if v.WitnessInfo != nil && v.WitnessInfo.Rank == int(witnessID+1) {
			isOurSlot = true
			currentLeader = v.AccountName
			break
		}
	}

	// Find the most recent production timestamp per validator address
	lastProducedAt := make(map[string]int64)
	for _, b := range s.recentBlocks {
		if b.IsOurs && !b.Missed {
			if _, seen := lastProducedAt[b.Address]; !seen {
				lastProducedAt[b.Address] = b.Timestamp
			}
		}
	}

	// Compute next slot per validator
	nextSlotPerValidator := make(map[string]int64) // address → inMs
	nextOurSlotInMs := int64(-1)
	nextOurSlotLeader := ""
	if s.totalSRs > 0 {
		n := int64(s.totalSRs)
		rotation := n * int64(tron.BlockTime) * 1000
		for _, v := range s.validators {
			if v.WitnessInfo == nil {
				nextSlotPerValidator[v.Address] = -1
				continue
			}
			var inMs int64
			if lastTs, ok := lastProducedAt[v.Address]; ok {
				inMs = lastTs + rotation - now
				for inMs <= 0 {
					inMs += rotation
				}
			} else {
				rank := int64(v.WitnessInfo.Rank)
				currentSlot := now / (int64(tron.BlockTime) * 1000)
				slotsUntil := ((rank - 1) - currentSlot%n + n) % n
				if slotsUntil == 0 {
					slotsUntil = n
				}
				inMs = slotsUntil*int64(tron.BlockTime)*1000 - (now % (int64(tron.BlockTime) * 1000))
			}
			nextSlotPerValidator[v.Address] = inMs
			if nextOurSlotInMs == -1 || inMs < nextOurSlotInMs {
				nextOurSlotInMs = inMs
				nextOurSlotLeader = v.AccountName
			}
		}
	}

	// build per-validator status + aggregates
	proposedTotal := 0
	missedTotal := 0
	consecMissedMax := 0
	validators := make([]ValidatorStatus, 0, len(s.validators))
	for _, v := range s.validators {
		st := s.validatorStats[v.Address]
		if st == nil {
			st = &validatorStat{}
		}
		rank := 0
		isActive := false
		voteCount := int64(0)
		if v.WitnessInfo != nil {
			rank = v.WitnessInfo.Rank
			isActive = v.WitnessInfo.IsJobs
			voteCount = v.WitnessInfo.VoteCount
		}
		nextSlot, ok := nextSlotPerValidator[v.Address]
		if !ok {
			nextSlot = -1
		}
		validators = append(validators, ValidatorStatus{
			Name:         v.AccountName,
			Address:      v.Address,
			Rank:         rank,
			Proposed:     st.proposed,
			Missed:       st.missed,
			ConsecMissed: st.consecMissed,
			IsActive:     isActive,
			VoteCount:    voteCount,
			NextSlotInMs: nextSlot,
		})
		proposedTotal += st.proposed
		missedTotal += st.missed
		if st.consecMissed > consecMissedMax {
			consecMissedMax = st.consecMissed
		}
	}

	blocks := make([]RecentBlock, len(s.recentBlocks))
	copy(blocks, s.recentBlocks)

	return Status{
		Epoch:             s.epoch,
		RoundProgress:     roundProgress,
		RoundDuration:     tron.RoundDuration,
		NextRoundInMs:     nextRoundInMs,
		IsOurSlot:         isOurSlot,
		CurrentLeader:     currentLeader,
		ProposedTotal:     proposedTotal,
		MissedTotal:       missedTotal,
		ConsecMissed:      consecMissedMax,
		RecentBlocks:      blocks,
		NextOurSlotInMs:   nextOurSlotInMs,
		NextOurSlotLeader: nextOurSlotLeader,
		Validators:        validators,
		TotalSRs:          s.totalSRs,
	}
}
