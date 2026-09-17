package blockwatcher

import (
	"sort"
	"strings"

	"github.com/kilnfi/tron-validator-watcher/internal/tron"
)

// WitnessMetrics holds the per-validator values derived from the network-wide
// witness list (votes, ranking and lifetime block counters).
type WitnessMetrics struct {
	Rank            int
	Votes           int64
	VotesPercentage float64
	VotesMarginToSR int64
	TotalProduced   int64
	TotalMissed     int64
	Found           bool
}

// witnessBoard is a ranked snapshot of every witness on the network, used to
// derive per-validator metrics without extra RPC calls.
type witnessBoard struct {
	byAddress     map[string]WitnessMetrics
	activeSRCount int
}

// newWitnessBoard ranks the witnesses by vote count (highest first) and
// precomputes, for each of them, the vote share and the distance in votes to
// the active SR cutoff.
func newWitnessBoard(witnesses []tron.Witness) *witnessBoard {
	sorted := make([]tron.Witness, len(witnesses))
	copy(sorted, witnesses)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].VoteCount > sorted[j].VoteCount })

	var totalVotes int64
	activeCount := 0
	for _, w := range sorted {
		totalVotes += w.VoteCount
		if w.IsJobs {
			activeCount++
		}
	}

	board := &witnessBoard{
		byAddress:     make(map[string]WitnessMetrics, len(sorted)),
		activeSRCount: activeCount,
	}

	for i, w := range sorted {
		rank := i + 1

		var pct float64
		if totalVotes > 0 {
			pct = float64(w.VoteCount) / float64(totalVotes) * 100
		}

		board.byAddress[normalizeAddress(w.Address)] = WitnessMetrics{
			Rank:            rank,
			Votes:           w.VoteCount,
			VotesPercentage: pct,
			VotesMarginToSR: votesMarginToSR(sorted, activeCount, rank, w.VoteCount),
			TotalProduced:   w.TotalProduced,
			TotalMissed:     w.TotalMissed,
			Found:           true,
		}
	}

	return board
}

// votesMarginToSR returns how many votes separate a witness from the active SR
// cutoff. For a witness inside the set it is the cushion over the top challenger
// (positive); for one outside it is the deficit to the last seat (negative).
func votesMarginToSR(sorted []tron.Witness, activeCount, rank int, votes int64) int64 {
	if activeCount <= 0 {
		return 0
	}
	if rank <= activeCount {
		// Challenger sitting just outside the set (rank activeCount+1).
		if activeCount < len(sorted) {
			return votes - sorted[activeCount].VoteCount
		}
		return votes
	}
	// Last seat inside the set (rank activeCount).
	return votes - sorted[activeCount-1].VoteCount
}

// metricsFor returns the computed metrics for a witness given its hex address.
func (b *witnessBoard) metricsFor(hexAddress string) (WitnessMetrics, bool) {
	m, ok := b.byAddress[normalizeAddress(hexAddress)]
	return m, ok
}

func normalizeAddress(address string) string {
	return strings.ToLower(strings.TrimSpace(address))
}
