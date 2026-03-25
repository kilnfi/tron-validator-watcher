package tron

// API endpoints
const (
	APIListWitnessesEndpoint  = "/wallet/listwitnesses"
	APIGetAccountInfoEndpoint = "/wallet/getaccount"
	APIGetLatestBlockEndpoint = "/wallet/getnowblock"
	APIGetBlockByNumEndpoint  = "/wallet/getblockbynum"
)

// Chain parameters
const (
	RoundDuration        = 21600 // Round duration in seconds (6h)
	MaxSlotsPerRound     = 7200  // Number of slots per round
	BlockTime            = 3     // Block time in seconds
	BlocksPerRound       = RoundDuration / BlockTime
	GenesisBlockTime     = 0 // Genesis block timestamp
	MaintenanceSkipSlots = 2 // Maintenance window slots skipped at start of round
	SingleRepeat         = 1 // Each SR produces this many consecutive blocks
)
