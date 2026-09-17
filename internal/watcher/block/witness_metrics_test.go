package blockwatcher

import (
	"testing"

	"github.com/kilnfi/tron-validator-watcher/internal/tron"
	"github.com/stretchr/testify/require"
)

func TestNewWitnessBoard(t *testing.T) {
	t.Parallel()

	// Two active SRs (A, B) and two challengers (C, D), given out of order to
	// make sure the board ranks them by vote count.
	witnesses := []tron.Witness{
		{Address: "c", VoteCount: 60, TotalProduced: 6, TotalMissed: 3},
		{Address: "a", VoteCount: 100, IsJobs: true, TotalProduced: 10, TotalMissed: 1},
		{Address: "d", VoteCount: 40},
		{Address: "b", VoteCount: 80, IsJobs: true, TotalProduced: 8, TotalMissed: 2},
	}

	board := newWitnessBoard(witnesses)

	require.Equal(t, 2, board.activeSRCount)

	// Top active SR: rank 1, cushion over the first challenger (C at 60) = 40.
	a, ok := board.metricsFor("A") // also checks address normalization
	require.True(t, ok)
	require.Equal(t, 1, a.Rank)
	require.Equal(t, int64(100), a.Votes)
	require.InDelta(t, 100.0/280.0*100, a.VotesPercentage, 1e-9)
	require.Equal(t, int64(40), a.VotesMarginToSR)
	require.Equal(t, int64(10), a.TotalProduced)
	require.Equal(t, int64(1), a.TotalMissed)

	// Last active SR: rank 2, cushion over the challenger just below the cutoff.
	b, ok := board.metricsFor("b")
	require.True(t, ok)
	require.Equal(t, 2, b.Rank)
	require.Equal(t, int64(20), b.VotesMarginToSR)

	// First challenger outside the set: rank 3, negative margin (deficit to the
	// last seat, B at 80).
	c, ok := board.metricsFor("c")
	require.True(t, ok)
	require.Equal(t, 3, c.Rank)
	require.Equal(t, int64(-20), c.VotesMarginToSR)

	// Unknown address is not found.
	_, ok = board.metricsFor("unknown")
	require.False(t, ok)
}

func TestFrozenAmount(t *testing.T) {
	t.Parallel()

	account := tron.Account{
		FrozenV2: []tron.FrozenV2{
			{Amount: 1_000_000},
			{Type: "ENERGY", Amount: 2_000_000},
			{Type: "TRON_POWER"},
		},
	}
	require.Equal(t, int64(3_000_000), account.FrozenAmount())
}
