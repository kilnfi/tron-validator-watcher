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
	RoundDuration              = 21600 // Durée d'un round en secondes (6h)
	MaxSlotsPerRound           = 7200  // Nombre de slots par round
	BlockTime                  = 3     // Temps d'un bloc en secondes
	NumberOfValidators         = 27    // Nombre total de validateurs
	BlocksPerRound             = RoundDuration / BlockTime
	BlocksPerRoundPerValidator = BlocksPerRound / NumberOfValidators
	GenesisBlockTime           = 0 // Départ du bloc genesis
	MaintenanceSkipSlots       = 2 // Fenêtre de maintenance
	SingleRepeat               = 1 // Valeur de répétition unique
)
